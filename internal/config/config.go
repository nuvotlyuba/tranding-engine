package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type HTTP struct {
	Host    string        `env:"HOST, default=127.0.0.1"`
	Port    int           `env:"PORT, default=8080"`
	Timeout TimeoutConfig `env:"TIMEOUT"`
}

type TimeoutConfig struct {
	ReadTimeout     time.Duration `env:"READ_TIMEOUT,default=5s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT,default=10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT,default=120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT,default=20s"`
}

type PostgresDB struct {
}

type Config struct {
	HTTP
}

func Load(ctx context.Context) (Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
