#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_EMAIL="${TEST_EMAIL:-admin@example.com}"
TEST_PASSWORD="${TEST_PASSWORD:-password}"
REQUIRE_GITHUB_COMMIT="${REQUIRE_GITHUB_COMMIT:-false}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
fail() { echo -e "${RED}FAIL${NC} $1"; exit 1; }
pass() { echo -e "${GREEN}PASS${NC} $1"; }
info() { echo -e "${YELLOW}INFO${NC} $1"; }

echo "=== GitHub Commit E2E Test ==="

# 1. Login
info "1. Logging in..."
LOGIN=$(curl -s -X POST "$API_BASE_URL/auth/login" -H "Content-Type: application/json" -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
TOKEN=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then fail "Login failed"; fi
pass "Login OK"
AUTH="Authorization: Bearer $TOKEN"

# 2. Submit accepted solution
info "2. Submitting solution..."
SOLUTION="const fs = require('fs'); const input = fs.readFileSync(0, 'utf8').trim().split(' ').map(Number); console.log(input[0] + input[1]);"
SUBMIT=$(curl -s -X POST "$API_BASE_URL/submit" -H "Content-Type: application/json" -H "$AUTH" -d "{\"problemId\":1,\"languageId\":1,\"code\":\"$SOLUTION\"}")
SUB_ID=$(echo "$SUBMIT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['submission']['id'])" 2>/dev/null || echo "")
if [ -z "$SUB_ID" ]; then fail "Submit failed"; fi
pass "Submission: $SUB_ID"

# 3. Poll for result
info "3. Polling submission..."
for i in $(seq 1 20); do
    sleep 2
    RESULT=$(curl -s "$API_BASE_URL/submissions/$SUB_ID" -H "$AUTH")
    STATUS=$(echo "$RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['submission']['status'])" 2>/dev/null || echo "Pending")
    echo "  [${i}] $STATUS"
    case "$STATUS" in
        Accepted|"Wrong Answer"|"Compilation Error"|"Runtime Error"|"Time Limit Exceeded"|"Internal Error")
            break ;;
    esac
done

# 4. Verify result
info "4. Checking result..."
if [ "$STATUS" != "Accepted" ]; then
    echo "Final status: $STATUS"
    fail "Submission not Accepted"
fi
pass "Verdict: Accepted"

# 5. Check commit metadata
info "5. Checking commit metadata..."
FINAL=$(curl -s "$API_BASE_URL/submissions/$SUB_ID" -H "$AUTH")
REPO_PATH=$(echo "$FINAL" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['submission'].get('repoPath',''))" 2>/dev/null || echo "")
COMMIT_SHA=$(echo "$FINAL" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['submission'].get('commitSha',''))" 2>/dev/null || echo "")
COMMIT_URL=$(echo "$FINAL" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['submission'].get('commitUrl',''))" 2>/dev/null || echo "")

if [ -n "$REPO_PATH" ]; then pass "repoPath: $REPO_PATH"; else pass "repoPath: (empty)"; fi
if [ -n "$COMMIT_SHA" ]; then pass "commitSha: $COMMIT_SHA"; else pass "commitSha: (empty)"; fi
if [ -n "$COMMIT_URL" ]; then pass "commitUrl: $COMMIT_URL"; else info "commitUrl: (empty)"; fi

if [ "$REQUIRE_GITHUB_COMMIT" = "true" ]; then
    if [ -z "$COMMIT_URL" ]; then fail "REQUIRE_GITHUB_COMMIT=true but no commitUrl"; fi
fi

echo ""
echo -e "${GREEN}=== GitHub E2E Complete ===${NC}"
