//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm/logger"

	"github.com/twitter-for-ai-agents/backend-go/internal/db"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/schema"
	appredis "github.com/twitter-for-ai-agents/backend-go/internal/redis"
)

var (
	baseURL   = "http://localhost:8080"
	dsn       = "postgres://agent:agent@localhost:5432/agents?sslmode=disable"
	redisAddr = "localhost:6379"

	database *db.DB
	rc       appredis.RedisInterface
	raw      *goredis.Client
)

func TestMain(m *testing.M) {
	if v := os.Getenv("E2E_BASE_URL"); v != "" {
		baseURL = v
	}
	if v := os.Getenv("E2E_DB_DSN"); v != "" {
		dsn = v
	}
	if v := os.Getenv("E2E_REDIS_ADDR"); v != "" {
		redisAddr = v
	}

	ctx := context.Background()
	waitForHealth(ctx)

	var err error
	database, err = db.Connect(ctx, dsn)
	if err != nil {
		panic("integration: connect db: " + err.Error())
	}
	database.Gorm().Logger = logger.Default.LogMode(logger.Silent)
	if err := schema.Migrate(ctx, database); err != nil {
		panic("integration: migrate: " + err.Error())
	}

	rc = appredis.New(redisAddr, 30*time.Second, 24*time.Hour)
	raw = goredis.NewClient(&goredis.Options{Addr: redisAddr})
	if err := raw.Ping(ctx).Err(); err != nil {
		panic("integration: connect redis: " + err.Error())
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	reset(ctx)

	code := m.Run()

	reset(ctx)
	database.Close()
	raw.Close()
	os.Exit(code)
}

func waitForHealth(ctx context.Context) {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		if resp, err := client.Get(baseURL + "/healthz"); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	fmt.Fprintln(os.Stderr, "integration: API not healthy in time — is `make integration-up` running?")
	os.Exit(1)
}

func reset(ctx context.Context) {
	if err := database.Gorm().Exec(
		"TRUNCATE thread_subscribers, memories, agent_seen, posts, agents RESTART IDENTITY CASCADE",
	).Error; err != nil {
		panic("integration: truncate: " + err.Error())
	}
	for _, pat := range []string{"feed:recent:*", "agent:active:*", "agent:turn:*"} {
		iter := raw.Scan(ctx, 0, pat, 0).Iterator()
		for iter.Next(ctx) {
			raw.Del(ctx, iter.Val())
		}
	}
}

func waitFor(t *testing.T, timeout time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", what)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
