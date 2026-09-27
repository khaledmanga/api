package config

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

	"api/src/ent"
	"entgo.io/ent/dialect"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

var registeredMySQLTLSConfigs = struct {
	sync.Mutex
	names map[string]struct{}
}{names: make(map[string]struct{})}

type DatabaseStrategy interface {
	Connect() (*ent.Client, error)
}

type MySQLStrategy struct {
	cfg *DatabaseConfig
}

func NewMySQLStrategy(cfg *DatabaseConfig) *MySQLStrategy {
	return &MySQLStrategy{
		cfg: cfg,
	}
}

func (s *MySQLStrategy) Connect() (*ent.Client, error) {
	dsn, err := mysqlDSN(s.cfg)
	if err != nil {
		return nil, err
	}
	return ent.Open(dialect.MySQL, dsn)
}

func NewSQLDB(cfg *DatabaseConfig) (*sql.DB, error) {
	dsn, err := mysqlDSN(cfg)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func mysqlDSN(cfg *DatabaseConfig) (string, error) {
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
				return "", err
			}
			mysqlConfig.TLSConfig = tlsConfigName
		}
	} else if cfg.TLSCA != "" {
		return "", fmt.Errorf("DB_TLS_CA requires DB_TLS=true")
	}
	return mysqlConfig.FormatDSN(), nil
}

func registerMySQLTLSConfig(cfg *DatabaseConfig) (string, error) {
	rootCAs, err := certificatePool(cfg.TLSCA)
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

func certificatePool(caPEM string) (*x509.CertPool, error) {
	caPEM = strings.ReplaceAll(caPEM, `\n`, "\n")
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(caPEM)) {
		return nil, fmt.Errorf("CA value contains no valid PEM certificates")
	}
	return rootCAs, nil
}

type Database struct {
	strategy DatabaseStrategy
}

func NewDatabase(strategy DatabaseStrategy) *Database {
	return &Database{
		strategy: strategy,
	}
}

func (d *Database) Connect() (*ent.Client, error) {
	return d.strategy.Connect()
}
