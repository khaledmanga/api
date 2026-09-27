package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
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

func TestApplyEnvironmentDefaultsSessionCookieSameSiteToStrict(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SAME_SITE", "")

	cfg := &Config{}
	if err := applyEnvironment(cfg); err != nil {
		t.Fatalf("apply environment: %v", err)
	}
	if cfg.SessionCookieSameSite != "Strict" {
		t.Fatalf("SameSite = %q, want Strict", cfg.SessionCookieSameSite)
	}
}

func TestApplyEnvironmentRejectsInvalidSessionCookieSameSite(t *testing.T) {
	t.Setenv("SESSION_COOKIE_SAME_SITE", "unsafe")

	err := applyEnvironment(&Config{})
	if err == nil || !strings.Contains(err.Error(), "SESSION_COOKIE_SAME_SITE") {
		t.Fatalf("error = %v, want SESSION_COOKIE_SAME_SITE validation error", err)
	}
}

func TestMySQLDSNEnablesTLSWhenConfigured(t *testing.T) {
	dsn, err := mysqlDSN(&DatabaseConfig{
		Host: "mysql.example.com",
		Port: 10204,
		User: "user",
		Name: "defaultdb",
		TLS:  true,
	})
	if err != nil {
		t.Fatalf("mysql DSN: %v", err)
	}
	if !strings.Contains(dsn, "parseTime=true") || !strings.Contains(dsn, "tls=true") {
		t.Fatalf("DSN = %q, want TLS option enabled", dsn)
	}
}

func TestMySQLDSNRegistersCustomTLSCA(t *testing.T) {
	caPEM := testCertificatePEM(t)
	dsn, err := mysqlDSN(&DatabaseConfig{
		Host:  "mysql.example.com",
		Port:  10204,
		User:  "user",
		Name:  "defaultdb",
		TLS:   true,
		TLSCA: caPEM,
	})
	if err != nil {
		t.Fatalf("mysql DSN: %v", err)
	}
	if !strings.Contains(dsn, "tls=custom-") {
		t.Fatalf("DSN = %q, want custom TLS configuration", dsn)
	}
}

func TestCertificatePoolAcceptsEscapedPEMNewlines(t *testing.T) {
	caPEM := testCertificatePEM(t)
	escapedPEM := strings.ReplaceAll(caPEM, "\n", `\n`)
	if _, err := certificatePool(escapedPEM); err != nil {
		t.Fatalf("certificate pool: %v", err)
	}
}

func TestCertificatePoolRejectsInvalidPEM(t *testing.T) {
	if _, err := certificatePool("not a certificate"); err == nil {
		t.Fatal("expected invalid CA PEM to be rejected")
	}
}

func TestTLSCARequiresTLS(t *testing.T) {
	_, err := mysqlDSN(&DatabaseConfig{TLSCA: "certificate"})
	if err == nil || !strings.Contains(err.Error(), "DB_TLS_CA requires DB_TLS=true") {
		t.Fatalf("error = %v, want DB_TLS requirement", err)
	}
}

func testCertificatePEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create CA certificate: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}))
}
