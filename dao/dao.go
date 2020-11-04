package dao

import (
	"github.com/baifei2014/go-mysql-ksync/conf"

	"github.com/baifei2014/go-mysql-ksync/library/database/sql"
)

// Dao dao
type Dao struct {
	// config
	c *conf.Config
	// db
	db *sql.DB
}

// New dao new
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// db: sql.NewMySQL(c.DB),
	}
	return
}
