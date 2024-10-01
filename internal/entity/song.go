package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSongNotFound     = errors.New("song not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)

type Song struct {
	ID    uuid.UUID
	Group string
	Song  string
	SongDetail
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SongDetail struct {
	ReleaseDate time.Time
	Text        string
	Link        string
}
