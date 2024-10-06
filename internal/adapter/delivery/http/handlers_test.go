package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/httplog/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/vadimbarashkov/online-song-library/internal/entity"

	httpMock "github.com/vadimbarashkov/online-song-library/mocks/http"
)

var (
	fixedUUID = uuid.New()
	fixedTime = time.Now()
)

func setupServer(t testing.TB) (*httpexpect.Expect, *httpMock.MockSongUseCase) {
	t.Helper()

	logger := httplog.NewLogger("", httplog.Options{Writer: io.Discard})
	songUseCaseMock := httpMock.NewMockSongUseCase(t)
	r := NewRouter(logger, songUseCaseMock, nil)

	server := httptest.NewServer(r)
	t.Cleanup(func() {
		server.Close()
	})

	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: nil,
	}), songUseCaseMock
}

func TestPing(t *testing.T) {
	const path = "/api/v1/ping"

	t.Run("success", func(t *testing.T) {
		e, _ := setupServer(t)

		e.GET(path).
			Expect().
			Status(http.StatusOK).
			Body().IsEqual("pong")
	})
}

func TestSongHandler_AddSong(t *testing.T) {
	const path = "/api/v1/songs"

	t.Run("empty request body", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.POST(path).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", emptyRequestBodyResp.Message)
	})

	t.Run("invalid request body", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.POST(path).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", invalidRequestBodyResp.Message)
	})

	t.Run("validation error", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.POST(path).
			WithJSON(map[string]any{
				"group": "Test Group",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.ContainsKey("message")
		resp.Value("details").Array().Length().IsEqual(1)
	})

	t.Run("server error", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("AddSong", mock.Anything, mock.Anything).
			Once().
			Return(nil, errors.New("unknown error"))

		resp := e.POST(path).
			WithJSON(map[string]any{
				"group": "Test Group",
				"song":  "Test Song",
			}).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", serverErrResp.Message)
	})

	t.Run("success", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("AddSong", mock.Anything, entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
			}).
			Once().
			Return(&entity.Song{
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		resp := e.POST(path).
			WithJSON(map[string]any{
				"group": "Test Group",
				"song":  "Test Song",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.HasValue("id", fixedUUID)
		resp.HasValue("groupName", "Test Group")
		resp.HasValue("name", "Test Song")
		resp.Value("songDetail").Object().
			HasValue("releaseDate", fixedTime.Format("02.01.2006")).
			HasValue("text", "Test Text").
			HasValue("link", "https://example.com")
		resp.HasValue("created_at", fixedTime)
		resp.HasValue("updated_at", fixedTime)
	})
}

func TestSongHandler_FetchSongs(t *testing.T) {
	const path = "/api/v1/songs"

	t.Run("server error", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("FetchSongs", mock.Anything, mock.Anything).
			Once().
			Return(nil, nil, errors.New("unknown error"))

		resp := e.GET(path).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", serverErrResp.Message)
	})

	t.Run("success", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("FetchSongs", mock.Anything, mock.Anything).
			Once().
			Return([]*entity.Song{
				{
					ID:        fixedUUID,
					GroupName: "Test Group",
					Name:      "Test Song",
					SongDetail: entity.SongDetail{
						ReleaseDate: fixedTime,
						Text:        "Test Text",
						Link:        "https://example.com",
					},
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			}, &entity.Pagination{
				Offset: entity.DefaultOffset,
				Limit:  entity.DefaultLimit,
				Items:  1,
				Total:  1,
			}, nil)

		resp := e.GET(path).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		songs := resp.Value("songs").Array()
		song := songs.Value(0).Object()

		songs.Length().IsEqual(1)
		song.HasValue("id", fixedUUID)
		song.HasValue("groupName", "Test Group")
		song.HasValue("name", "Test Song")
		song.Value("songDetail").Object().
			HasValue("releaseDate", fixedTime.Format("02.01.2006")).
			HasValue("text", "Test Text").
			HasValue("link", "https://example.com")
		song.HasValue("created_at", fixedTime)
		song.HasValue("updated_at", fixedTime)

		resp.Value("pagination").Object().
			HasValue("offset", entity.DefaultOffset).
			HasValue("limit", entity.DefaultLimit).
			HasValue("items", 1).
			HasValue("total", 1)
	})
}

