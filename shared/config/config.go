package config

import (
	"os"
	"strconv"
)

// Config holds all configuration values loaded from environment variables.
type Config struct {
	AppEnv                         string
	DatabaseURL                    string
	RedisAddr                      string
	JWTSecret                      string
	AuthServicePort                string
	APIGatewayPort                 string
	RunServicePort                 string
	SubmissionServicePort          string
	RepoServicePort                string
	PistonBaseURL                  string
	GitHubAppID                    string
	GitHubPrivateKeyPath           string
	GitHubDefaultBranch            string
	GitHubAPIBaseURL               string
	GitHubWebhookSecret            string
	GitHubAppName                  string
	GitHubAppInstallURL            string
	GitHubOAuthClientID            string
	GitHubOAuthClientSecret        string
	GitHubOAuthRedirectURL         string
	GitHubOAuthFrontendRedirectURL string
	FrontendURL                    string
	AuthServiceURL                 string
	RunServiceURL                  string
	SubmissionServiceURL           string
	RepoServiceURL                 string
	CORSAllowedOrigins             string
	MaxBodyBytes                   int
	RateLimitRequests              int
	RateLimitWindowSeconds         int
	UpstreamTimeoutSeconds         int
	AccessTokenTTLMinutes          int
	RefreshTokenTTLDays            int
	RefreshCookieName              string
	RefreshCookieSecure            bool
	RefreshCookieSameSite          string
	RefreshCookieDomain            string
	RefreshCookiePath              string
}

// defaults defines the fallback values for configuration keys.
var defaults = map[string]string{
	"APP_ENV":                            "development",
	"DATABASE_URL":                       "postgres://innogen:innogen@localhost:5432/innogen?sslmode=disable",
	"REDIS_ADDR":                         "localhost:6379",
	"JWT_SECRET":                         "innogen-dev-secret",
	"AUTH_SERVICE_PORT":                  "8081",
	"API_GATEWAY_PORT":                   "8080",
	"RUN_SERVICE_PORT":                   "8082",
	"SUBMISSION_SERVICE_PORT":            "8083",
	"REPO_SERVICE_PORT":                  "8084",
	"PISTON_BASE_URL":                    "http://localhost:2000",
	"GITHUB_APP_ID":                      "",
	"GITHUB_PRIVATE_KEY_PATH":            "",
	"GITHUB_DEFAULT_BRANCH":              "main",
	"GITHUB_API_BASE_URL":                "https://api.github.com",
	"GITHUB_WEBHOOK_SECRET":              "",
	"GITHUB_APP_NAME":                    "rinnogen",
	"GITHUB_APP_INSTALL_URL":             "https://github.com/apps/rinnogen/installations/new",
	"GITHUB_OAUTH_CLIENT_ID":             "",
	"GITHUB_OAUTH_CLIENT_SECRET":         "",
	"GITHUB_OAUTH_REDIRECT_URL":          "http://localhost:8080/github/oauth/callback",
	"GITHUB_OAUTH_FRONTEND_REDIRECT_URL": "http://localhost:5173/github",
	"FRONTEND_URL":                       "http://localhost:5173",
	"AUTH_SERVICE_URL":                   "http://localhost:8081",
	"RUN_SERVICE_URL":                    "http://localhost:8082",
	"SUBMISSION_SERVICE_URL":             "http://localhost:8083",
	"REPO_SERVICE_URL":                   "http://localhost:8084",
	"CORS_ALLOWED_ORIGINS":               "http://localhost:3000,http://localhost:5173",
	"MAX_BODY_BYTES":                     "1048576",
	"RATE_LIMIT_REQUESTS":                "120",
	"RATE_LIMIT_WINDOW_SECONDS":          "60",
	"UPSTREAM_TIMEOUT_SECONDS":           "30",
	"ACCESS_TOKEN_TTL_MINUTES":           "15",
	"REFRESH_TOKEN_TTL_DAYS":             "7",
	"REFRESH_COOKIE_NAME":                "refreshToken",
	"REFRESH_COOKIE_SECURE":              "false",
	"REFRESH_COOKIE_SAMESITE":            "Lax",
	"REFRESH_COOKIE_DOMAIN":              "",
	"REFRESH_COOKIE_PATH":                "/auth",
}

