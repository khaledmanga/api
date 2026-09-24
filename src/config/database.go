package database

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"github.com/go-sql-driver/mysql"
)

type DatabaseStrategy interface {
	Connect() (*ent.Client, error)
}

type MySQLStrategy struct {
	cfg *config.Config
}

func NewMySQLStrategy(cfg *config.Config) *MySQLStrategy {
	return &MySQLStrategy{
		cfg: cfg,
	}
}

func (s *MySQLStrategy) Connect() (*ent.Client, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=True",
		s.cfg.Database.User,
		s.cfg.Database.Password,
		s.cfg.Database.Host,
		s.cfg.Database.Port,
		s.cfg.Database.Name,
	)

	return ent.Open(dialect.MySQL, dsn)
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
