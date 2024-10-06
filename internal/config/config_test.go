package config

import (
	"os"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer_Addr(t *testing.T) {
	s := HTTPServer{
		Port: 8080,
	}

	assert.Equal(t, ":8080", s.Addr())
}

func TestPostgres_DSN(t *testing.T) {
	p := Postgres{
		User:     "test",
		Password: "test",
		Host:     "localhost",
		Port:     5432,
		DB:       "test",
		SSLMode:  "disable",
	}

	assert.Equal(t, "postgres://test:test@localhost:5432/test?sslmode=disable", p.DSN())
}

func TestLoad(t *testing.T) {
	t.Run("non-existent path", func(t *testing.T) {
		cfg, err := Load("/non-existent/path/.env")

		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, cfg)
	})

	t.Run("invalid environment variable", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		data := `ENV=test
HTTP_SERVER_PORT=not number
POSTGRES_USER=test
POSTGRES_PASSWORD=test
POSTGRES_DB=test
`

		f := createTempFile(t, ".env", []byte(data))
		cfg, err := Load(f.Name())

		assert.Error(t, err)
		assert.ErrorIs(t, err, env.ParseError{})
		assert.Nil(t, cfg)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			os.Clearenv()
		})

		data := `ENV=test
MUSIC_INFO_API=https://example.com.api
POSTGRES_USER=test
POSTGRES_PASSWORD=test
POSTGRES_DB=test
`

		f := createTempFile(t, ".env", []byte(data))
		cfg, err := Load(f.Name())

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test", cfg.Env)
		assert.Equal(t, "https://example.com.api", cfg.MusicInfoAPI)
		assert.Equal(t, "test", cfg.Postgres.User)
		assert.Equal(t, "test", cfg.Postgres.Password)
		assert.Equal(t, "test", cfg.Postgres.DB)
	})
}

func createTempFile(t testing.TB, name string, data []byte) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})

	if _, err := f.Write(data); err != nil {
		t.Fatalf("Failed to write data to temp file: %v", err)
	}

	return f
}
