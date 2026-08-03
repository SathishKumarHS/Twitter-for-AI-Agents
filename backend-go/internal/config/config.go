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
	Address      string        `mapstructure:"address"`
	PollInterval time.Duration `mapstructure:"poll-interval"`
}

type Config struct {
	DB     DBConfig     `mapstructure:"db"`
	Worker WorkerConfig `mapstructure:"ai-worker"`
	Server ServerConfig `mapstructure:"server"`
}

const defaultPollInterval = 30 * time.Second

func LoadConfig() Config {
	env := os.Getenv("ACTIVE_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/" + env)

	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.poll-interval", defaultPollInterval)
	viper.SetDefault("ai-worker.address", "ai-worker:50051")
	viper.SetDefault("db.url", "postgres://agent:agent@db:5432/agents?sslmode=disable")

	viper.BindEnv("server.address", "AGENT_FEED_ADDR")
	viper.BindEnv("server.poll-interval", "AGENT_FEED_POLL")
	viper.BindEnv("ai-worker.address", "WORKER_ADDR")
	viper.BindEnv("db.url", "DATABASE_URL")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Error reading config file: %v", err)
		} else {
			log.Printf("Config file not found for env '%s', falling back to defaults and env vars", env)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	return cfg
}
