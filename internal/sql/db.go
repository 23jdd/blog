package sql

import "github.com/jmoiron/sqlx"

func GetDB() *sqlx.DB {
	return db
}
