package dto

import (
	"strings"
	"testing"
)

func TestRunCodeByteLimit(t *testing.T) {
	req := RunRequest{ProblemID: "00000000-0000-0000-0000-000000000001", LanguageID: "00000000-0000-0000-0000-000000000002", Code: strings.Repeat("é", 51200)}
	if _, msg := req.Validate(); msg != "" {
		t.Fatal(msg)
	}
	req.Code += "a"
	if field, msg := req.Validate(); field != "code" || msg == "" {
		t.Fatal("source over byte limit accepted")
	}
}
