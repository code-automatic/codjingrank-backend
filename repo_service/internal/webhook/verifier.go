package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifySignature checks the X-Hub-Signature-256 header against the request body.
// Returns true if the signature is valid, false otherwise (including when secret is empty).
func VerifySignature(body []byte, signatureHeader string, secret []byte) bool {
	if len(secret) == 0 || signatureHeader == "" {
		return false
	}
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	expectedMAC, err := hex.DecodeString(signatureHeader[7:])
	if err != nil || len(expectedMAC) == 0 {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	computedMAC := mac.Sum(nil)
	return hmac.Equal(computedMAC, expectedMAC)
}
