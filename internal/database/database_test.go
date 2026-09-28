package database

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

	"api/internal/config"
)

func TestMySQLDSNEnablesTLSWhenConfigured(t *testing.T) {
	dsn, err := MySQLDSN(&config.DatabaseConfig{
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
	dsn, err := MySQLDSN(&config.DatabaseConfig{
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
	if _, err := CertificatePool(escapedPEM); err != nil {
		t.Fatalf("certificate pool: %v", err)
	}
}

func TestCertificatePoolRejectsInvalidPEM(t *testing.T) {
	if _, err := CertificatePool("not a certificate"); err == nil {
		t.Fatal("expected invalid CA PEM to be rejected")
	}
}

func TestTLSCARequiresTLS(t *testing.T) {
	_, err := MySQLDSN(&config.DatabaseConfig{TLSCA: "certificate"})
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
