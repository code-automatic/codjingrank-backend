package response

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	for _, body := range []string{`{"code":"x","wrapperCode":"override"}`, `{"code":"x"} {"code":"second"}`, `null`} {
		var req struct {
			Code string `json:"code"`
		}
		if err := DecodeJSON(httptest.NewRequest("POST", "/", strings.NewReader(body)), &req); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}
