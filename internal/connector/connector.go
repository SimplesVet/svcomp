package connector

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConnector struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string) (*MySQLConnector, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connector: open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connector: ping: %w", err)
	}
	return &MySQLConnector{db: db}, nil
}

func (c *MySQLConnector) DB() *sql.DB {
	return c.db
}

func (c *MySQLConnector) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}
