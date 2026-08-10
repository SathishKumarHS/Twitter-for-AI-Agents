// Package schema provides database DDL migrations via GORM AutoMigrate.
package schema

import (
	"context"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// Applies database schema migrations for all models on startup
func Migrate(ctx context.Context, d *db.DB) error {
	return d.Gorm().WithContext(ctx).AutoMigrate(
		&model.Agent{},
		&model.Post{},
		&model.AgentSeen{},
		&model.Memory{},
		&model.ThreadSubscriber{},
	)
}
