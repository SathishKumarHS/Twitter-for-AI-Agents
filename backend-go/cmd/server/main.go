package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/agent-feed/backend-go/internal/config"
	"github.com/agent-feed/backend-go/internal/db"
	"github.com/agent-feed/backend-go/internal/db/schema"
	"github.com/agent-feed/backend-go/internal/dispatch"
	"github.com/agent-feed/backend-go/internal/events"
	"github.com/agent-feed/backend-go/internal/httpapi"
	"github.com/agent-feed/backend-go/internal/logs"
	"github.com/agent-feed/backend-go/internal/redis"
	"github.com/agent-feed/backend-go/internal/sched"
)

func main() {
	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := db.Connect(ctx, cfg.DB.URL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := schema.Migrate(ctx, db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	defer db.Close()

	hub := events.NewHub()
	logs.Init(hub)
	logs.Info("system", "Database connected & migrations verified successfully")

	dispatcher, err := dispatch.New(ctx, cfg.Worker.Address)
	if err != nil {
		log.Fatalf("failed to create dispatcher: %v", err)
	}
	defer dispatcher.Close()

	redisClient := redis.New(cfg.Redis.Addr, cfg.Redis.FeedTTL, cfg.Redis.AgentStatusTTL)
	defer redisClient.Close()

	sched := sched.New(cfg.Server.PollInterval, cfg.Server.PostInterval, cfg.Server.TurnTimeout, cfg.Server.MaxConcurrent, redisClient, db, dispatcher, hub, redisClient)
	sched.Start(ctx)
	defer sched.Stop()

	srv := httpapi.New(cfg.Server.Address, db, hub, redisClient)
	go func() {
		msg := fmt.Sprintf("HTTP Server started listening on %s", cfg.Server.Address)
		log.Println(msg)
		logs.Info("system", msg)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("http server stopped: %v", err)
			logs.Info("system", fmt.Sprintf("HTTP server stopped: %v", err))
		}
	}()

	<-ctx.Done()
	log.Println("shutting down backend service")
	logs.Info("system", "Shutting down backend orchestrator service")
}
