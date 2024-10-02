package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
)

var ErrSongNotFound = errors.New("song not found")

// Song represents a musical composition with associated details.
type Song struct {
	ID         uuid.UUID // Unique identifier for the song
	Group      string    // Name of the musical group or artist
	Song       string    // Title of the song
	SongDetail           // Contains additional details about the song
	CreatedAt  time.Time // Timestamp when the song was created
	UpdatedAt  time.Time // Timestamp when the song was last updated
}

// SongDetail holds detailed information about a song.
type SongDetail struct {
	ReleaseDate time.Time // Release date of the song
	Text        string    // Lyrics or text of the song
	Link        string    // Link to the song (e.g., streaming link)
}

// Pagination is used to control the page and limit of results for paginated queries.
type Pagination struct {
	Page  int
	Limit int
}

// NewPagination returns a new Pagination struct with default values if the provided page or limit are invalid.
// It ensures that the page and limit values are at least 1 and the default limit, respectively.
func NewPagination(page, limit int) Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}

	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

// GetOffset calculates and returns the offset based on the current page and limit.
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}
