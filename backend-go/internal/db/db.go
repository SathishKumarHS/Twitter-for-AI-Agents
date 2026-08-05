// Package db provides the Postgres connection (GORM) shared by the domain stores.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/agents"
	lastseen "github.com/twitter-for-ai-agents/backend-go/internal/db/last_seen"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/memories"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/posts"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/subscriptions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	gorm            *gorm.DB
	AgentsDB        agents.AgentsInterface
	PostDB          posts.PostInterface
	LastSeenDB      lastseen.LastSeenInterface
	MemoriesDB      memories.MemoriesInterface
	SubscriptionsDB subscriptions.SubscriptionInterface
}

func Connect(ctx context.Context, dsn string) (*DB, error) {
	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := g.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	lastSeenDB := lastseen.NewLastSeenDB(g)
	return &DB{
		gorm:            g,
		AgentsDB:        agents.NewAgentsDB(g),
		PostDB:          posts.NewPostsDB(g, lastSeenDB),
		MemoriesDB:      memories.NewMemoriesDB(g),
		LastSeenDB:      lastSeenDB,
		SubscriptionsDB: subscriptions.NewSubscriptionDB(g),
	}, nil
}

func (d *DB) Close() {
	if sqlDB, err := d.gorm.DB(); err == nil {
		sqlDB.Close()
	}
}

func (d *DB) Gorm() *gorm.DB { return d.gorm }
