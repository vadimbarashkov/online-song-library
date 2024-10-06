package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

// handlePing handles the ping request.
//
//	@Summary		Server healthcehck
//	@Description	Responds with "pong" to verify the server is running.
//	@Tags			healthcheck
//	@Produce		plain
//	@Success		200	"Server is running"
//	@Router			/api/v1/ping [get]
func handlePing(logger *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		logger.Debug("handling ping request", slog.String("reqID", reqID))

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	})
}

// songHandler struct handles HTTP requests related to songs.
type songHandler struct {
	logger      *slog.Logger
	songUseCase songUseCase
	validate    *validator.Validate
}

// newSongHandler initializes a new songHandler instance.
func newSongHandler(logger *slog.Logger, songUseCase songUseCase, validate *validator.Validate) *songHandler {
	return &songHandler{
		logger:      logger,
		songUseCase: songUseCase,
		validate:    validate,
	}
}

// prepareLogger enhances the logger with the request ID.
func (h *songHandler) prepareLogger(ctx context.Context) *slog.Logger {
	reqID := middleware.GetReqID(ctx)
	return h.logger.With(slog.String("reqID", reqID))
}

// addSongRequestToEntity converts an addSongRequest to an entity.Song.
func (h *songHandler) addSongRequestToEntity(req addSongRequest) entity.Song {
	return entity.Song{
		GroupName: req.Group,
		Name:      req.Song,
	}
}

// updateSongRequestToEntity converts an updateSongRequest to an entity.Song.
func (h *songHandler) updateSongRequestToEntity(req updateSongRequest) entity.Song {
	releaseDate, _ := time.Parse("02.01.2006", req.ReleaseDate)

	return entity.Song{
		GroupName: req.GroupName,
		Name:      req.Name,
		SongDetail: entity.SongDetail{
			ReleaseDate: releaseDate,
			Text:        req.Text,
			Link:        req.Link,
		},
	}
}

// entityToSongSchema converts an entity.Song to songSchema for response.
func (h *songHandler) entityToSongSchema(song *entity.Song) songSchema {
	return songSchema{
		ID:        song.ID,
		GroupName: song.GroupName,
		Name:      song.Name,
		SongDetail: songDetailSchema{
			ReleaseDate: song.SongDetail.ReleaseDate.Format("02.01.2006"),
			Text:        song.SongDetail.Text,
			Link:        song.SongDetail.Link,
		},
		CreatedAt: song.CreatedAt,
		UpdatedAt: song.UpdatedAt,
	}
}

// entityToSongWithVersesSchema converts an entity.SongWithVerses to songWithVersesSchema for response.
func (h *songHandler) entityToSongWithVersesSchema(song *entity.SongWithVerses) songWithVersesSchema {
	return songWithVersesSchema{
		ID:        song.ID,
		GroupName: song.GroupName,
		Name:      song.Name,
		Verses:    song.Verses,
		CreatedAt: song.CreatedAt,
		UpdatedAt: song.UpdatedAt,
	}
}

// entityToPaginationSchema converts entity.Pagination to paginationSchema for response.
func (h *songHandler) entityToPaginationSchema(pagination *entity.Pagination) paginationSchema {
	return paginationSchema{
		Offset: pagination.Offset,
		Limit:  pagination.Limit,
		Items:  pagination.Items,
		Total:  pagination.Total,
	}
}

// addSong handles adding a new song to the library.
//
//	@Summary		Add a new song
//	@Description	Adds a new song to the library
//	@Tags			songs
//	@Accept			json
//	@Produce		json
//	@Param			song	body		addSongRequest	true	"Add Song"
//	@Success		201		{object}	songSchema
//	@Failure		400		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/songs [post]
func (h *songHandler) addSong(w http.ResponseWriter, r *http.Request) {
	logger := h.prepareLogger(r.Context())
	logger.Debug("handling add song request")

	var req addSongRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			logger.Debug("empty request body", slog.Any("err", err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, emptyRequestBodyResp)
			return
		}

		logger.Debug("invalid request body", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidRequestBodyResp)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		logger.Debug("validation error", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, validationError(err))
		return
	}

	logger.Debug("adding song")

	song, err := h.songUseCase.AddSong(r.Context(), h.addSongRequestToEntity(req))
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		logger.Debug("failed to add song", slog.Any("err", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("song added successfully", slog.Any("songID", song.ID))

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, h.entityToSongSchema(song))
}

