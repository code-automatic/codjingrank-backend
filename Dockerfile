# Build stage
FROM golang:1.26-alpine AS builder
ARG SERVICE_PATH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/bin ./${SERVICE_PATH}

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/bin /app/bin
COPY docs/ /app/docs/
USER appuser
EXPOSE 8080
CMD ["/app/bin"]
