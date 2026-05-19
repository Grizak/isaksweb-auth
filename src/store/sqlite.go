package store

import (
	"database/sql"
	"embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema embed.FS

var DB *sql.DB

func Init(dsn string) error {
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	DB.Exec("PRAGMA journal_mode=WAl;")
	DB.Exec("PRAGMA foreign_keys=ON;")

	return runMigrations()
}

func runMigrations() error {
	schema, err := schema.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	_, err = DB.Exec(string(schema))
	return err
}
