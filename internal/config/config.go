package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type HTTP struct {
	Host    string        `env:"HOST,default=127.0.0.1"`
	Port    int           `env:"PORT,default=8080"`
	Timeout TimeoutConfig `env:"TIMEOUT"`
}

type TimeoutConfig struct {
	ReadTimeout     time.Duration `env:"READ_TIMEOUT,default=5s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT,default=10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT,default=120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT,default=20s"`
}

type PostgresDB struct {
	Host     string `env:"POSTGRES_HOST,default=localhost"`
	Port     int    `env:"POSTGRES_PORT,default=5432"`
	User     string `env:"POSTGRES_USER,required,default=engine"`
	Password string `env:"POSTGRES_PASSWORD,require,default=engine"`
	DBName   string `env:"POSTGRES_DB,required"`
	SSLMode  string `env:"POSTGRES_SSL_MODE,default=disable"`

	// pool настройки
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS,default=10"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS,default=2"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME,default=1h"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME,default=30m"`
}

type Config struct {
	HTTP
	Postgres PostgresDB
}

func Load(ctx context.Context) (Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
