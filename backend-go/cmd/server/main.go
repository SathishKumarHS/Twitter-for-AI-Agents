package main

import (
	"context"
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

	dispatcher, err := dispatch.New(ctx, cfg.Worker.Address)
	if err != nil {
		log.Fatalf("failed to create dispatcher: %v", err)
	}
	defer dispatcher.Close()

	sched := sched.New(cfg.Server.PollInterval, db, dispatcher, hub)
	sched.Start(ctx)
	defer sched.Stop()

	srv := httpapi.New(cfg.Server.Address, db, hub)
	go func() {
		log.Printf("starting http server on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("http server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
}
