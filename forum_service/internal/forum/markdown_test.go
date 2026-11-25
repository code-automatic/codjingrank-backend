package forum

import (
	"strings"
	"testing"
)

func TestMarkdownRejectsUnsafeDestinationsAndEscapesHTML(t *testing.T) {
	for _, input := range []string{"[link](javascript:alert)", "![x](data:text/html,abc)", "[link](java&#x73;cript:alert)", "[link][x]\n\n[x]: javascript:alert"} {
		if _, err := safeMarkdown(input); err == nil {
			t.Fatalf("unsafe destination accepted: %q", input)
		}
	}
	got, err := safeMarkdown("hello <script>alert(1)</script> [site](https://example.org)")
	if err != nil || strings.Contains(got, "<script>") || !strings.Contains(got, "https://example.org") {
		t.Fatalf("unsafe output %q: %v", got, err)
	}
	code := "```go\nif a < b && ready {}\n```\n\n`<tag>`"
	got, err = safeMarkdown(code)
	if err != nil || got != code {
		t.Fatalf("code corrupted: %q %v", got, err)
	}
}
