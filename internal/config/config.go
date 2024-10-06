package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Environment constants for application deployment.
const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

// Config holds the configuration settings for the application.
type Config struct {
	Env            string `env:"ENV" envDefault:"dev"`
	MigrationsPath string `env:"MIGRATIONS_PATH" envDefault:"migrations"`
	MusicInfoAPI   string `env:"MUSIC_INFO_API,required"`
	HTTPServer     `envPrefix:"HTTP_SERVER_"`
	Postgres       `envPrefix:"POSTGRES_"`
}

// HTTPServer contains settings related to the HTTP server.
type HTTPServer struct {
	Port           int           `env:"PORT" envDefault:"8080"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" envDefault:"1m"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" envDefault:"1048576"`
	CertFile       string        `env:"CERT_FILE"`
	KeyFile        string        `env:"KEY_FILE"`
}

// Addr returns the address <host:port> on which the HTTP server will listen.
func (s *HTTPServer) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

// Postgres contains settings required to connect to a PostgreSQL database.
type Postgres struct {
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	DB       string `env:"DB,required"`
	SSLMode  string `env:"SSLMODE" envDefault:"disable"`
}

// DSN returns the Data Source Name (DSN) used to connect to the PostgreSQL database.
func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode)
}

// Load reads environment variables from a specified .env file and parses them into the Config struct.
func Load(path string) (*Config, error) {
	const op = "config.Load"

	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("%s: failed to load env vars from env file: %w", op, err)
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("%s: failed to parse Config struct: %w", op, err)
	}

	return &cfg, nil
}
