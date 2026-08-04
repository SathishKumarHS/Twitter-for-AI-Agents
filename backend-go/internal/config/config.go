package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type DBConfig struct {
	URL string `mapstructure:"url"`
}

type WorkerConfig struct {
	Address string `mapstructure:"address"`
}

type ServerConfig struct {
	Address       string        `mapstructure:"address"`
	PollInterval  time.Duration `mapstructure:"poll-interval"`
	PostInterval  time.Duration `mapstructure:"post-interval"`
	MaxConcurrent int           `mapstructure:"max-concurrent"`
	TurnTimeout   time.Duration `mapstructure:"turn-timeout"`
}

type RedisConfig struct {
	Addr           string        `mapstructure:"addr"`
	FeedTTL        time.Duration `mapstructure:"feed-ttl"`
	AgentStatusTTL time.Duration `mapstructure:"agent-status-ttl"`
}

type LimitsConfig struct {
	FeedContext        int `mapstructure:"feed-context"`
	Memories           int `mapstructure:"memories"`
	LogRingBuffer      int `mapstructure:"log-ring-buffer"`
	FeedAPIDefault     int `mapstructure:"feed-api-default"`
	LogsAPIDefault     int `mapstructure:"logs-api-default"`
	MemoriesAPIDefault int `mapstructure:"memories-api-default"`
}

type Config struct {
	DB     DBConfig     `mapstructure:"db"`
	Worker WorkerConfig `mapstructure:"ai-worker"`
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Limits LimitsConfig `mapstructure:"limits"`
}

func LoadConfig() Config {
	env := os.Getenv("ACTIVE_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/" + env)
	viper.AddConfigPath("backend-go/config/" + env)
	viper.AddConfigPath("../backend-go/config/" + env)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file for env '%s': %v", env, err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return cfg
}
