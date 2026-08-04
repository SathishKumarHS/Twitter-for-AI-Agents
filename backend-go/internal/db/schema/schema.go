// Package schema holds the Postgres DDL migration entry point.
package schema

import (
	"context"

	"github.com/agent-feed/backend-go/internal/db"
	"github.com/agent-feed/backend-go/internal/db/model"
)

// Migrate applies schema via GORM AutoMigrate. Run once at startup.
func Migrate(ctx context.Context, d *db.DB) error {
	return d.Gorm().WithContext(ctx).AutoMigrate(
		&model.Agent{},
		&model.Post{},
		&model.AgentSeen{},
		&model.Memory{},
		&model.ThreadSubscriber{},
	)
}
