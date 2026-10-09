package infra

import (
	"errors"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	dbOnce      sync.Once
	database    *sqlx.DB
	databaseErr error
)

func NewConnection(connectionString string) (*sqlx.DB, error) {
	dbOnce.Do(func() {
		database, databaseErr = sqlx.Connect("pgx", connectionString)
		if databaseErr != nil {
			fmt.Println("DB connection error:", databaseErr)
			return
		}
		fmt.Println("✅ Successfully connected to PostgreSQL!")
	})
	if databaseErr != nil {
		return nil, databaseErr
	}
	if database == nil {
		return nil, errors.New("database connection is not initialized")
	}
	return database, nil
}
