// Package config defines backend configuration structures and loads environment configuration files.
package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Holds database connection configuration settings
type DBConfig struct {
	URL string `mapstructure:"url"`
}

// Holds AI worker gRPC client address configuration
type WorkerConfig struct {
	Address string `mapstructure:"address"`
}

// Holds orchestrator HTTP server, gRPC tool server, and scheduler timing configurations
type ServerConfig struct {
	Address          string        `mapstructure:"address"`
	PollInterval     time.Duration `mapstructure:"poll-interval"`
	PostInterval     time.Duration `mapstructure:"post-interval"`
	MaxConcurrent    int           `mapstructure:"max-concurrent"`
	TurnTimeout      time.Duration `mapstructure:"turn-timeout"`
	ToolsGRPCAddress string        `mapstructure:"tools-grpc-address"`
}

// Holds Redis cache host address and key TTL settings
type RedisConfig struct {
	Addr           string        `mapstructure:"addr"`
	FeedTTL        time.Duration `mapstructure:"feed-ttl"`
	AgentStatusTTL time.Duration `mapstructure:"agent-status-ttl"`
}

// Holds default and maximum limits for tool executions
type LimitsConfig struct {
	FeedToolDefault int `mapstructure:"feed-tool-default"`
	FeedToolMax     int `mapstructure:"feed-tool-max"`
}

// Top-level backend application configuration container
type Config struct {
	DB     DBConfig     `mapstructure:"db"`
	Worker WorkerConfig `mapstructure:"ai-worker"`
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Limits LimitsConfig `mapstructure:"limits"`
}

// Loads the active environment YAML configuration file (e.g. config/local/config.yaml) using Viper
func LoadConfig() Config {
	env := os.Getenv("ACTIVE_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/" + env)
	viper.AddConfigPath("backend-go/config/" + env)
	viper.AddConfigPath("../config/" + env)
	viper.AddConfigPath("../../config/" + env)

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Error reading config file", "env", env, "error", err)
		os.Exit(1)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("Unable to decode config into struct", "error", err)
		os.Exit(1)
	}

	return cfg
}
