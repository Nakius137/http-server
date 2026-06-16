package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"http/v1/dev/internal/config"
	"http/v1/dev/internal/server"
)

func main() {
	logger := slog.Default()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed loading the config", "error", err)
		return
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed connecting to the db", "error", err)
		return
	}
	defer db.Close()

	dbCtx, dbStop := context.WithTimeout(ctx, 10*time.Second)
	defer dbStop()

	if err := db.PingContext(dbCtx); err != nil {
		logger.Error("failed pinging the db", "error", err)
		return
	}

	srv, err := server.NewServer(
		server.WithLogger(logger),
		server.WithDatabase(db),
	)
	if err != nil {
		logger.Error("error with initialize server", "error", err)
		return
	}

	err = srv.Run(ctx)
	if err != nil {
		logger.Error("error with run server", "error", err)
		return
	}
}
