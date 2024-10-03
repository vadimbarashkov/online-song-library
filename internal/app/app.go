package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/httplog/v2"
	"github.com/vadimbarashkov/online-song-library/internal/adapter/api"
	"github.com/vadimbarashkov/online-song-library/internal/config"
	"github.com/vadimbarashkov/online-song-library/internal/usecase"
	"github.com/vadimbarashkov/online-song-library/pkg/postgres"
	"golang.org/x/sync/errgroup"

	delivery "github.com/vadimbarashkov/online-song-library/internal/adapter/delivery/http"
	repo "github.com/vadimbarashkov/online-song-library/internal/adapter/repository/postgres"
)

const migrationsPath = "file://migrations"

func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app.Run"

	db, err := postgres.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(migrationsPath, cfg.Postgres.DSN()); err != nil {
		return fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}

	songRepo := repo.NewSongRepository(db)
	musicInfoAPI := api.NewMusicInfoAPI("http://example.com", nil)
	songUseCase := usecase.NewSongUseCase(musicInfoAPI, songRepo)

	logger := setupLogger(cfg.Env)
	r := delivery.NewRouter(logger, songUseCase)

	server := &http.Server{
		Addr:           cfg.HTTPServer.Addr(),
		Handler:        r,
		ReadTimeout:    cfg.HTTPServer.ReadTimeout,
		WriteTimeout:   cfg.HTTPServer.WriteTimeout,
		IdleTimeout:    cfg.HTTPServer.IdleTimeout,
		MaxHeaderBytes: cfg.HTTPServer.MaxHeaderBytes,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error

		switch cfg.Env {
		case config.EnvProd:
			err = server.ListenAndServeTLS(cfg.HTTPServer.CertFile, cfg.HTTPServer.KeyFile)
		default:
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("%s: server error occurred: %w", op, err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("%s: failed to shutdown server: %w", op, err)
		}

		return nil
	})

	return g.Wait()
}

func setupLogger(env string) *httplog.Logger {
	opt := httplog.Options{
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		RequestHeaders:  true,
		ResponseHeaders: true,
	}

	switch env {
	case config.EnvTest:
		opt.JSON = true
	case config.EnvProd:
		opt.LogLevel = slog.LevelInfo
		opt.JSON = true
	}

	logger := httplog.NewLogger("online-song-library", opt)
	logger.Logger = logger.With(slog.String("env", env))

	return logger
}
