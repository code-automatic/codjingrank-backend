#!/usr/bin/env bash
set -euo pipefail

# Test GitHub webhook locally with HMAC-SHA256 signature
# Requires: GITHUB_WEBHOOK_SECRET env var, openssl

SECRET="${GITHUB_WEBHOOK_SECRET:-}"
if [ -z "$SECRET" ]; then
  echo "Error: GITHUB_WEBHOOK_SECRET is not set"
  exit 1
fi

BODY='{"action":"created","installation":{"id":123,"account":{"login":"test","type":"User","id":456,"avatar_url":""}}}'
SIG=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')

echo "Sending installation.created webhook..."
curl -s -X POST http://localhost:8084/webhooks/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: installation" \
  -H "X-GitHub-Delivery: test-$(date +%s)" \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo ""

echo ""
echo "Done. Check repo_service logs for details."
