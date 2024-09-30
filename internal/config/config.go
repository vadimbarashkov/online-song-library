package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

type Config struct {
	Env      string `env:"ENV" envDefault:"dev"`
	Postgres `envPrefix:"POSTGRES_"`
}

type Postgres struct {
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	DB       string `env:"DB,required"`
	SSLMode  string `env:"SSLMODE" envDefault:"disable"`
}

func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}

func Load(path ...string) (*Config, error) {
	const op = "config.Load"

	var err error

	if len(path) > 0 {
		err = godotenv.Load(path[0])
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		return nil, fmt.Errorf("%s: failed to load env vars from env file: %w", op, err)
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse Config struct: %w", op, err)
	}

	return &cfg, nil
}
