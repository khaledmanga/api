package config

import (
	"strings"
	"testing"
)

func TestApplyEnvironmentUsesVercelPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("PORT", "3001")

	cfg := &Config{Server: ServerConfig{Port: 8080}}
	if err := applyEnvironment(cfg); err != nil {
		t.Fatalf("apply environment: %v", err)
	}
	if cfg.Server.Port != 3001 {
		t.Fatalf("server port = %d, want 3001", cfg.Server.Port)
	}
}

func TestApplyEnvironmentParsesTLSFlags(t *testing.T) {
	t.Setenv("DB_TLS", "true")
	t.Setenv("REDIS_TLS", "true")

	cfg := &Config{}
	if err := applyEnvironment(cfg); err != nil {
		t.Fatalf("apply environment: %v", err)
	}
	if !cfg.Database.TLS || !cfg.Redis.TLS {
		t.Fatalf("TLS flags = database:%t redis:%t, want both true", cfg.Database.TLS, cfg.Redis.TLS)
	}
}

func TestApplyEnvironmentRejectsInvalidTLSFlag(t *testing.T) {
	t.Setenv("DB_TLS", "required")

	err := applyEnvironment(&Config{})
	if err == nil || !strings.Contains(err.Error(), "parse DB_TLS") {
		t.Fatalf("error = %v, want DB_TLS parsing error", err)
	}
}

func TestApplyEnvironmentParsesSessionCookieSameSite(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SAME_SITE", " none ")

	cfg := &Config{}
	if err := applyEnvironment(cfg); err != nil {
		t.Fatalf("apply environment: %v", err)
	}
	if cfg.SessionCookieSameSite != "None" {
		t.Fatalf("SameSite = %q, want None", cfg.SessionCookieSameSite)
	}
}

func TestApplyEnvironmentDefaultsSessionCookieSameSiteToNone(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SAME_SITE", "")

	cfg := &Config{}
	if err := applyEnvironment(cfg); err != nil {
		t.Fatalf("apply environment: %v", err)
	}
	if cfg.SessionCookieSameSite != "None" {
		t.Fatalf("SameSite = %q, want None", cfg.SessionCookieSameSite)
	}
}

func TestApplyEnvironmentRejectsInvalidSessionCookieSameSite(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SAME_SITE", "unsafe")

	err := applyEnvironment(&Config{})
	if err == nil || !strings.Contains(err.Error(), "SESSION_COOKIE_SAME_SITE") {
		t.Fatalf("error = %v, want SESSION_COOKIE_SAME_SITE validation error", err)
	}
}
