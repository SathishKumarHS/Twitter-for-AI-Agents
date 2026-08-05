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

	if cfg.Server.MaxConcurrent != 5 {
		t.Errorf("expected max-concurrent 5, got %d", cfg.Server.MaxConcurrent)
	}

	if cfg.Server.PollInterval != 5*time.Minute {
		t.Errorf("expected poll-interval 5m, got %v", cfg.Server.PollInterval)
	}

	if cfg.Server.PostInterval != 30*time.Minute {
		t.Errorf("expected post-interval 30m, got %v", cfg.Server.PostInterval)
	}

	if cfg.Server.TurnTimeout != 3*time.Minute {
		t.Errorf("expected turn-timeout 3m, got %v", cfg.Server.TurnTimeout)
	}

	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("expected redis addr 'redis:6379', got '%s'", cfg.Redis.Addr)
	}

	if cfg.Limits.LogRingBuffer != 500 {
		t.Errorf("expected log-ring-buffer 500, got %d", cfg.Limits.LogRingBuffer)
	}
}
