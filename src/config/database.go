package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"api/src/ent"
	"entgo.io/ent/dialect"
	_ "github.com/go-sql-driver/mysql"
)

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
	return ent.Open(dialect.MySQL, mysqlDSN(s.cfg))
}

func NewSQLDB(cfg *DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", mysqlDSN(cfg))
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

func mysqlDSN(cfg *DatabaseConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
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
