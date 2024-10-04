package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

// musicInfoAPI defines the interface for fetching song information from an external Music Info API.
type musicInfoAPI interface {
	FetchSongInfo(ctx context.Context, song entity.Song) (*entity.SongDetail, error)
}

// songRepository defines the interface for song repository operations.
type songRepository interface {
	Save(ctx context.Context, song entity.Song) (*entity.Song, error)
	GetAll(ctx context.Context, pagination entity.Pagination, filters ...entity.SongFilter) ([]*entity.Song, *entity.Pagination, error)
	GetByID(ctx context.Context, songID uuid.UUID) (*entity.Song, error)
	Update(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error)
	Delete(ctx context.Context, songID uuid.UUID) (int64, error)
}

// SongUseCase encapsulates the business logic for managing songs.
type SongUseCase struct {
	musicInfoApi musicInfoAPI
	songRepo     songRepository
}

// NewSongUseCase creates a new instance of SongUseCase with the provided musicInfoAPI and songRepository implementations.
func NewSongUseCase(musicInfoAPI musicInfoAPI, songRepo songRepository) *SongUseCase {
	return &SongUseCase{
		musicInfoApi: musicInfoAPI,
		songRepo:     songRepo,
	}
}

// CreateSong creates a new song by fetching its details from the music info API and saving it to the repository.
// It returns the saved song or an error if the process fails.
func (uc *SongUseCase) AddSong(ctx context.Context, song entity.Song) (*entity.Song, error) {
	const op = "usecase.AddSong"

	songDetail, err := uc.musicInfoApi.FetchSongInfo(ctx, song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch song detail from music info api: %w", op, err)
	}

	song.SongDetail = *songDetail

	savedSong, err := uc.songRepo.Save(ctx, song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to add song: %w", op, err)
	}

	return savedSong, nil
}

// FetchSongs retrieves all songs from the repository that match the provided filter and pagination parameters.
// It returns a slice of songs or an error if the retrieval fails.
func (uc *SongUseCase) FetchSongs(
	ctx context.Context,
	pagination entity.Pagination,
	filters ...entity.SongFilter,
) ([]*entity.Song, *entity.Pagination, error) {
	const op = "usecase.FetchSongs"

	songs, pgn, err := uc.songRepo.GetAll(ctx, pagination, filters...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to fetch songs: %w", op, err)
	}

	return songs, pgn, nil
}

// FetchSongText retrieves the text of a specific song by its ID, applying pagination if specified.
// It returns the song text or an error if the retrieval fails.
func (uc *SongUseCase) FetchSongWithVerses(
	ctx context.Context,
	songID uuid.UUID,
	pagination entity.Pagination,
) (*entity.SongWithVerses, *entity.Pagination, error) {
	const op = "usecase.FetchSongText"

	song, err := uc.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to fetch song: %w", op, err)
	}

	verses := strings.Split(song.SongDetail.Text, "\n\n")
	versesCount := uint64(len(verses))

	if pagination.IsEmpty() {
		pagination.SetDefault()
	}

	offset := pagination.Offset
	if offset > versesCount {
		offset = versesCount
	}

	limit := offset + pagination.Limit
	if limit > versesCount {
		limit = versesCount
	}

	pagination.Items = uint64(len(verses[offset:limit]))
	pagination.Total = versesCount

	return &entity.SongWithVerses{
		ID:        song.ID,
		GroupName: song.GroupName,
		Name:      song.Name,
		Verses:    verses[offset:limit],
		CreatedAt: song.CreatedAt,
		UpdatedAt: song.UpdatedAt,
	}, &pagination, nil
}

// ModifySong updates an existing song in the repository based on the provided song ID and new song data.
// It returns the updated song or an error if the modification fails.
func (uc *SongUseCase) ModifySong(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error) {
	const op = "usecase.ModifySong"

	updatedSong, err := uc.songRepo.Update(ctx, songID, song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify song: %w", op, err)
	}

	return updatedSong, nil
}

// RemoveSong deletes a song from the repository based on its ID.
// It returns the number of deleted records or an error if the deletion fails.
func (uc *SongUseCase) RemoveSong(ctx context.Context, songID uuid.UUID) (int64, error) {
	const op = "usecase.RemoveSong"

	deleted, err := uc.songRepo.Delete(ctx, songID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to remove song: %w", op, err)
	}

	return deleted, nil
}
