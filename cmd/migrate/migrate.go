package main

import (
	"log/slog"
	"os"

	"travelando/internal"
	"travelando/internal/migrations"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
	"go.leapkit.dev/core/db"
)

// The migrate command is used to ship our application
// with the latest database schema migrator. which can be invoked
// by running `migrate`.
func main() {
	conn, err := internal.DBFn()
	if err != nil {
		slog.Error("connecting to the database", "error", err)
		os.Exit(1)
	}

	err = db.RunMigrations(migrations.All, conn)
	if err != nil {
		slog.Error("running migrations", "error", err)
		os.Exit(1)
	}
}
