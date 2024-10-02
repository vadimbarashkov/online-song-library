package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
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

	fixedUUID = uuid.New()
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

		song, err := repo.Save(context.Background(), entity.Song{
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
		assert.Nil(t, song)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`INSERT INTO songs`).
			WithArgs("Test Group", "Test Song", fixedTime, "Test Text", "https://example.com").
			WillReturnRows(rows)

		song, err := repo.Save(context.Background(), entity.Song{
			Group: "Test Group",
			Song:  "Test Song",
			SongDetail: entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			},
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

func TestSongRepository_GetAll(t *testing.T) {
	values := make([][]driver.Value, 0, 100)
	for i := 0; i < cap(values); i++ {
		groupName := fmt.Sprintf("Group %d", i)
		song := fmt.Sprintf("Song %d", i)

		values = append(values, []driver.Value{fixedUUID, groupName, song, fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime})
	}

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WillReturnError(errUnknown)

		res, err := repo.GetAll(context.Background(), entity.SongFilter{}, entity.Pagination{})

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, res)
	})

	t.Run("success with empty pagination", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).AddRows(values[:20]...)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WillReturnRows(rows)

		songs, err := repo.GetAll(context.Background(), entity.SongFilter{}, entity.Pagination{})

		assert.NoError(t, err)
		assert.NotNil(t, songs)
		assert.Len(t, songs, 20)
		assert.Equal(t, fixedUUID, songs[0].ID)
		assert.Equal(t, "Group 0", songs[0].Group)
		assert.Equal(t, "Song 0", songs[0].Song)
		assert.Equal(t, fixedTime, songs[0].SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", songs[0].SongDetail.Text)
		assert.Equal(t, "https://example.com", songs[0].Link)
		assert.Equal(t, fixedTime, songs[0].CreatedAt)
		assert.Equal(t, fixedTime, songs[0].UpdatedAt)
	})

	t.Run("success with not empty pagination", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).AddRows(values[40:50]...)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WillReturnRows(rows)

		songs, err := repo.GetAll(context.Background(), entity.SongFilter{}, entity.NewPagination(5, 10))

		assert.NoError(t, err)
		assert.NotNil(t, songs)
		assert.Len(t, songs, 10)
		assert.Equal(t, fixedUUID, songs[0].ID)
		assert.Equal(t, "Group 40", songs[0].Group)
		assert.Equal(t, "Song 40", songs[0].Song)
		assert.Equal(t, fixedTime, songs[0].SongDetail.ReleaseDate)
		assert.Equal(t, "Test Text", songs[0].SongDetail.Text)
		assert.Equal(t, "https://example.com", songs[0].Link)
		assert.Equal(t, fixedTime, songs[0].CreatedAt)
		assert.Equal(t, fixedTime, songs[0].UpdatedAt)
	})
}

func TestSongRepository_GetByID(t *testing.T) {
	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(fixedUUID).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.GetByID(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Nil(t, res)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(fixedUUID).
			WillReturnError(errUnknown)

		res, err := repo.GetByID(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(fixedUUID).
			WillReturnRows(rows)

		song, err := repo.GetByID(context.Background(), fixedUUID)

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

func TestSongRepository_Update(t *testing.T) {
	t.Run("empty song", func(t *testing.T) {
		repo, _ := initSongRepository(t)

		res, err := repo.Update(context.Background(), uuid.New(), entity.Song{})

		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), fixedUUID).
			WillReturnError(sql.ErrNoRows)

		res, err := repo.Update(context.Background(), fixedUUID, entity.Song{
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

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), fixedUUID).
			WillReturnError(errUnknown)

		res, err := repo.Update(context.Background(), fixedUUID, entity.Song{
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

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "New Test Text", "https://new-example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), fixedUUID).
			WillReturnRows(rows)

		song, err := repo.Update(context.Background(), fixedUUID, entity.Song{
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

func TestSongRepository_Delete(t *testing.T) {
	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(fixedUUID).
			WillReturnError(errUnknown)

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnknown)
		assert.Zero(t, deleted)
	})

	t.Run("rows affected error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(fixedUUID).
			WillReturnResult(sqlmock.NewErrorResult(errAffectedRows))

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errAffectedRows)
		assert.Zero(t, deleted)
	})

	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(fixedUUID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Zero(t, deleted)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(fixedUUID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})
}
