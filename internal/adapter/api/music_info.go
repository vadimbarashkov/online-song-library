package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

// songDetailSchema defines the structure of the song details returned by the external API.
type songDetailSchema struct {
	ReleaseDate time.Time `json:"releaseDate"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
}

// MusicInfoAPI is an API client used to fetch song information from an external music service.
type MusicInfoAPI struct {
	baseURL string
	client  *http.Client
}

// NewMusicInfoAPI creates a new instance of MusicInfoAPI with the provided
// base URL and HTTP client. If no client is provided, the default client is used.
func NewMusicInfoAPI(baseURL string, client *http.Client) *MusicInfoAPI {
	if client == nil {
		client = http.DefaultClient
	}

	return &MusicInfoAPI{
		baseURL: baseURL,
		client:  client,
	}
}

// songDetailSchemaToEntity maps the external API song detail schema to the
// internal entity.SongDetail structure used within the application.
func (api *MusicInfoAPI) songDetailSchemaToEntity(songDetail songDetailSchema) *entity.SongDetail {
	return &entity.SongDetail{
		ReleaseDate: songDetail.ReleaseDate,
		Text:        songDetail.Text,
		Link:        songDetail.Link,
	}
}

// FetchSongInfo retrieves song details from the external API by
// performing an HTTP GET request. It requires the song group and title as parameters.
func (api *MusicInfoAPI) FetchSongInfo(ctx context.Context, group, song string) (*entity.SongDetail, error) {
	const op = "adapter.api.MusicInfoAPI.FetchSongInfo"

	path, err := url.JoinPath(api.baseURL, "/info")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to form path: %w", op, err)
	}

	url, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse url: %w", op, err)
	}

	query := url.Query()
	query.Set("group", group)
	query.Set("song", song)
	url.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", op, err)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch song info: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: unexpected status code: %d", op, resp.StatusCode)
	}

	var songDetail songDetailSchema

	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		return nil, fmt.Errorf("%s: failed to decode response body: %w", op, err)
	}

	return api.songDetailSchemaToEntity(songDetail), nil
}
