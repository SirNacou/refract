package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear all relevant env vars
	clearEnv(t)

	// Set only required fields
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer os.Unsetenv("ZITADEL_AUDIENCE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port=8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected Server.Host=0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port=5432, got %d", cfg.Database.Port)
	}

	if cfg.Redis.Port != 6379 {
		t.Errorf("expected Redis.Port=6379, got %d", cfg.Redis.Port)
	}

	if cfg.Worker.WorkerID != 0 {
		t.Errorf("expected Worker.WorkerID=0, got %d", cfg.Worker.WorkerID)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("expected Logging.Level=info, got %s", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("expected Logging.Format=json, got %s", cfg.Logging.Format)
	}

	// Verify Zitadel default issuer
	if cfg.Zitadel.Issuer != "https://zitadel.nacou.uk" {
		t.Errorf("expected Zitadel.Issuer default, got %s", cfg.Zitadel.Issuer)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnv(t)

	// Set custom values
	os.Setenv("API_PORT", "9000")
	os.Setenv("API_BASE_URL", "https://api.example.com")
	os.Setenv("POSTGRES_HOST", "db.example.com")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("ZITADEL_ISSUER", "https://auth.example.com")
	os.Setenv("ZITADEL_AUDIENCE", "custom-audience")
	os.Setenv("ZITADEL_CLIENT_ID", "custom-client-id")
	os.Setenv("ZITADEL_CLIENT_SECRET", "custom-client-secret")
	os.Setenv("WORKER_ID", "42")
	os.Setenv("LOG_LEVEL", "debug")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("expected Server.Port=9000, got %d", cfg.Server.Port)
	}

	if cfg.Server.BaseURL != "https://api.example.com" {
		t.Errorf("expected Server.BaseURL=https://api.example.com, got %s", cfg.Server.BaseURL)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected Database.Host=db.example.com, got %s", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("expected Database.Port=5433, got %d", cfg.Database.Port)
	}

	if cfg.Redis.Host != "redis.example.com" {
		t.Errorf("expected Redis.Host=redis.example.com, got %s", cfg.Redis.Host)
	}

	if cfg.Redis.Port != 6380 {
		t.Errorf("expected Redis.Port=6380, got %d", cfg.Redis.Port)
	}

	if cfg.Zitadel.Issuer != "https://auth.example.com" {
		t.Errorf("expected Zitadel.Issuer=https://auth.example.com, got %s", cfg.Zitadel.Issuer)
	}

	if cfg.Zitadel.Audience != "custom-audience" {
		t.Errorf("expected Zitadel.Audience=custom-audience, got %s", cfg.Zitadel.Audience)
	}

	if cfg.Worker.WorkerID != 42 {
		t.Errorf("expected Worker.WorkerID=42, got %d", cfg.Worker.WorkerID)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected Logging.Level=debug, got %s", cfg.Logging.Level)
	}
}

func TestLoad_DatabaseURL(t *testing.T) {
	clearEnv(t)

	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db?sslmode=require")
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Database.DatabaseURL != "postgres://user:pass@host:5432/db?sslmode=require" {
		t.Errorf("expected DatabaseURL to be set, got %s", cfg.Database.DatabaseURL)
	}

	dsn := cfg.Database.GetDatabaseDSN()
	if dsn != "postgres://user:pass@host:5432/db?sslmode=require" {
		t.Errorf("expected GetDatabaseDSN() to return DATABASE_URL, got %s", dsn)
	}
}

func TestValidate_MissingZitadelAudience(t *testing.T) {
	clearEnv(t)

	// ZITADEL_ISSUER has a default, but ZITADEL_AUDIENCE does not
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when ZITADEL_AUDIENCE is not set")
	}

	if !contains(err.Error(), "ZITADEL_AUDIENCE") {
		t.Errorf("expected error to mention ZITADEL_AUDIENCE, got: %v", err)
	}
}

func TestValidate_InvalidWorkerID(t *testing.T) {
	clearEnv(t)

	os.Setenv("WORKER_ID", "1024") // Out of range
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when WORKER_ID is out of range")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	clearEnv(t)

	os.Setenv("API_PORT", "70000") // Out of range
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when API_PORT is out of range")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	clearEnv(t)

	os.Setenv("LOG_LEVEL", "invalid")
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when LOG_LEVEL is invalid")
	}
}

func TestGetDatabaseDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	dsn := cfg.GetDatabaseDSN()

	if dsn != expected {
		t.Errorf("expected DSN=%s, got %s", expected, dsn)
	}
}

func TestGetRedisAddr(t *testing.T) {
	cfg := RedisConfig{
		Host: "redis.example.com",
		Port: 6380,
	}

	expected := "redis.example.com:6380"
	addr := cfg.GetRedisAddr()

	if addr != expected {
		t.Errorf("expected Redis addr=%s, got %s", expected, addr)
	}
}

func TestServerAddress(t *testing.T) {
	cfg := ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	}

	expected := "0.0.0.0:8080"
	addr := cfg.Address()

	if addr != expected {
		t.Errorf("expected Server addr=%s, got %s", expected, addr)
	}
}

