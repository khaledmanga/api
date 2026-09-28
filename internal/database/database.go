package database

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"api/internal/config"
	"api/ent"
	"entgo.io/ent/dialect"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

var registeredMySQLTLSConfigs = struct {
	sync.Mutex
	names map[string]struct{}
}{names: make(map[string]struct{})}

type Strategy interface {
	Connect() (*ent.Client, error)
}

type MySQLStrategy struct {
	cfg *config.DatabaseConfig
}

func NewMySQLStrategy(cfg *config.DatabaseConfig) *MySQLStrategy {
	return &MySQLStrategy{
		cfg: cfg,
	}
}

func (s *MySQLStrategy) Connect() (*ent.Client, error) {
	dsn, err := MySQLDSN(s.cfg)
	if err != nil {
		return nil, fmt.Errorf("build MySQL DSN: %w", err)
	}
	client, err := ent.Open(dialect.MySQL, dsn)
	if err != nil {
		return nil, fmt.Errorf("open ent MySQL connection: %w", err)
	}
	return client, nil
}

func NewSQLDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn, err := MySQLDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build MySQL DSN: %w", err)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sql database: %w", err)
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func MySQLDSN(cfg *config.DatabaseConfig) (string, error) {
	mysqlConfig := mysqlDriver.NewConfig()
	mysqlConfig.User = cfg.User
	mysqlConfig.Passwd = cfg.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlConfig.DBName = cfg.Name
	mysqlConfig.ParseTime = true

	if cfg.TLS {
		mysqlConfig.TLSConfig = "true"
		if cfg.TLSCA != "" {
			tlsConfigName, err := registerMySQLTLSConfig(cfg)
			if err != nil {
				return "", fmt.Errorf("register MySQL TLS: %w", err)
			}
			mysqlConfig.TLSConfig = tlsConfigName
		}
	} else if cfg.TLSCA != "" {
		return "", fmt.Errorf("DB_TLS_CA requires DB_TLS=true")
	}
	return mysqlConfig.FormatDSN(), nil
}

func registerMySQLTLSConfig(cfg *config.DatabaseConfig) (string, error) {
	rootCAs, err := CertificatePool(cfg.TLSCA)
	if err != nil {
		return "", fmt.Errorf("configure MySQL TLS CA: %w", err)
	}
	hash := sha256.Sum256([]byte(cfg.Host + "\x00" + cfg.TLSCA))
	name := fmt.Sprintf("custom-%x", hash[:])

	registeredMySQLTLSConfigs.Lock()
	defer registeredMySQLTLSConfigs.Unlock()
	if _, exists := registeredMySQLTLSConfigs.names[name]; !exists {
		if err := mysqlDriver.RegisterTLSConfig(name, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.Host,
			RootCAs:    rootCAs,
		}); err != nil {
			return "", fmt.Errorf("register MySQL TLS configuration: %w", err)
		}
		registeredMySQLTLSConfigs.names[name] = struct{}{}
	}
	return name, nil
}

func CertificatePool(caPEM string) (*x509.CertPool, error) {
	caPEM = strings.ReplaceAll(caPEM, `\n`, "\n")
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(caPEM)) {
		return nil, fmt.Errorf("CA value contains no valid PEM certificates")
	}
	return rootCAs, nil
}

type Database struct {
	strategy Strategy
}

func NewDatabase(strategy Strategy) *Database {
	return &Database{
		strategy: strategy,
	}
}

func (d *Database) Connect() (*ent.Client, error) {
	return d.strategy.Connect()
}
