package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
	"github.com/vadimbarashkov/online-song-library/mocks/usecase"
)

var (
	fixedUUID = uuid.New()
	fixedTime = time.Now()
)

func initSongUseCase(t testing.TB) (*SongUseCase, *usecase.MockMusicInfoAPI, *usecase.MockSongRepository) {
	t.Helper()

	musicInfoAPIMock := usecase.NewMockMusicInfoAPI(t)
	songRepoMock := usecase.NewMockSongRepository(t)
	uc := NewSongUseCase(musicInfoAPIMock, songRepoMock)

	return uc, musicInfoAPIMock, songRepoMock
}

func TestSongUseCase_CreateSong(t *testing.T) {
	t.Run("music info api error", func(t *testing.T) {
		uc, musicInfoAPIMock, _ := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), "Test Group", "Test Song").
			Once().
			Return(nil, errors.New("api error"))

		song, err := uc.CreateSong(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch song detail from music info api")
		assert.Nil(t, song)
	})

	t.Run("song repository error", func(t *testing.T) {
		uc, musicInfoAPIMock, songRepoMock := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), "Test Group", "Test Song").
			Once().
			Return(&entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			}, nil)

		songRepoMock.
			On("Save", context.Background(), entity.Song{
				Group: "Test Group",
				Song:  "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
			}).
			Once().
			Return(nil, errors.New("unknown error"))

		song, err := uc.CreateSong(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to save song")
		assert.Nil(t, song)
	})

	t.Run("success", func(t *testing.T) {
		uc, musicInfoAPIMock, songRepoMock := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), "Test Group", "Test Song").
			Once().
			Return(&entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			}, nil)

		songRepoMock.
			On("Save", context.Background(), entity.Song{
				Group: "Test Group",
				Song:  "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
			}).
			Once().
			Return(&entity.Song{
				ID:    fixedUUID,
				Group: "Test Group",
				Song:  "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		song, err := uc.CreateSong(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
		})

		assert.NoError(t, err)
		assert.NotNil(t, song)
		assert.Equal(t, fixedUUID, song.ID)
		assert.Equal(t, "Test Group", song.Group)
		assert.Equal(t, "Test Song", song.Song)
		assert.Equal(t, fixedTime, song.SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", song.SongDetail.Text)
		assert.Equal(t, "https://example.com", song.SongDetail.Link)
		assert.Equal(t, fixedTime, song.CreatedAt)
		assert.Equal(t, fixedTime, song.UpdatedAt)
	})
}

func TestSongUseCase_FetchSongText(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetByID", context.Background(), fixedUUID).
			Once().
			Return(nil, errors.New("unknown error"))

		song, err := uc.FetchSongText(context.Background(), fixedUUID, nil)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch song")
		assert.Nil(t, song)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetByID", context.Background(), fixedUUID).
			Once().
			Return(&entity.Song{
				ID:    fixedUUID,
				Group: "Test Group",
				Song:  "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "line1\nline2\n\nline4\nline4\n",
					Link:        "https://example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		songText, err := uc.FetchSongText(context.Background(), fixedUUID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, songText)
		assert.Equal(t, fixedUUID, songText.ID)
		assert.Equal(t, "Test Group", songText.Group)
		assert.Equal(t, "Test Song", songText.Song)
		assert.Len(t, songText.Text, 2)
		assert.Equal(t, fixedTime, songText.CreateAt)
		assert.Equal(t, fixedTime, songText.UpdatedAt)
	})
}

func TestSongUseCase_FetchSongs(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		var filter *entity.SongFilter
		var pagination *entity.Pagination

		songRepoMock.
			On("GetAll", context.Background(), filter, pagination).
			Once().
			Return(nil, errors.New("unknown error"))

		songs, err := uc.FetchSongs(context.Background(), filter, pagination)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch songs")
		assert.Nil(t, songs)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		var filter *entity.SongFilter
		var pagination *entity.Pagination

		songRepoMock.
			On("GetAll", context.Background(), filter, pagination).
			Once().
			Return([]*entity.Song{
				{
					ID:    fixedUUID,
					Group: "Test Group",
					Song:  "Test Song",
					SongDetail: entity.SongDetail{
						ReleaseDate: fixedTime,
						Text:        "Test Text",
						Link:        "https://example.com",
					},
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			}, nil)

		songs, err := uc.FetchSongs(context.Background(), filter, pagination)

		assert.NoError(t, err)
		assert.NotNil(t, songs)
		assert.Len(t, songs, 1)
		assert.Equal(t, fixedUUID, songs[0].ID)
		assert.Equal(t, "Test Group", songs[0].Group)
		assert.Equal(t, "Test Song", songs[0].Song)
		assert.Equal(t, fixedTime, songs[0].SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", songs[0].SongDetail.Text)
		assert.Equal(t, "https://example.com", songs[0].Link)
		assert.Equal(t, fixedTime, songs[0].CreatedAt)
		assert.Equal(t, fixedTime, songs[0].UpdatedAt)
	})
}

func TestSongUseCase_ModifySong(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("Update", context.Background(), fixedUUID, entity.Song{
				SongDetail: entity.SongDetail{
					Text: "New Test Text",
					Link: "https://new-example.com",
				},
			}).
			Once().
			Return(nil, errors.New("unknown error"))

		song, err := uc.ModifySong(context.Background(), fixedUUID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to modify song")
		assert.Nil(t, song)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("Update", context.Background(), fixedUUID, entity.Song{
				SongDetail: entity.SongDetail{
					Text: "New Test Text",
					Link: "https://new-example.com",
				},
			}).
			Once().
			Return(&entity.Song{
				ID:    fixedUUID,
				Group: "Test Group",
				Song:  "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "New Test Text",
					Link:        "https://new-example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		song, err := uc.ModifySong(context.Background(), fixedUUID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, song)
		assert.Equal(t, fixedUUID, song.ID)
		assert.Equal(t, "Test Group", song.Group)
		assert.Equal(t, "Test Song", song.Song)
		assert.Equal(t, fixedTime, song.SongDetail.ReleaseDate)
		assert.Equal(t, "New Test Text", song.SongDetail.Text)
		assert.Equal(t, "https://new-example.com", song.SongDetail.Link)
		assert.Equal(t, fixedTime, song.CreatedAt)
		assert.Equal(t, fixedTime, song.UpdatedAt)
	})
}

func TestSongUseCase_RemoveSong(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("Delete", context.Background(), fixedUUID).
			Once().
			Return(int64(0), errors.New("unknown error"))

		deleted, err := uc.RemoveSong(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to remove song")
		assert.Zero(t, deleted)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("Delete", context.Background(), fixedUUID).
			Once().
			Return(int64(1), nil)

		deleted, err := uc.RemoveSong(context.Background(), fixedUUID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})
}