func TestRedisConfig_DefaultCacheSettings(t *testing.T) {
	clearEnv(t)

	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Redis.CacheTTL != 1*time.Hour {
		t.Errorf("expected Redis.CacheTTL=1h, got %v", cfg.Redis.CacheTTL)
	}

	if cfg.Redis.MaxRetries != 3 {
		t.Errorf("expected Redis.MaxRetries=3, got %d", cfg.Redis.MaxRetries)
	}

	if cfg.Redis.PoolSize != 10 {
		t.Errorf("expected Redis.PoolSize=10, got %d", cfg.Redis.PoolSize)
	}
}

func TestSecurityConfig_DefaultRateLimits(t *testing.T) {
	clearEnv(t)

	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Security.RateLimitPerUser != 100 {
		t.Errorf("expected Security.RateLimitPerUser=100, got %d", cfg.Security.RateLimitPerUser)
	}

	if cfg.Security.RateLimitPerAPIKey != 1000 {
		t.Errorf("expected Security.RateLimitPerAPIKey=1000, got %d", cfg.Security.RateLimitPerAPIKey)
	}

	if cfg.Security.RateLimitWindow != 1*time.Hour {
		t.Errorf("expected Security.RateLimitWindow=1h, got %v", cfg.Security.RateLimitWindow)
	}
}

func TestSecurityConfig_CORSArrayParsing(t *testing.T) {
	clearEnv(t)

	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://example.com,https://app.example.com")
	os.Setenv("ZITADEL_AUDIENCE", "test-audience")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	expectedOrigins := []string{
		"http://localhost:3000",
		"https://example.com",
		"https://app.example.com",
	}

	if len(cfg.Security.CORSAllowedOrigins) != len(expectedOrigins) {
		t.Errorf("expected %d CORS origins, got %d", len(expectedOrigins), len(cfg.Security.CORSAllowedOrigins))
	}

	for i, expected := range expectedOrigins {
		if cfg.Security.CORSAllowedOrigins[i] != expected {
			t.Errorf("expected CORS origin[%d]=%s, got %s", i, expected, cfg.Security.CORSAllowedOrigins[i])
		}
	}
}

func TestZitadelConfig_AllFields(t *testing.T) {
	clearEnv(t)

	os.Setenv("ZITADEL_ISSUER", "https://zitadel.example.com")
	os.Setenv("ZITADEL_AUDIENCE", "my-api")
	os.Setenv("ZITADEL_CLIENT_ID", "my-client-id")
	os.Setenv("ZITADEL_CLIENT_SECRET", "my-client-secret")
	defer clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Zitadel.Issuer != "https://zitadel.example.com" {
		t.Errorf("expected Zitadel.Issuer=https://zitadel.example.com, got %s", cfg.Zitadel.Issuer)
	}

	if cfg.Zitadel.Audience != "my-api" {
		t.Errorf("expected Zitadel.Audience=my-api, got %s", cfg.Zitadel.Audience)
	}

	if cfg.Zitadel.ClientID != "my-client-id" {
		t.Errorf("expected Zitadel.ClientID=my-client-id, got %s", cfg.Zitadel.ClientID)
	}

	if cfg.Zitadel.ClientSecret != "my-client-secret" {
		t.Errorf("expected Zitadel.ClientSecret=my-client-secret, got %s", cfg.Zitadel.ClientSecret)
	}
}

func TestValidate_MissingZitadelIssuer(t *testing.T) {
	// Test that validation correctly rejects empty Zitadel.Issuer
	cfg := &Config{
		Server: ServerConfig{
			Port:    8080,
			Host:    "0.0.0.0",
			BaseURL: "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Database: "test",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		Zitadel: ZitadelConfig{
			Issuer:   "", // Empty - should fail
			Audience: "my-api",
		},
		Worker: WorkerConfig{
			WorkerID: 0,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when Zitadel.Issuer is empty")
	}

	if !contains(err.Error(), "ZITADEL_ISSUER") {
		t.Errorf("expected error to mention ZITADEL_ISSUER, got: %v", err)
	}
}

// clearEnv clears all environment variables used by config
func clearEnv(t *testing.T) {
	t.Helper()

	vars := []string{
		"API_PORT", "API_BASE_URL", "API_HOST",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB",
		"POSTGRES_SSLMODE", "POSTGRES_MAX_OPEN_CONNS", "POSTGRES_MAX_IDLE_CONNS",
		"POSTGRES_CONN_MAX_LIFETIME", "POSTGRES_CONN_MAX_IDLE_TIME", "DATABASE_URL",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"REDIS_CACHE_TTL", "REDIS_MAX_RETRIES", "REDIS_POOL_SIZE",
		"REDIS_MIN_IDLE_CONNS", "REDIS_CONN_MAX_IDLE_TIME",
		"ZITADEL_ISSUER", "ZITADEL_AUDIENCE", "ZITADEL_CLIENT_ID", "ZITADEL_CLIENT_SECRET",
		"WORKER_ID",
		"SAFE_BROWSING_API_KEY", "RATE_LIMIT_PER_USER", "RATE_LIMIT_PER_API_KEY",
		"RATE_LIMIT_WINDOW", "CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_METHODS", "CORS_ALLOWED_HEADERS",
		"LOG_LEVEL", "LOG_FORMAT",
	}

	for _, v := range vars {
		os.Unsetenv(v)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
