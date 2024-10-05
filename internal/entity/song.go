package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultOffset uint64 = 0
	DefaultLimit  uint64 = 20
)

// ErrSongNotFound is returned when a requested song is not found in the database.
var ErrSongNotFound = errors.New("song not found")

// Song represents a musical composition with associated details.
type Song struct {
	ID         uuid.UUID // Unique identifier for the song
	GroupName  string    // Name of the musical group or artist
	Name       string    // Title of the song
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

type SongWithVerses struct {
	ID        uuid.UUID
	GroupName string
	Name      string
	Verses    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SongFilterField int

const (
	SongGroupNameFilterField SongFilterField = iota
	SongNameFilterField
	SongReleaseYearFilterField
	SongReleaseDateFilterField
	SongReleaseDateAfterFilterField
	SongReleaseDateBeforeFilterField
	SongTextFilterField
)

type SongFilter struct {
	Field SongFilterField
	Value any
}

// Pagination is used to control the pagination of query results by specifying the page number
// and the number of items per page (limit).
type Pagination struct {
	Offset uint64 // Current page number (1-based)
	Limit  uint64 // Maximum number of items per page
	Items  uint64
	Total  uint64
}

func (p *Pagination) IsEmpty() bool {
	return p.Offset == 0 && p.Limit == 0
}

func (p *Pagination) SetDefault() {
	p.Offset = DefaultOffset
	p.Limit = DefaultLimit
}
