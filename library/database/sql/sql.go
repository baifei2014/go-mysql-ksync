package sql

import (
	// "time"
	"database/sql"
)

type DB struct {
	write  *conn
	read   []*conn
	idx    int64
	master *DB
}

// conn database connection
type conn struct {
	*sql.DB
	conf *Config
}

func Open(c *Config) (*sql.DB, error) {
	db, err := connect(c, c.DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connect(c *Config, dataSourceName string) (*sql.DB, error) {
	d, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(c.Active)
	d.SetMaxIdleConns(c.Idle)
	// d.SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	return d, nil
}
