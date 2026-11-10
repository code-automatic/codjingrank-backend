package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/auth_service/internal/dto"
	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/auth_service/internal/repository"
	"innogen-backend/shared/config"
	"innogen-backend/shared/middleware"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUUIDAuthAgainstSchema(t *testing.T) {
	url := os.Getenv("CATALOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set disposable database URL")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	cfg := &config.Config{AccessTokenTTLMinutes: 15, RefreshTokenTTLDays: 7, RefreshCookieName: "refresh", RefreshCookiePath: "/"}
	svc := New(repository.New(pool), jwt.NewService("test-secret"), cfg)
	result, err := svc.Register(ctx, dto.RegisterRequest{Email: "uuid-auth@example.org", Username: "uuid-auth", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, e := pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, result.User.ID); e != nil {
			t.Error(e)
		}
	})
	if !middleware.ValidUUID(result.User.ID) {
		t.Fatalf("not UUID: %v", result.User.ID)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/login", nil)
	login, err := svc.Login(ctx, rec, req, dto.LoginRequest{Email: "UUID-AUTH@EXAMPLE.ORG", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwt.NewService("test-secret").ParseToken(login.AccessToken)
	if err != nil || claims.UserID != result.User.ID {
		t.Fatalf("claims: %+v %v", claims, err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("missing refresh cookie")
	}
	refresh := httptest.NewRequest("POST", "/auth/refresh", nil)
	refresh.RemoteAddr = "127.0.0.1:12345"
	refresh.AddCookie(cookies[0])
	refreshed, err := svc.Refresh(ctx, httptest.NewRecorder(), refresh)
	if err != nil || refreshed.User.ID != result.User.ID {
		t.Fatalf("refresh: %+v %v", refreshed, err)
	}
	if _, err = svc.Refresh(ctx, httptest.NewRecorder(), refresh); err != ErrRefreshTokenRevoked {
		t.Fatalf("replay accepted: %v", err)
	}
	var tokens int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM refresh_tokens WHERE user_id=$1 AND revoked_at IS NULL`, result.User.ID).Scan(&tokens); err != nil || tokens != 1 {
		t.Fatalf("rotation lost active token: %d %v", tokens, err)
	}
}
