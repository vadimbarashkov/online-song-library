package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fixedTime = time.Now()

func TestMusicInfoAPI_FetchSongInfo(t *testing.T) {
	t.Run("invalid base url", func(t *testing.T) {
		api := NewMusicInfoAPI("https://[::1]:namedport", nil)

		songDetail, err := api.FetchSongInfo(context.Background(), "Test Group", "Test Song")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to form path")
		assert.Nil(t, songDetail)
	})

	t.Run("request failure", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return nil, errors.New("network error")
				},
			},
		}

		api := NewMusicInfoAPI("https://example.com", client)

		songDetail, err := api.FetchSongInfo(context.Background(), "Test Group", "Test Song")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch song info")
		assert.Nil(t, songDetail)
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		api := NewMusicInfoAPI(server.URL, nil)

		songDetail, err := api.FetchSongInfo(context.Background(), "Test Group", "Test Song")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "unexpected status code: 500")
		assert.Nil(t, songDetail)
	})

	t.Run("invalid response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid JSON"))
		}))
		defer server.Close()

		api := NewMusicInfoAPI(server.URL, nil)

		songDetail, err := api.FetchSongInfo(context.Background(), "Test Group", "Test Song")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to decode response body")
		assert.Nil(t, songDetail)
	})

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/info", r.URL.Path)
			assert.Equal(t, "Test Group", r.URL.Query().Get("group"))
			assert.Equal(t, "Test Song", r.URL.Query().Get("song"))

			resp := songDetailSchema{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			}

			respData, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(respData)
		}))
		defer server.Close()

		api := NewMusicInfoAPI(server.URL, nil)

		songDetail, err := api.FetchSongInfo(context.Background(), "Test Group", "Test Song")

		assert.NoError(t, err)
		assert.NotNil(t, songDetail)
		assert.True(t, fixedTime.Equal(songDetail.ReleaseDate))
		assert.Equal(t, "Test Text", songDetail.Text)
		assert.Equal(t, "https://example.com", songDetail.Link)
	})
}
