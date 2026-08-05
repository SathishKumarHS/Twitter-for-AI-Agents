package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/twitter-for-ai-agents/backend-go/internal/config"
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/schema"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/logs"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler"
	"github.com/twitter-for-ai-agents/backend-go/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := events.NewHub()
	logs.Init(hub)

	db, err := db.Connect(ctx, cfg.DB.URL)
	if err != nil {
		slog.Error("connect database failed", "error", err)
		os.Exit(1)
	}
	if err := schema.Migrate(ctx, db); err != nil {
		slog.Error("migrate database failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logs.Info(logs.SystemNamespace, "Database connected & migrations verified successfully")

	dispatcher, err := dispatch.New(ctx, cfg.Worker.Address)
	if err != nil {
		slog.Error("failed to create dispatcher", "error", err)
		os.Exit(1)
	}
	defer dispatcher.Close()

	redisClient := redis.New(cfg.Redis.Addr, cfg.Redis.FeedTTL, cfg.Redis.AgentStatusTTL)
	defer redisClient.Close()

	sched := scheduler.New(cfg.Server.PollInterval, cfg.Server.PostInterval, cfg.Server.TurnTimeout, cfg.Server.MaxConcurrent, redisClient, db, dispatcher, hub, redisClient)
	sched.Start(ctx)
	defer sched.Stop()

	srv := server.New(cfg.Server.Address, db, hub, redisClient)
	go func() {
		logs.Info(logs.SystemNamespace, fmt.Sprintf("HTTP Server started listening on %s", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != nil {
			logs.Info(logs.SystemNamespace, fmt.Sprintf("HTTP server stopped: %v", err))
		}
	}()

	<-ctx.Done()
	logs.Info(logs.SystemNamespace, "Shutting down backend orchestrator service")
}
