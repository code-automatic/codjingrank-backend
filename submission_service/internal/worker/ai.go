package worker

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"
)

func notifyAI(ctx context.Context, submissionID string) {
	url, secret := os.Getenv("AI_SERVICE_URL"), os.Getenv("AI_INTERNAL_SECRET")
	if url == "" || secret == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(url, "/")+"/internal/reviews/"+submissionID, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := (&http.Client{Timeout: time.Second}).Do(req)
	if err == nil {
		resp.Body.Close()
	}
}
