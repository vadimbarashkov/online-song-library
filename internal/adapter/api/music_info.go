package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
	"github.com/vadimbarashkov/online-song-library/pkg/validate"
)

// songDetailSchema defines the structure of the song details returned by the external API.
type songDetailSchema struct {
	ReleaseDate string `json:"releaseDate" validate:"required,releaseDate"`
	Text        string `json:"text" validate:"required"`
	Link        string `json:"link" validate:"required"`
}

// MusicInfoAPI is an API client used to fetch song information from an external music service.
type MusicInfoAPI struct {
	baseURL  string
	client   *http.Client
	validate *validator.Validate
}

// NewMusicInfoAPI creates a new instance of MusicInfoAPI with the provided base URL and HTTP client.
// If no client is provided, the default HTTP client is used. It also registers custom validations.
func NewMusicInfoAPI(baseURL string, client *http.Client) *MusicInfoAPI {
	if client == nil {
		client = http.DefaultClient
	}

	v := validator.New()
	_ = v.RegisterValidation("releaseDate", validate.ReleaseDateValidation)

	return &MusicInfoAPI{
		baseURL:  baseURL,
		client:   client,
		validate: v,
	}
}

// songDetailSchemaToEntity maps the external API song detail schema to the internal entity.SongDetail structure.
// It parses the release date from string format and returns a SongDetail entity.
func (api *MusicInfoAPI) songDetailSchemaToEntity(songDetail songDetailSchema) *entity.SongDetail {
	releaseDate, _ := time.Parse("02.01.2006", songDetail.ReleaseDate)

	return &entity.SongDetail{
		ReleaseDate: releaseDate,
		Text:        songDetail.Text,
		Link:        songDetail.Link,
	}
}

// FetchSongInfo retrieves song details from the external API by performing an HTTP GET request.
// The song's group name and title are passed as query parameters. It returns a SongDetail entity or an error.
func (api *MusicInfoAPI) FetchSongInfo(ctx context.Context, song entity.Song) (*entity.SongDetail, error) {
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
	query.Set("group", song.GroupName)
	query.Set("song", song.Name)
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

	if err := api.validate.Struct(songDetail); err != nil {
		return nil, fmt.Errorf("%s: validation error: %w", op, err)
	}

	return api.songDetailSchemaToEntity(songDetail), nil
}
