package config

import (
	"api/src/ent"
	 "entgo.io/ent/dialect/sql"
)

type DatabaseStrategy interface {
	Connect() (*ent.Client, error)
}

type MySQLStrategy struct {}

func (s MySQLStrategy) 