// fetchSongs handles fetching multiple songs with optional filters and pagination.
//
//	@Summary		Fetch multiple songs
//	@Description	Retrieves a list of songs from the library
//	@Tags			songs
//	@Accept			json
//	@Produce		json
//	@Param			limit				query		int		false	"Limit the number of items"
//	@Param			offset				query		int		false	"Offset for pagination"
//	@Param			groupName			query		string	false	"Filter by group name"
//	@Param			name				query		string	false	"Filter by song name"
//	@Param			releaseYear			query		string	false	"Filter by release year"
//	@Param			releaseDate			query		string	false	"Filter by exact release date (dd.MM.yyyy)"
//	@Param			releaseDateAfter	query		string	false	"Filter songs released after the specified date (dd.MM.yyyy)"
//	@Param			releaseDateBefore	query		string	false	"Filter songs released before the specified date (dd.MM.yyyy)"
//	@Param			text				query		string	false	"Filter by song text"
//	@Success		200					{object}	songsResponse
//	@Failure		500					{object}	errorResponse
//	@Router			/api/v1/songs [get]
func (h *songHandler) fetchSongs(w http.ResponseWriter, r *http.Request) {
	logger := h.prepareLogger(r.Context())
	logger.Debug("handling fetch songs request")

	pagination := parsePagination(r)
	filters := parseSongFilters(r)

	logger.Debug(
		"fetching songs",
		slog.Any("pagination", pagination),
		slog.Any("filters", filters),
	)

	songs, pgn, err := h.songUseCase.FetchSongs(r.Context(), pagination, filters...)
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		logger.Debug("failed to fetch songs", slog.Any("err", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("songs fetched successfully", slog.Uint64("items", pgn.Items))

	resp := songsResponse{
		Songs:      make([]songSchema, 0),
		Pagination: h.entityToPaginationSchema(pgn),
	}
	for _, song := range songs {
		resp.Songs = append(resp.Songs, h.entityToSongSchema(song))
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// fetchSongWithVerses handles fetching a song along with its verses by song ID.
//
//	@Summary		Fetch a song with verses
//	@Description	Retrieves a song along with its verses using the song ID
//	@Tags			songs
//	@Accept			json
//	@Produce		json
//	@Param			songID	path		string	true	"Song ID"
//	@Param			limit	query		int		false	"Limit the number of verses"
//	@Param			offset	query		int		false	"Offset for pagination"
//	@Success		200		{object}	songWithVersesResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/songs/{songID}/text [get]
func (h *songHandler) fetchSongWithVerses(w http.ResponseWriter, r *http.Request) {
	logger := h.prepareLogger(r.Context())
	logger.Debug("handling fetch song with verses request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug(
			"invalid song ID",
			slog.String("songID", songIDParam),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	pagination := parsePagination(r)

	logger.Debug(
		"fetching song with verses",
		slog.Any("songID", songID),
		slog.Any("pagination", pagination),
	)

	song, pgn, err := h.songUseCase.FetchSongWithVerses(r.Context(), songID, pagination)
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		if errors.Is(err, entity.ErrSongNotFound) {
			logger.Debug(
				"song not found",
				slog.Any("songID", songID),
				slog.Any("err", err),
			)

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		logger.Debug(
			"failed to fetch song with verses",
			slog.Any("songID", songID),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("song with verses fetched successfully")

	resp := songWithVersesResponse{
		Song:       h.entityToSongWithVersesSchema(song),
		Pagination: h.entityToPaginationSchema(pgn),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// modifySong handles modifying a song's details using its unique ID.
//
//	@Summary		Modify a song
//	@Description	Updates a song's information using the song ID
//	@Tags			songs
//	@Accept			json
//	@Produce		json
//	@Param			songID	path		string				true	"Song ID"
//	@Param			song	body		updateSongRequest	true	"Update Song"
//	@Success		200		{object}	songSchema
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/songs/{songID} [patch]
func (h *songHandler) modifySong(w http.ResponseWriter, r *http.Request) {
	logger := h.prepareLogger(r.Context())
	logger.Debug("handling modify song request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug(
			"invalid song ID",
			slog.String("songID", songIDParam),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	var req updateSongRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			logger.Debug("empty request body", slog.Any("err", err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, emptyRequestBodyResp)
			return
		}

		logger.Debug("invalid request body", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidRequestBodyResp)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		logger.Debug("validation error", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, validationError(err))
		return
	}

	logger.Debug("song modification", slog.Any("songID", songID))

	song, err := h.songUseCase.ModifySong(r.Context(), songID, h.updateSongRequestToEntity(req))
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		if errors.Is(err, entity.ErrSongNotFound) {
			logger.Debug(
				"song not found",
				slog.Any("songID", songID),
				slog.Any("err", err),
			)

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		logger.Debug(
			"failed to modify song",
			slog.Any("songID", songID),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("song modified successfully", slog.Any("songID", song.ID))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, h.entityToSongSchema(song))
}

// removeSong handles deleting a song by its unique ID.
//
//	@Summary		Remove a song
//	@Description	Deletes a song using the song ID
//	@Tags			songs
//	@Param			songID	path	string	true	"Song ID"
//	@Success		204		"Song deleted successfully"
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/songs/{songID} [delete]
func (h *songHandler) removeSong(w http.ResponseWriter, r *http.Request) {
	logger := h.prepareLogger(r.Context())
	logger.Debug("handling remove song request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug(
			"invalid song ID",
			slog.String("songID", songIDParam),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	logger.Debug("removing song", slog.Any("songID", songID))

	removed, err := h.songUseCase.RemoveSong(r.Context(), songID)
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		if errors.Is(err, entity.ErrSongNotFound) {
			logger.Debug(
				"song not found",
				slog.Any("songID", songID),
				slog.Any("err", err),
			)

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		logger.Debug(
			"failed to modify song",
			slog.Any("songID", songID),
			slog.Any("err", err),
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	if removed > 1 {
		logger.Error("removed more than one object", slog.Int64("removed", removed))
	}

	w.WriteHeader(http.StatusNoContent)
}
