package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/example/go-cli-login/internal/auth"
	"github.com/example/go-cli-login/internal/cli"
	"github.com/example/go-cli-login/internal/config"
	"github.com/example/go-cli-login/internal/db"
	"github.com/example/go-cli-login/internal/session"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll("data", 0700); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create data directory:", err)
		os.Exit(1)
	}

	database, err := sql.Open("sqlite3", cfg.DatabasePath+"?_foreign_keys=on")
	if err != nil {
		fmt.Fprintln(os.Stderr, "database error:", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		fmt.Fprintln(os.Stderr, "migration error:", err)
		os.Exit(1)
	}

	repo := db.NewRepository(database)
	authService := auth.NewService(repo, cfg)
	sessionManager := session.NewManager(repo, cfg)

	app := cli.NewApp(authService, sessionManager, cfg)
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "application error:", err)
		os.Exit(1)
	}
}
