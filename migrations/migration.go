package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var fs embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(fs)

	return goose.Up(db, ".")
}
