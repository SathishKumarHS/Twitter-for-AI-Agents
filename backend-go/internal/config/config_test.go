package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigLocal(t *testing.T) {
	_ = os.Setenv("ACTIVE_ENV", "local")

	cfg := LoadConfig()

	if cfg.DB.URL == "" {
		t.Error("expected non-empty DB URL")
	}

	if cfg.Worker.Address != "ai-worker:50051" {
		t.Errorf("expected worker address 'ai-worker:50051', got '%s'", cfg.Worker.Address)
	}

	if cfg.Server.Address != ":8080" {
		t.Errorf("expected server address ':8080', got '%s'", cfg.Server.Address)
	}

	if cfg.Server.MaxConcurrent != 4 {
		t.Errorf("expected max-concurrent 4, got %d", cfg.Server.MaxConcurrent)
	}

	if cfg.Server.ToolsGRPCAddress != ":50052" {
		t.Errorf("expected tools-grpc-address ':50052', got '%s'", cfg.Server.ToolsGRPCAddress)
	}

	if cfg.Server.PollInterval != 1*time.Minute {
		t.Errorf("expected poll-interval 1m, got %v", cfg.Server.PollInterval)
	}

	if cfg.Server.PostInterval != 5*time.Minute {
		t.Errorf("expected post-interval 5m, got %v", cfg.Server.PostInterval)
	}

	if cfg.Server.TurnTimeout != 5*time.Minute {
		t.Errorf("expected turn-timeout 5m, got %v", cfg.Server.TurnTimeout)
	}

	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("expected redis addr 'redis:6379', got '%s'", cfg.Redis.Addr)
	}

	if cfg.Limits.FeedToolDefault != 10 {
		t.Errorf("expected feed-tool-default 10, got %d", cfg.Limits.FeedToolDefault)
	}

	if cfg.Limits.FeedToolMax != 20 {
		t.Errorf("expected feed-tool-max 20, got %d", cfg.Limits.FeedToolMax)
	}
}
