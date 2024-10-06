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

// Run initializes and starts the application server.
// It accepts a context for cancellation and a configuration object containing
// application settings. The function performs the following tasks:
//
//  1. Connects to the PostgreSQL database using the provided Data Source Name (DSN).
//  2. Runs database migrations based on the provided migration path.
//  3. Initializes the song repository and the music information API client.
//  4. Sets up the song use case logic that interacts with the repository and API.
//  5. Configures the HTTP server with routing and timeout settings.
//  6. Starts the server in a separate goroutine, handling both TLS and non-TLS modes
//     depending on the environment configuration.
//  7. Waits for the context to be done (indicating shutdown) and gracefully shuts down
//     the server, ensuring all active connections are completed before exiting.
func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app.Run"

	logger := setupLogger(cfg.Env)

	logger.Info("connecting to the database")

	db, err := postgres.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}
	defer db.Close()

	logger.Info("running database migrations")

	if err := postgres.RunMigrations(cfg.MigrationsPath, cfg.Postgres.DSN()); err != nil {
		return fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}

	logger.Info("preparing server")

	songRepo := repo.NewSongRepository(db)
	musicInfoAPI := api.NewMusicInfoAPI(cfg.MusicInfoAPI, nil)
	songUseCase := usecase.NewSongUseCase(musicInfoAPI, songRepo)

	r := delivery.NewRouter(logger, songUseCase, &delivery.RouterOptions{
		SwaggerHost: cfg.HTTPServer.Host,
		SwaggerPort: cfg.HTTPServer.Port,
	})

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
		logger.Info("running server", slog.Any("addr", cfg.HTTPServer.Addr()))

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

		logger.Info("shutting down the server")

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("%s: failed to shutdown server: %w", op, err)
		}

		return nil
	})

	return g.Wait()
}

// setupLogger configures the HTTP logger based on the application environment.
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
