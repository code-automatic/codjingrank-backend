package service

import (
	"context"
	"errors"
	"innogen-backend/shared/catalog"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/dto"
	"innogen-backend/submission_service/internal/repository"
	"io"
	"log/slog"
	"testing"
)

func TestRejectedSubmitDoesNotPersistOrQueue(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	_, err := f.Pool.Exec(ctx, `UPDATE packages SET status='archived' WHERE id=$1`, f.PackageID)
	if err != nil {
		t.Fatal(err)
	}
	svc := New(repository.New(f.Pool), nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	_, err = svc.CreateSubmission(ctx, f.UserID, dto.SubmitRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "print(3)"})
	if !errors.Is(err, catalog.ErrProblemNotFound) {
		t.Fatalf("policy: %v", err)
	}
	var count int
	if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM submissions WHERE user_id=$1`, f.UserID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unauthorized submission persisted: %d %v", count, err)
	}
}
