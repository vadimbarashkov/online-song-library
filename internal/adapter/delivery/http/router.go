package http

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
	"github.com/vadimbarashkov/online-song-library/pkg/validate"
)

// songUseCase defines the interface for the song use case layer.
// It includes methods for adding, fetching, modifying, and removing songs.
type songUseCase interface {
	AddSong(ctx context.Context, song entity.Song) (*entity.Song, error)
	FetchSongs(
		ctx context.Context,
		pagination entity.Pagination,
		filters ...entity.SongFilter,
	) ([]*entity.Song, *entity.Pagination, error)
	FetchSongWithVerses(
		ctx context.Context,
		songID uuid.UUID,
		pagination entity.Pagination,
	) (*entity.SongWithVerses, *entity.Pagination, error)
	ModifySong(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error)
	RemoveSong(ctx context.Context, songID uuid.UUID) (int64, error)
}

// NewRouter initializes a new HTTP router for the application.
// It sets up middleware for logging, CORS, and error handling, as well as route definitions.
func NewRouter(logger *httplog.Logger, songUseCase songUseCase) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"POST", "GET", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           84600,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", handlePing)

		r.Route("/songs", func(r chi.Router) {
			validate := newValidate()
			h := newSongHandler(songUseCase, validate)

			r.Post("/", h.addSong)
			r.Get("/", h.fetchSongs)

			r.Route("/{songID}", func(r chi.Router) {
				r.Get("/text", h.fetchSongWithVerses)
				r.Patch("/", h.modifySong)
				r.Delete("/", h.removeSong)
			})
		})
	})

	return r
}

// newValidate initializes a new validator for request validation.
// It registers custom validation rules and sets a tag name function for JSON field mapping.
func newValidate() *validator.Validate {
	v := validator.New()

	_ = v.RegisterValidation("releaseDate", validate.ReleaseDateValidation)

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}
