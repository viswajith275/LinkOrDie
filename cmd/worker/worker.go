package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"link_or_die/internal/checker"
	"link_or_die/internal/config"
	"link_or_die/internal/store"
	"link_or_die/internal/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadConfig()

	db, err := store.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := store.CreateLinkRepo(db)

	w := worker.NewWorker(db, repo, checker.NewChecker(cfg), cfg, slog.Default())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := w.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
