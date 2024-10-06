package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
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

// SongWithVerses represents a song with its lyrics broken down into verses.
type SongWithVerses struct {
	ID        uuid.UUID // Unique identifier for the song
	GroupName string    // Name of the musical group or artist
	Name      string    // Title of the song
	Verses    []string  // Lyrics of the song, divided into verses
	CreatedAt time.Time // Timestamp when the song was created
	UpdatedAt time.Time // Timestamp when the song was last updated
}

// SongFilterField defines the various fields that can be used to filter song queries.
const (
	SongGroupNameFilterField SongFilterField = iota
	SongNameFilterField
	SongReleaseYearFilterField
	SongReleaseDateFilterField
	SongReleaseDateAfterFilterField
	SongReleaseDateBeforeFilterField
	SongTextFilterField
)

// SongFilterField represents the type for specifying different song filter fields.
type SongFilterField int

// SongFilter defines the structure for filtering songs based on specific fields and values.
type SongFilter struct {
	Field SongFilterField // The field to filter by (e.g., name, group, release date)
	Value any             // The value to match against the specified field
}

// Pagination defaults for controlling the query result set.
const (
	DefaultOffset uint64 = 0
	DefaultLimit  uint64 = 20
)

// Pagination is used to control the pagination of query results by specifying the page number
// and the number of items per page (limit).
type Pagination struct {
	Offset uint64 // Current page number (1-based)
	Limit  uint64 // Maximum number of items per page
	Items  uint64 // The number of items in the current page
	Total  uint64 // The total number of items across all pages
}

// IsEmpty checks if the pagination values are not set.
func (p *Pagination) IsEmpty() bool {
	return p.Offset == 0 && p.Limit == 0
}

// SetDefault sets the default offset and limit values for pagination.
func (p *Pagination) SetDefault() {
	p.Offset = DefaultOffset
	p.Limit = DefaultLimit
}