// getEnv returns the value of the environment variable named by key, or the
// default if the variable is not set or is empty.
func getEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaults[key]
}

// Load reads configuration from environment variables and returns a populated
// Config pointer. Missing or empty variables fall back to sensible defaults.
func Load() *Config {
	return &Config{
		AppEnv:                         getEnv("APP_ENV"),
		DatabaseURL:                    getEnv("DATABASE_URL"),
		RedisAddr:                      getEnv("REDIS_ADDR"),
		JWTSecret:                      getEnv("JWT_SECRET"),
		AuthServicePort:                getEnv("AUTH_SERVICE_PORT"),
		APIGatewayPort:                 getEnv("API_GATEWAY_PORT"),
		RunServicePort:                 getEnv("RUN_SERVICE_PORT"),
		SubmissionServicePort:          getEnv("SUBMISSION_SERVICE_PORT"),
		RepoServicePort:                getEnv("REPO_SERVICE_PORT"),
		PistonBaseURL:                  getEnv("PISTON_BASE_URL"),
		GitHubAppID:                    getEnv("GITHUB_APP_ID"),
		GitHubPrivateKeyPath:           getEnv("GITHUB_PRIVATE_KEY_PATH"),
		GitHubDefaultBranch:            getEnv("GITHUB_DEFAULT_BRANCH"),
		GitHubAPIBaseURL:               getEnv("GITHUB_API_BASE_URL"),
		GitHubWebhookSecret:            getEnv("GITHUB_WEBHOOK_SECRET"),
		GitHubAppName:                  getEnv("GITHUB_APP_NAME"),
		GitHubAppInstallURL:            getEnv("GITHUB_APP_INSTALL_URL"),
		GitHubOAuthClientID:            getEnv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret:        getEnv("GITHUB_OAUTH_CLIENT_SECRET"),
		GitHubOAuthRedirectURL:         getEnv("GITHUB_OAUTH_REDIRECT_URL"),
		GitHubOAuthFrontendRedirectURL: getEnv("GITHUB_OAUTH_FRONTEND_REDIRECT_URL"),
		FrontendURL:                    getEnv("FRONTEND_URL"),
		AuthServiceURL:                 getEnv("AUTH_SERVICE_URL"),
		RunServiceURL:                  getEnv("RUN_SERVICE_URL"),
		SubmissionServiceURL:           getEnv("SUBMISSION_SERVICE_URL"),
		RepoServiceURL:                 getEnv("REPO_SERVICE_URL"),
		CORSAllowedOrigins:             getEnv("CORS_ALLOWED_ORIGINS"),
		MaxBodyBytes:                   atoi(getEnv("MAX_BODY_BYTES")),
		RateLimitRequests:              atoi(getEnv("RATE_LIMIT_REQUESTS")),
		RateLimitWindowSeconds:         atoi(getEnv("RATE_LIMIT_WINDOW_SECONDS")),
		UpstreamTimeoutSeconds:         atoi(getEnv("UPSTREAM_TIMEOUT_SECONDS")),
		AccessTokenTTLMinutes:          atoi(getEnv("ACCESS_TOKEN_TTL_MINUTES")),
		RefreshTokenTTLDays:            atoi(getEnv("REFRESH_TOKEN_TTL_DAYS")),
		RefreshCookieName:              getEnv("REFRESH_COOKIE_NAME"),
		RefreshCookieSecure:            getEnv("REFRESH_COOKIE_SECURE") == "true",
		RefreshCookieSameSite:          getEnv("REFRESH_COOKIE_SAMESITE"),
		RefreshCookieDomain:            getEnv("REFRESH_COOKIE_DOMAIN"),
		RefreshCookiePath:              getEnv("REFRESH_COOKIE_PATH"),
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
