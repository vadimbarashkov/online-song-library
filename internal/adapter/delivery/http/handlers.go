package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

func handlePing(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

// songHandler struct handles HTTP requests related to songs.
type songHandler struct {
	songUseCase songUseCase
	validate    *validator.Validate
}

// newSongHandler initializes a new songHandler instance.
func newSongHandler(songUseCase songUseCase, validate *validator.Validate) *songHandler {
	return &songHandler{
		songUseCase: songUseCase,
		validate:    validate,
	}
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

func (h *songHandler) addSong(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	logger.Debug("handling add song request")

	var req addSongRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			logger.Debug("empty request body")

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

	song, err := h.songUseCase.AddSong(r.Context(), h.addSongRequestToEntity(req))
	if err != nil {
		httplog.LogEntrySetField(r.Context(), "err", slog.AnyValue(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("Song added successfully", slog.Any("songID", song.ID))

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, h.entityToSongSchema(song))
}

func (h *songHandler) fetchSongs(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	logger.Info("handling fetch songs request")

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

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("Songs fetched successfully", slog.Any("items", pgn.Items))

	resp := songsResponse{Pagination: h.entityToPaginationSchema(pgn)}
	for _, song := range songs {
		resp.Songs = append(resp.Songs, h.entityToSongSchema(song))
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

func (h *songHandler) fetchSongWithVerses(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	logger.Info("handling fetch song with verses request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug("invalid song ID", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	logger.Debug("fetching song with verses", slog.Any("songID", songID))

	pagination := parsePagination(r)

	song, pgn, err := h.songUseCase.FetchSongWithVerses(r.Context(), songID, pagination)
	if err != nil {
		if errors.Is(err, entity.ErrSongNotFound) {
			logger.Debug("song not found", slog.Any("songID", songID))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		httplog.LogEntrySetFields(r.Context(), map[string]any{
			"songID": songID,
			"err":    err,
		})

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("song with verses fetched successfully", slog.Any("songID", song.ID))

	resp := songWithVersesResponse{
		Song:       h.entityToSongWithVersesSchema(song),
		Pagination: h.entityToPaginationSchema(pgn),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

func (h *songHandler) modifySong(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	logger.Debug("handling modify song request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug("invalid song ID", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	var req updateSongRequest

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		if errors.Is(err, io.EOF) {
			logger.Debug("empty request body")

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
		if errors.Is(err, entity.ErrSongNotFound) {
			httplog.LogEntrySetField(r.Context(), "songID", slog.AnyValue(songID))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		httplog.LogEntrySetFields(r.Context(), map[string]any{
			"songID": songID,
			"err":    err,
		})

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	logger.Debug("song modified successfully", slog.Any("songID", songID))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, h.entityToSongSchema(song))
}

func (h *songHandler) removeSong(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	logger.Info("handling remove song request")

	songIDParam := chi.URLParam(r, "songID")

	songID, err := uuid.Parse(songIDParam)
	if err != nil {
		logger.Debug("invalid song ID", slog.Any("err", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, invalidSongIDParamResp)
		return
	}

	logger.Debug("removing song", slog.Any("songID", songID))

	removed, err := h.songUseCase.RemoveSong(r.Context(), songID)
	if err != nil {
		if errors.Is(err, entity.ErrSongNotFound) {
			httplog.LogEntrySetField(r.Context(), "songID", slog.AnyValue(songID))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, songNotFoundErrResp)
			return
		}

		httplog.LogEntrySetFields(r.Context(), map[string]any{
			"songID": songID,
			"err":    err,
		})

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, serverErrResp)
		return
	}

	if removed > 1 {
		logger.Error("removed more than one object", slog.Int64("removed", removed))
	}

	w.WriteHeader(http.StatusNoContent)
}
