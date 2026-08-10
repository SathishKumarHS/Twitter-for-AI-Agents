// Main application entry point for the backend-go orchestrator service.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/config"
	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/schema"
	"github.com/twitter-for-ai-agents/backend-go/internal/dispatch"
	"github.com/twitter-for-ai-agents/backend-go/internal/events"
	"github.com/twitter-for-ai-agents/backend-go/internal/redis"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/evaluator"
	"github.com/twitter-for-ai-agents/backend-go/internal/scheduler/notifier"
	"github.com/twitter-for-ai-agents/backend-go/internal/server"
	"github.com/twitter-for-ai-agents/backend-go/internal/tools"
)

// Boots database, runs DDL migrations, initializes Redis/gRPC/Scheduler, and starts HTTP & gRPC servers
func main() {
	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := events.NewHub()

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

	dispatcher, err := dispatch.New(ctx, cfg.Worker.Address)
	if err != nil {
		slog.Error("failed to create dispatcher", "error", err)
		os.Exit(1)
	}
	defer dispatcher.Close()

	redisClient := redis.New(cfg.Redis.Addr, cfg.Redis.FeedTTL, cfg.Redis.AgentStatusTTL)
	defer redisClient.Close()

	eval := evaluator.New(
		db,
		dispatcher,
		redisClient,
		redisClient,
		cfg.Server.TurnTimeout,
	)

	sched := scheduler.New(
		cfg.Server.PollInterval,
		cfg.Server.PostInterval,
		cfg.Server.MaxConcurrent,
		db,
		eval,
	)
	sched.Start(ctx)
	defer sched.Stop()

	srv := server.New(cfg.Server.Address, db, hub, redisClient, eval)
	go func() {
		slog.Info("HTTP server started listening", "address", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("HTTP server stopped", "error", err)
		}
	}()

	notif := notifier.New(db.SubscriptionsDB, eval)

	// Tool execution gRPC server for Python AI Worker callbacks
	toolListener, err := net.Listen("tcp", cfg.Server.ToolsGRPCAddress)
	if err != nil {
		slog.Error("failed to listen for tool-execution gRPC", "error", err)
		os.Exit(1)
	}
	toolGRPC := grpc.NewServer()
	twitterforaiworkerpb.RegisterToolExecutorServer(toolGRPC, tools.NewToolGRPCServer(
		db.AgentsDB,
		db.PostDB,
		db.MemoriesDB,
		db.LastSeenDB,
		db.SubscriptionsDB,
		hub,
		redisClient,
		notif,
		cfg.Limits.FeedToolDefault,
		cfg.Limits.FeedToolMax,
	))
	go func() {
		slog.Info("Tool-execution gRPC server listening", "address", cfg.Server.ToolsGRPCAddress)
		if err := toolGRPC.Serve(toolListener); err != nil {
			slog.Error("Tool-execution gRPC server stopped", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down backend orchestrator service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	toolGRPC.GracefulStop()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown", "error", err)
	}
}
