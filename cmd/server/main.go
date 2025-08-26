package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"tinyurl/internal/app"
	"tinyurl/internal/config"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.Build(ctx, cfg)
	if err != nil {
		panic(err)
	}
	a.Log.Info("server starting", "addr", cfg.HTTP.Address)

	go func() {
		if err := a.Server.Start(); err != nil {
			a.Log.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	a.Log.Info("shutting down...")

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(shCtx); err != nil {
		a.Log.Error("shutdown error", "err", err)
	}
	if err := a.DBClose(); err != nil {
		a.Log.Error("db close error", "err", err)
	}
	a.Log.Info("bye")
}
