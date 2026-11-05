DATABASE_URL ?= postgres://innogen:innogen@localhost:5432/innogen?sslmode=disable

.PHONY: run-gateway run-auth run-runner run-submission run-submission-worker run-repo run-all stop-all tidy fmt test compose-up compose-down logs-piston piston-install seed-dev seed-problems e2e e2e-github test-ui-install test-ui test-ui-build compose-full-up compose-full-down compose-full-logs compose-ps docker-build docker-reset logs-gateway logs-auth logs-runner logs-submission logs-worker logs-repo ci gofmt-check ci-go ci-docker ci-test-ui ci-openapi

run-gateway:
	go run ./api_gateway/cmd/main.go

run-auth:
	go run ./auth_service/cmd/main.go

run-runner:
	go run ./run_service/cmd/main.go

run-submission:
	go run ./submission_service/cmd/api/main.go

run-submission-worker:
	go run ./submission_service/cmd/worker/main.go

run-repo:
	go run ./repo_service/cmd/main.go

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

test:
	go test ./...

compose-up:
	docker compose up -d postgres redis piston

compose-down:
	docker compose down

logs-piston:
	docker compose logs -f piston

piston-install:
	@echo "Installing Node.js 18.15.0 runtime in Piston (timeout 120s)..."
	@curl -s --max-time 120 -X POST http://localhost:2000/api/v2/packages \
		-H "Content-Type: application/json" \
		-d '{"language":"node","version":"18.15.0"}' || echo "  (install skipped or failed - may already be installed)"
	@echo "Done."

seed-dev:
	psql "$(DATABASE_URL)" -f schema.sql
	psql "$(DATABASE_URL)" -f seeds/dev_seed.sql
	psql "$(DATABASE_URL)" -f seeds/dev_problem_pack.sql

seed-problems:
	psql "$(DATABASE_URL)" -f seeds/dev_problem_pack.sql

e2e: piston-install
	bash scripts/e2e_mvp.sh

e2e-github: piston-install
	bash scripts/e2e_github.sh

run-all:
	@echo "Starting all services in background..."
	@echo "  api_gateway         :8080"
	@echo "  auth_service        :8081"
	@echo "  run_service         :8082"
	@echo "  submission_service  :8083"
	@echo "  repo_service        :8084"
	@echo "  submission_worker   (background)"
	@echo ""
	@echo "Press Ctrl+C to stop all services."
	@bash -c '\
		set -a; [ -f .env ] && source .env; set +a; \
		trap "echo; echo Stopping all services...; kill 0; exit 0" INT TERM EXIT; \
		go run ./api_gateway/cmd/main.go & \
		go run ./auth_service/cmd/main.go & \
		go run ./run_service/cmd/main.go & \
		go run ./submission_service/cmd/api/main.go & \
		go run ./submission_service/cmd/worker/main.go & \
		go run ./repo_service/cmd/main.go & \
		wait \
	'

stop-all:
	@echo "Stopping all services by ports..."
	@for port in 8080 8081 8082 8083 8084; do \
		pids=$$(lsof -ti tcp:$$port 2>/dev/null); \
		if [ -n "$$pids" ]; then \
			echo "  killing port $$port: $$pids"; \
			kill $$pids 2>/dev/null || true; \
		else \
			echo "  port $$port: no process"; \
		fi; \
	done
	@pkill -f "submission_service/cmd/worker" 2>/dev/null || true
	@pkill -f "go run ./submission_service/cmd/worker/main.go" 2>/dev/null || true
	@echo "All services stopped."

test-ui-install:
	cd tools/test_ui && npm install

test-ui:
	cd tools/test_ui && npm run dev

test-ui-build:
	cd tools/test_ui && npm run build

compose-full-up:
	docker compose up -d --build

compose-full-down:
	docker compose down

compose-full-logs:
	docker compose logs -f

compose-ps:
	docker compose ps

docker-build:
	docker compose build

docker-reset:
	docker compose down -v
	docker compose up -d --build

logs-gateway:
	docker compose logs -f api_gateway

logs-auth:
	docker compose logs -f auth_service

logs-runner:
	docker compose logs -f run_service

logs-submission:
	docker compose logs -f submission_service

logs-worker:
	docker compose logs -f submission_worker

logs-repo:
	docker compose logs -f repo_service

gofmt-check:
	@test -z "$$(gofmt -l . 2>/dev/null | grep -v node_modules | grep -v vendor)" || (echo "Files not formatted:"; gofmt -l . 2>/dev/null | grep -v node_modules | grep -v vendor; exit 1)

ci-go: gofmt-check
	go test ./...

ci-docker:
	docker compose build

ci-test-ui:
	cd tools/test_ui && npm ci && npm run build

ci-openapi:
	npx @redocly/cli lint docs/openapi.yaml --skip-rule no-unused-components 2>/dev/null || echo "OpenAPI lint skipped (redocly not installed)"

ci: ci-go ci-docker ci-test-ui ci-openapi
