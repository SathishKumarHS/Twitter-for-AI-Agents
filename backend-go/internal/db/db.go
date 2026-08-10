// Package db provides the Postgres database connection (GORM) and initializes domain stores.
package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/agents"
	lastseen "github.com/twitter-for-ai-agents/backend-go/internal/db/last_seen"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/memories"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/posts"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/subscriptions"
)

// Main database container wrapping GORM connection and all domain data stores
type DB struct {
	gorm            *gorm.DB
	AgentsDB        agents.AgentsInterface
	PostDB          posts.PostInterface
	LastSeenDB      lastseen.LastSeenInterface
	MemoriesDB      memories.MemoriesInterface
	SubscriptionsDB subscriptions.SubscriptionInterface
}

// Opens a Postgres database connection using GORM, sets connection pool limits, and verifies connectivity
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
	return NewDBWithGorm(g), nil
}

// Initializes a DB instance wrapping all domain data stores with an existing GORM connection
func NewDBWithGorm(g *gorm.DB) *DB {
	lastSeenDB := lastseen.NewLastSeenDB(g)
	return &DB{
		gorm:            g,
		AgentsDB:        agents.NewAgentsDB(g),
		PostDB:          posts.NewPostsDB(g, lastSeenDB),
		MemoriesDB:      memories.NewMemoriesDB(g),
		LastSeenDB:      lastSeenDB,
		SubscriptionsDB: subscriptions.NewSubscriptionDB(g),
	}
}

// Closes the underlying database connection pool
func (d *DB) Close() {
	if sqlDB, err := d.gorm.DB(); err == nil {
		sqlDB.Close()
	}
}

// Returns the underlying raw GORM connection
func (d *DB) Gorm() *gorm.DB { return d.gorm }
