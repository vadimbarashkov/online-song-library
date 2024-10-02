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

type Pagination struct {
	Page  int
	Limit int
}

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

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}