func TestSongHandler_FetchSongWithVerses(t *testing.T) {
	const path = "/api/v1/songs/{songID}/text"

	t.Run("invalid song id", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.GET(path, "invalid uuid").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", invalidSongIDParamResp.Message)
	})

	t.Run("song not found", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("FetchSongWithVerses", mock.Anything, fixedUUID, mock.Anything).
			Once().
			Return(nil, nil, entity.ErrSongNotFound)

		resp := e.GET(path, fixedUUID).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", songNotFoundErrResp.Message)
	})

	t.Run("server error", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("FetchSongWithVerses", mock.Anything, fixedUUID, mock.Anything).
			Once().
			Return(nil, nil, errors.New("unknown error"))

		resp := e.GET(path, fixedUUID).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", serverErrResp.Message)
	})

	t.Run("success", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("FetchSongWithVerses", mock.Anything, fixedUUID, mock.Anything).
			Once().
			Return(&entity.SongWithVerses{
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Name",
				Verses:    []string{"Line1\nLine2\n", "Line3\nLine4\n"},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, &entity.Pagination{
				Offset: entity.DefaultOffset,
				Limit:  entity.DefaultLimit,
				Items:  2,
				Total:  2,
			}, nil)

		resp := e.GET(path, fixedUUID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		song := resp.Value("song").Object()

		song.HasValue("id", fixedUUID)
		song.HasValue("groupName", "Test Group")
		song.HasValue("name", "Test Name")
		song.Value("verses").Array().Length().IsEqual(2)
		song.HasValue("created_at", fixedTime)
		song.HasValue("updated_at", fixedTime)

		resp.Value("pagination").Object().
			HasValue("offset", entity.DefaultOffset).
			HasValue("limit", entity.DefaultLimit).
			HasValue("items", 2).
			HasValue("total", 2)
	})
}

func TestSongHandler_ModifySong(t *testing.T) {
	const path = "/api/v1/songs/{songID}"

	t.Run("invalid song id", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.PATCH(path, "invalid uuid").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", invalidSongIDParamResp.Message)
	})

	t.Run("empty request body", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.PATCH(path, fixedUUID).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", emptyRequestBodyResp.Message)
	})

	t.Run("invalid request body", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.PATCH(path, fixedUUID).
			WithJSON("invalid body").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", invalidRequestBodyResp.Message)
	})

	t.Run("validation error", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.PATCH(path, fixedUUID).
			WithJSON(map[string]any{
				"link": "invalid url",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.ContainsKey("message")
		resp.Value("details").Array().Length().IsEqual(1)
	})

	t.Run("song not found", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("ModifySong", mock.Anything, fixedUUID, mock.Anything).
			Once().
			Return(nil, entity.ErrSongNotFound)

		resp := e.PATCH(path, fixedUUID).
			WithJSON(map[string]any{
				"text": "New Test Text",
				"link": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", songNotFoundErrResp.Message)
	})

	t.Run("server error", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("ModifySong", mock.Anything, fixedUUID, mock.Anything).
			Once().
			Return(nil, errors.New("unknown error"))

		resp := e.PATCH(path, fixedUUID).
			WithJSON(map[string]any{
				"text": "New Test Text",
				"link": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", serverErrResp.Message)
	})

	t.Run("success", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("ModifySong", mock.Anything, fixedUUID, entity.Song{
				SongDetail: entity.SongDetail{
					Text: "New Test Text",
					Link: "https://new-example.com",
				},
			}).
			Once().
			Return(&entity.Song{
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "New Test Text",
					Link:        "https://new-example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		resp := e.PATCH(path, fixedUUID).
			WithJSON(map[string]any{
				"text": "New Test Text",
				"link": "https://new-example.com",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.HasValue("id", fixedUUID)
		resp.HasValue("groupName", "Test Group")
		resp.HasValue("name", "Test Song")
		resp.Value("songDetail").Object().
			HasValue("releaseDate", fixedTime.Format("02.01.2006")).
			HasValue("text", "New Test Text").
			HasValue("link", "https://new-example.com")
		resp.HasValue("created_at", fixedTime)
		resp.HasValue("updated_at", fixedTime)
	})
}

func TestSongHandler_RemoveSong(t *testing.T) {
	const path = "/api/v1/songs/{songID}"

	t.Run("invalid song id", func(t *testing.T) {
		e, _ := setupServer(t)

		resp := e.DELETE(path, "invalid uuid").
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", invalidSongIDParamResp.Message)
	})

	t.Run("song not found", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("RemoveSong", mock.Anything, fixedUUID).
			Once().
			Return(int64(0), entity.ErrSongNotFound)

		resp := e.DELETE(path, fixedUUID).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", songNotFoundErrResp.Message)
	})

	t.Run("server error", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("RemoveSong", mock.Anything, fixedUUID).
			Once().
			Return(int64(0), errors.New("unknown error"))

		resp := e.DELETE(path, fixedUUID).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object()

		resp.HasValue("status", statusError)
		resp.HasValue("message", serverErrResp.Message)
	})

	t.Run("success", func(t *testing.T) {
		e, songUseCaseMock := setupServer(t)

		songUseCaseMock.
			On("RemoveSong", mock.Anything, fixedUUID).
			Once().
			Return(int64(1), nil)

		e.DELETE(path, fixedUUID).
			Expect().
			Status(http.StatusNoContent)
	})
}
