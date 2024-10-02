package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultPage defines the default page number for paginated results.
	DefaultPage = 1

	// DefaultLimit defines the default number of items per page for paginated results.
	DefaultLimit = 20
)

// ErrSongNotFound is returned when a requested song is not found in the database.
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

// SongFilter is used to filter song queries by various attributes.
type SongFilter struct {
	Group             string    // Filter by group name (partial match)
	Song              string    // Filter by song title (partial match)
	ReleaseYear       int       // Filter by year of release
	ReleaseDate       time.Time // Filter by exact release date
	ReleaseDateAfter  time.Time // Filter for songs released after a certain date
	ReleaseDateBefore time.Time // Filter for songs released before a certain date
	Text              string    // Filter by text content (partial match)
}

// Pagination is used to control the pagination of query results by specifying the page number
// and the number of items per page (limit).
type Pagination struct {
	Page  int // Current page number (1-based)
	Limit int // Maximum number of items per page
}

// NewPagination creates and returns a new Pagination object with the specified page and limit values.
// If the provided page is less than 1, it defaults to DefaultPage. Similarly, if the limit is less than 1,
// it defaults to DefaultLimit. This ensures that invalid values do not disrupt pagination logic.
func NewPagination(page, limit int) *Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}

	return &Pagination{
		Page:  page,
		Limit: limit,
	}
}

// GetOffset calculates the offset for paginated queries based on the current page and limit.
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}
