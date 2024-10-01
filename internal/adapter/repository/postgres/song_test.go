package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

var (
	errUnknown      = errors.New("unknown error")
	errAffectedRows = errors.New("affected rows error")

	columns = []string{"id", "group_name", "song", "release_date", "text", "link", "created_at", "updated_at"}

	fixedTime = time.Now()
)

func initSongRepository(t testing.TB) (*SongRepository, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		mockDB.Close()
	})

	db := sqlx.NewDb(mockDB, "sqlmock")
	t.Cleanup(func() {
		db.Close()
	})

	return NewSongRepository(db), mock
}

func TestSongRepository_Save(t *testing.T) {
	t.Run("empty song", func(t *testing.T) {
		repo, _ := initSongRepository(t)

		res, err := repo.Save(context.Background(), entity.Song{})

		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`INSERT INTO songs`).
			WithArgs("Test Group", "Test Song", fixedTime, "Test Text", "https://example.com").
			WillReturnError(errUnknown)

		savedSong, err := repo.Save(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
			SongDetail: entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, savedSong)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		rows := sqlmock.NewRows(columns).
			AddRow(songID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`INSERT INTO songs`).
			WithArgs("Test Group", "Test Song", fixedTime, "Test Text", "https://example.com").
			WillReturnRows(rows)

		savedSong, err := repo.Save(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
			SongDetail: entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, savedSong)
		assert.Equal(t, songID, savedSong.ID)
		assert.Equal(t, "Test Group", savedSong.Group)
		assert.Equal(t, "Test Song", savedSong.Song)
		assert.Equal(t, fixedTime, savedSong.SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", savedSong.SongDetail.Text)
		assert.Equal(t, "https://example.com", savedSong.SongDetail.Link)
		assert.Equal(t, fixedTime, savedSong.CreatedAt)
		assert.Equal(t, fixedTime, savedSong.UpdatedAt)
	})
}

func TestSongRepository_GetAll(t *testing.T) {
	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WillReturnError(errUnknown)

		res, err := repo.GetAll(context.Background())

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		rows := sqlmock.NewRows(columns).
			AddRow(songID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WillReturnRows(rows)

		res, err := repo.GetAll(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 1)
		assert.Equal(t, songID, res[0].ID)
		assert.Equal(t, "Test Group", res[0].Group)
		assert.Equal(t, "Test Song", res[0].Song)
		assert.Equal(t, fixedTime, res[0].SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", res[0].SongDetail.Text)
		assert.Equal(t, "https://example.com", res[0].SongDetail.Link)
		assert.Equal(t, fixedTime, res[0].CreatedAt)
		assert.Equal(t, fixedTime, res[0].UpdatedAt)
	})
}

func TestSongRepository_GetByID(t *testing.T) {
	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(songID).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.GetByID(context.Background(), songID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Nil(t, res)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(songID).
			WillReturnError(errUnknown)

		res, err := repo.GetByID(context.Background(), songID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		rows := sqlmock.NewRows(columns).
			AddRow(songID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(songID).
			WillReturnRows(rows)

		res, err := repo.GetByID(context.Background(), songID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, songID, res.ID)
		assert.Equal(t, "Test Group", res.Group)
		assert.Equal(t, "Test Song", res.Song)
		assert.Equal(t, fixedTime, res.SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", res.SongDetail.Text)
		assert.Equal(t, "https://example.com", res.SongDetail.Link)
		assert.Equal(t, fixedTime, res.CreatedAt)
		assert.Equal(t, fixedTime, res.UpdatedAt)
	})
}

func TestSongRepository_Update(t *testing.T) {
	t.Run("empty song", func(t *testing.T) {
		repo, _ := initSongRepository(t)

		res, err := repo.Update(context.Background(), uuid.New(), entity.Song{})

		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), songID).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.Update(context.Background(), songID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Nil(t, res)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), songID).
			WillReturnError(errUnknown)

		res, err := repo.Update(context.Background(), songID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		rows := sqlmock.NewRows(columns).
			AddRow(songID, "Test Group", "Test Song", fixedTime, "New Test Text", "https://new-example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), songID).
			WillReturnRows(rows)

		res, err := repo.Update(context.Background(), songID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, songID, res.ID)
		assert.Equal(t, "Test Group", res.Group)
		assert.Equal(t, "Test Song", res.Song)
		assert.Equal(t, fixedTime, res.SongDetail.ReleaseDate)
		assert.Equal(t, "New Test Text", res.SongDetail.Text)
		assert.Equal(t, "https://new-example.com", res.SongDetail.Link)
		assert.Equal(t, fixedTime, res.CreatedAt)
		assert.Equal(t, fixedTime, res.UpdatedAt)
	})
}

func TestSongRepository_Delete(t *testing.T) {
	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(songID).
			WillReturnError(errUnknown)

		deleted, err := repo.Delete(context.Background(), songID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Zero(t, deleted)
	})

	t.Run("rows affected error", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(songID).
			WillReturnResult(sqlmock.NewErrorResult(errAffectedRows))

		deleted, err := repo.Delete(context.Background(), songID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errAffectedRows)
		assert.Zero(t, deleted)
	})

	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(songID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		deleted, err := repo.Delete(context.Background(), songID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Zero(t, deleted)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)
		songID := uuid.New()

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(songID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		deleted, err := repo.Delete(context.Background(), songID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})
}
