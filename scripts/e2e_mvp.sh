#!/usr/bin/env bash
set -euo pipefail

# RinnoGen MVP End-to-End Validation Script
# Requires: curl, bash, and jq (for JSON parsing)
# All services must be running before executing this script.

GATEWAY="http://localhost:8080"
EMAIL="admin@example.com"
PASSWORD="password"
TIMEOUT_SECONDS=30
POLL_INTERVAL=2

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

fail() { echo -e "${RED}FAIL${NC} $1"; exit 1; }
pass() { echo -e "${GREEN}PASS${NC} $1"; }
info() { echo -e "${YELLOW}INFO${NC} $1"; }

echo "=== RinnoGen MVP E2E Test ==="

# 1. Health check
echo ""
info "1. Checking gateway health..."
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY/health")
if [ "$HEALTH" != "200" ]; then fail "Gateway health returned $HEALTH"; fi
pass "Gateway health OK"

# 2. Login
echo ""
info "2. Logging in as admin..."
LOGIN_RESP=$(curl -s -X POST "$GATEWAY/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
ACCESS_TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.accessToken // empty')
if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
    echo "Login response: $LOGIN_RESP"
    fail "Failed to extract accessToken"
fi
pass "Login successful (token: ${ACCESS_TOKEN:0:20}...)"

AUTH_HEADER="Authorization: Bearer $ACCESS_TOKEN"

# 3. List subjects
echo ""
info "3. Listing subjects..."
SUBJECTS=$(curl -s "$GATEWAY/subjects")
SUBJECT_COUNT=$(echo "$SUBJECTS" | jq -r '.data.subjects | length')
if [ "$SUBJECT_COUNT" -lt 1 ]; then fail "Expected at least 1 subject, got $SUBJECT_COUNT"; fi
pass "Found $SUBJECT_COUNT subject(s)"

# 4. Run code (sample solution for sum-two-numbers)
echo ""
info "4. Running code via /run..."
SOLUTION='const input = require("fs").readFileSync(0, "utf8").trim();
const [a, b] = input.split(" ").map(Number);
console.log(a + b);'
RUN_RESP=$(curl -s -X POST "$GATEWAY/run" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "{\"problemId\":1,\"languageId\":1,\"code\":$(echo "$SOLUTION" | jq -Rs .)}")
RUN_STATUS=$(echo "$RUN_RESP" | jq -r '.data.status // "error"')
if [ "$RUN_STATUS" != "Accepted" ]; then
    echo "Run response: $RUN_RESP"
    fail "Run status is '$RUN_STATUS', expected 'Accepted'"
fi
pass "Run verdict: $RUN_STATUS"

# 5. Submit code
echo ""
info "5. Submitting code via /submit..."
SUBMIT_RESP=$(curl -s -X POST "$GATEWAY/submit" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "{\"problemId\":1,\"languageId\":1,\"code\":$(echo "$SOLUTION" | jq -Rs .)}")
SUBMISSION_ID=$(echo "$SUBMIT_RESP" | jq -r '.data.submission.id // empty')
if [ -z "$SUBMISSION_ID" ] || [ "$SUBMISSION_ID" = "null" ]; then
    echo "Submit response: $SUBMIT_RESP"
    fail "Failed to extract submission ID"
fi
pass "Submission created: $SUBMISSION_ID"

# 6. Poll submission until judged
echo ""
info "6. Polling submission $SUBMISSION_ID (timeout: ${TIMEOUT_SECONDS}s)..."
ELAPSED=0
STATUS="Pending"
while [ "$ELAPSED" -lt "$TIMEOUT_SECONDS" ]; do
    SUB_RESP=$(curl -s "$GATEWAY/submissions/$SUBMISSION_ID" -H "$AUTH_HEADER")
    STATUS=$(echo "$SUB_RESP" | jq -r '.data.submission.status // "error"')
    echo "  [${ELAPSED}s] Status: $STATUS"
    case "$STATUS" in
        "Accepted"|"Wrong Answer"|"Compilation Error"|"Runtime Error"|"Time Limit Exceeded"|"Internal Error")
            break
            ;;
    esac
    sleep "$POLL_INTERVAL"
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

# 7. Verify final status
echo ""
if [ "$STATUS" = "Accepted" ]; then
    pass "Final verdict: Accepted"
elif [ "$STATUS" = "Pending" ] || [ "$STATUS" = "Running" ]; then
    fail "Timed out after ${TIMEOUT_SECONDS}s — submission still $STATUS"
else
    fail "Unexpected final status: $STATUS"
fi

echo ""
echo -e "${GREEN}=== All E2E tests passed ===${NC}"
