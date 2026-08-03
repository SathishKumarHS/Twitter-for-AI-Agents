// Package db provides the Postgres connection (GORM) shared by the domain stores.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/agent-feed/backend-go/internal/db/agents"
	"github.com/agent-feed/backend-go/internal/db/memories"
	"github.com/agent-feed/backend-go/internal/db/posts"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps a GORM *gorm.DB connection.
type DB struct {
	gorm     *gorm.DB
	AgentsDB agents.Store
	PostDB   posts.Store
	MemoriesDB memories.Store
}

// Connect dials Postgres via GORM and returns a ready DB.
func Connect(ctx context.Context, dsn string) (*DB, error) {
	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
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
	return &DB{
		AgentsDB: agents.NewGorm(g),
		PostDB:   posts.NewGorm(g),
		MemoriesDB: memories.NewGorm(g),
	}, nil
}

func (d *DB) Close() {
	if sqlDB, err := d.gorm.DB(); err == nil {
		sqlDB.Close()
	}
}

// Gorm exposes the underlying GORM handle for AutoMigrate and queries.
func (d *DB) Gorm() *gorm.DB { return d.gorm }
