package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"

	"transport-auth/internal/config"
	apphttp "transport-auth/internal/http"
)

func main() {
	cfg := config.FromEnv()

	dsn := cfg.DBPath
	if strings.Contains(dsn, "?") {
		dsn += "&_fk=1"
	} else {
		dsn += "?_fk=1"
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, cfg.MigrationsPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	if err := ensureDefaultAdmin(db); err != nil {
		log.Fatalf("ensure default admin: %v", err)
	}

	server := apphttp.NewServer(db, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("api started on %s", cfg.AppAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}
}

func ensureDefaultAdmin(db *sql.DB) error {
	const query = `
INSERT INTO users(login, password_hash, is_admin)
VALUES ('admin', '$2a$10$FqMweFNuPDE0cePwW7JgtuQnHfT8f0VAAxV2n8BmlqK4wRklS1z/u', 1)
ON CONFLICT(login) DO UPDATE SET
	password_hash=excluded.password_hash,
	is_admin=1;
`
	_, err := db.Exec(query)
	return err
}
