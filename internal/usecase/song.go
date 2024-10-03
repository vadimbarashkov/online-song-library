package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

type musicInfoAPI interface {
	FetchSongInfo(ctx context.Context, group, song string) (*entity.SongDetail, error)
}

type songRepository interface {
	Save(ctx context.Context, song entity.Song) (*entity.Song, error)
	GetAll(ctx context.Context, filter *entity.SongFilter, pagination *entity.Pagination) ([]*entity.Song, error)
	GetByID(ctx context.Context, songID uuid.UUID) (*entity.Song, error)
	Update(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error)
	Delete(ctx context.Context, songID uuid.UUID) (int64, error)
}

type SongUseCase struct {
	musicInfoApi musicInfoAPI
	songRepo     songRepository
}

func NewSongUseCase(musicInfoAPI musicInfoAPI, songRepo songRepository) *SongUseCase {
	return &SongUseCase{
		musicInfoApi: musicInfoAPI,
		songRepo:     songRepo,
	}
}

func (uc *SongUseCase) CreateSong(ctx context.Context, song entity.Song) (*entity.Song, error) {
	const op = "usecase.CreateSong"

	songDetail, err := uc.musicInfoApi.FetchSongInfo(ctx, song.Group, song.Song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch song detail from music info api: %w", op, err)
	}

	song.SongDetail = *songDetail

	savedSong, err := uc.songRepo.Save(ctx, song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save song: %w", op, err)
	}

	return savedSong, nil
}

func (uc *SongUseCase) FetchSongs(ctx context.Context, filter *entity.SongFilter, pagination *entity.Pagination) ([]*entity.Song, error) {
	const op = "usecase.FetchSongs"

	songs, err := uc.songRepo.GetAll(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch songs: %w", op, err)
	}

	return songs, nil
}

func (uc *SongUseCase) FetchSongText(ctx context.Context, songID uuid.UUID, pagination *entity.Pagination) (*entity.SongText, error) {
	const op = "usecase.FetchSongText"

	song, err := uc.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch song: %w", op, err)
	}

	verses := strings.Split(song.SongDetail.Text, "\n\n")

	if pagination != nil {
		if pagination.Page < 1 {
			pagination.Page = entity.DefaultPage
		}
		if pagination.Limit < 1 {
			pagination.Limit = entity.DefaultLimit
		}
	} else {
		pagination = entity.NewPagination(entity.DefaultPage, entity.DefaultLimit)
	}

	offset := pagination.GetOffset()
	if offset > len(verses) {
		offset = len(verses)
	}

	limit := offset + pagination.Limit
	if limit > len(verses) {
		limit = len(verses)
	}

	return &entity.SongText{
		ID:         song.ID,
		Group:      song.Group,
		Song:       song.Song,
		Text:       verses[offset:limit],
		CreateAt:   song.CreatedAt,
		UpdatedAt:  song.UpdatedAt,
		Pagination: *pagination,
	}, nil
}

func (uc *SongUseCase) ModifySong(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error) {
	const op = "usecase.ModifySong"

	updatedSong, err := uc.songRepo.Update(ctx, songID, song)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to modify song: %w", op, err)
	}

	return updatedSong, nil
}

func (uc *SongUseCase) RemoveSong(ctx context.Context, songID uuid.UUID) (int64, error) {
	const op = "usecase.RemoveSong"

	deleted, err := uc.songRepo.Delete(ctx, songID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to remove song: %w", op, err)
	}

	return deleted, nil
}
