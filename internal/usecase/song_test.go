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

func TestSongUseCase_AddSong(t *testing.T) {
	t.Run("music info api error", func(t *testing.T) {
		uc, musicInfoAPIMock, _ := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
			}).
			Once().
			Return(nil, errors.New("api error"))

		song, err := uc.AddSong(context.Background(), entity.Song{
			GroupName: "Test Group",
			Name:      "Test Song",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch song detail from music info api")
		assert.Nil(t, song)
	})

	t.Run("song repository error", func(t *testing.T) {
		uc, musicInfoAPIMock, songRepoMock := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
			}).
			Once().
			Return(&entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			}, nil)

		songRepoMock.
			On("Save", context.Background(), entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
			}).
			Once().
			Return(nil, errors.New("unknown error"))

		song, err := uc.AddSong(context.Background(), entity.Song{
			GroupName: "Test Group",
			Name:      "Test Song",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to add song")
		assert.Nil(t, song)
	})

	t.Run("success", func(t *testing.T) {
		uc, musicInfoAPIMock, songRepoMock := initSongUseCase(t)

		musicInfoAPIMock.
			On("FetchSongInfo", context.Background(), entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
			}).
			Once().
			Return(&entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			}, nil)

		songRepoMock.
			On("Save", context.Background(), entity.Song{
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
			}).
			Once().
			Return(&entity.Song{
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "Test Text",
					Link:        "https://example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		song, err := uc.AddSong(context.Background(), entity.Song{
			GroupName: "Test Group",
			Name:      "Test Song",
		})

		assert.NoError(t, err)
		assert.NotNil(t, song)
		assert.Equal(t, fixedUUID, song.ID)
		assert.Equal(t, "Test Group", song.GroupName)
		assert.Equal(t, "Test Song", song.Name)
		assert.Equal(t, fixedTime, song.SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", song.SongDetail.Text)
		assert.Equal(t, "https://example.com", song.SongDetail.Link)
		assert.Equal(t, fixedTime, song.CreatedAt)
		assert.Equal(t, fixedTime, song.UpdatedAt)
	})
}

func TestSongUseCase_FetchSongs(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetAll", context.Background(), entity.Pagination{}).
			Once().
			Return(nil, nil, errors.New("unknown error"))

		songs, pagination, err := uc.FetchSongs(context.Background(), entity.Pagination{})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch songs")
		assert.Nil(t, songs)
		assert.Nil(t, pagination)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetAll", context.Background(), entity.Pagination{}).
			Once().
			Return([]*entity.Song{
				{
					ID:        fixedUUID,
					GroupName: "Test Group",
					Name:      "Test Song",
					SongDetail: entity.SongDetail{
						ReleaseDate: fixedTime,
						Text:        "Test Text",
						Link:        "https://example.com",
					},
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			}, &entity.Pagination{
				Offset: entity.DefaultOffset,
				Limit:  entity.DefaultLimit,
				Items:  1,
				Total:  1,
			}, nil)

		songs, pagination, err := uc.FetchSongs(context.Background(), entity.Pagination{})

		assert.NoError(t, err)
		assert.NotNil(t, songs)
		assert.Len(t, songs, 1)
		assert.Equal(t, fixedUUID, songs[0].ID)
		assert.Equal(t, "Test Group", songs[0].GroupName)
		assert.Equal(t, "Test Song", songs[0].Name)
		assert.Equal(t, fixedTime, songs[0].SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", songs[0].SongDetail.Text)
		assert.Equal(t, "https://example.com", songs[0].Link)
		assert.Equal(t, fixedTime, songs[0].CreatedAt)
		assert.Equal(t, fixedTime, songs[0].UpdatedAt)
		assert.NotNil(t, pagination)
		assert.Equal(t, entity.DefaultOffset, pagination.Offset)
		assert.Equal(t, entity.DefaultLimit, pagination.Limit)
		assert.Equal(t, uint64(1), pagination.Items)
		assert.Equal(t, uint64(1), pagination.Total)
	})
}

func TestSongUseCase_FetchSongWithVerses(t *testing.T) {
	t.Run("song repository error", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetByID", context.Background(), fixedUUID).
			Once().
			Return(nil, errors.New("unknown error"))

		song, pagination, err := uc.FetchSongWithVerses(context.Background(), fixedUUID, entity.Pagination{})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch song")
		assert.Nil(t, song)
		assert.Nil(t, pagination)
	})

	t.Run("success", func(t *testing.T) {
		uc, _, songRepoMock := initSongUseCase(t)

		songRepoMock.
			On("GetByID", context.Background(), fixedUUID).
			Once().
			Return(&entity.Song{
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Song",
				SongDetail: entity.SongDetail{
					ReleaseDate: fixedTime,
					Text:        "line1\nline2\n\nline4\nline4\n",
					Link:        "https://example.com",
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			}, nil)

		song, pagination, err := uc.FetchSongWithVerses(context.Background(), fixedUUID, entity.Pagination{})

		assert.NoError(t, err)
		assert.NotNil(t, song)
		assert.Equal(t, fixedUUID, song.ID)
		assert.Equal(t, "Test Group", song.GroupName)
		assert.Equal(t, "Test Song", song.Name)
		assert.Len(t, song.Verses, 2)
		assert.Equal(t, fixedTime, song.CreatedAt)
		assert.Equal(t, fixedTime, song.UpdatedAt)
		assert.NotNil(t, pagination)
		assert.Equal(t, entity.DefaultOffset, pagination.Offset)
		assert.Equal(t, entity.DefaultLimit, pagination.Limit)
		assert.Equal(t, uint64(2), pagination.Items)
		assert.Equal(t, uint64(2), pagination.Total)
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
				ID:        fixedUUID,
				GroupName: "Test Group",
				Name:      "Test Song",
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
		assert.Equal(t, "Test Group", song.GroupName)
		assert.Equal(t, "Test Song", song.Name)
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
