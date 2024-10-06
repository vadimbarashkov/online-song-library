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
	columns = []string{"id", "group_name", "name", "release_date", "text", "link", "created_at", "updated_at"}

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
	t.Run("without not nill fields", func(t *testing.T) {
		repo, _ := initSongRepository(t)

		song, err := repo.Save(context.Background(), entity.Song{})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "missing required fields for saving song")
		assert.Nil(t, song)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`INSERT INTO songs`).
			WithArgs("Test Group", "Test Song", fixedTime, "Test Text", "https://example.com").
			WillReturnError(errors.New("unknown error"))

		song, err := repo.Save(context.Background(), entity.Song{
			GroupName: "Test Group",
			Name:      "Test Song",
			SongDetail: entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to insert row into 'songs' table")
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
			GroupName: "Test Group",
			Name:      "Test Song",
			SongDetail: entity.SongDetail{
				ReleaseDate: fixedTime,
				Text:        "Test Text",
				Link:        "https://example.com",
			},
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

func TestSongRepository_GetAll(t *testing.T) {
	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs LIMIT 20 OFFSET 0`).
			WithoutArgs().
			WillReturnError(errors.New("unknown error"))

		songs, pagination, err := repo.GetAll(context.Background(), entity.Pagination{})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get rows from 'songs' table")
		assert.Nil(t, songs)
		assert.Nil(t, pagination)
	})

	t.Run("success", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs LIMIT 20 OFFSET 0`).
			WithoutArgs().
			WillReturnRows(rows)

		rows = sqlmock.NewRows([]string{"total_count"}).AddRow(uint64(1))

		mock.
			ExpectQuery(`SELECT COUNT\(\*\)`).
			WithoutArgs().
			WillReturnRows(rows)

		songs, pagination, err := repo.GetAll(context.Background(), entity.Pagination{})

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

	t.Run("success with non-empty pagination", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs LIMIT 10 OFFSET 40`).
			WithoutArgs().
			WillReturnRows(rows)

		rows = sqlmock.NewRows([]string{"total_count"}).AddRow(uint64(1))

		mock.
			ExpectQuery(`SELECT COUNT\(\*\)`).
			WithoutArgs().
			WillReturnRows(rows)

		songs, pagination, err := repo.GetAll(context.Background(), entity.Pagination{
			Offset: 40,
			Limit:  10,
		})

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
		assert.Equal(t, uint64(40), pagination.Offset)
		assert.Equal(t, uint64(10), pagination.Limit)
		assert.Equal(t, uint64(1), pagination.Items)
		assert.Equal(t, uint64(1), pagination.Total)
	})

	t.Run("success with not-empty filters", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		rows := sqlmock.NewRows(columns).
			AddRow(fixedUUID, "Test Group", "Test Song", fixedTime, "Test Text", "https://example.com", fixedTime, fixedTime)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs WHERE name ILIKE \$1 AND EXTRACT\(YEAR FROM release_date\) = \$2 LIMIT 20 OFFSET 0`).
			WithArgs("%Song%", fixedTime.Year()).
			WillReturnRows(rows)

		rows = sqlmock.NewRows([]string{"total_count"}).AddRow(uint64(1))

		mock.
			ExpectQuery(`SELECT COUNT\(\*\)`).
			WithoutArgs().
			WillReturnRows(rows)

		songs, pagination, err := repo.GetAll(
			context.Background(),
			entity.Pagination{},
			entity.SongFilter{
				Field: entity.SongNameFilterField,
				Value: "Song",
			},
			entity.SongFilter{
				Field: entity.SongReleaseYearFilterField,
				Value: fixedTime.Year(),
			},
		)

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

func TestSongRepository_GetByID(t *testing.T) {
	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(fixedUUID).
			WillReturnError(sql.ErrNoRows)

		song, err := repo.GetByID(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Nil(t, song)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`SELECT (.+) FROM songs`).
			WithArgs(fixedUUID).
			WillReturnError(errors.New("unknown error"))

		res, err := repo.GetByID(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get row from 'songs' table")
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
		assert.Equal(t, "Test Group", song.GroupName)
		assert.Equal(t, "Test Song", song.Name)
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

		song, err := repo.Update(context.Background(), uuid.New(), entity.Song{})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "no fields provided for update")
		assert.Nil(t, song)
	})

	t.Run("song not found", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), fixedUUID).
			WillReturnError(sql.ErrNoRows)

		song, err := repo.Update(context.Background(), fixedUUID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrSongNotFound)
		assert.Nil(t, song)
	})

	t.Run("unknown database error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectQuery(`UPDATE songs`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), fixedUUID).
			WillReturnError(errors.New("unknown error"))

		res, err := repo.Update(context.Background(), fixedUUID, entity.Song{
			SongDetail: entity.SongDetail{
				Text: "New Test Text",
				Link: "https://new-example.com",
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to update row from 'songs' table")
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
		assert.Equal(t, "Test Group", song.GroupName)
		assert.Equal(t, "Test Song", song.Name)
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
			WillReturnError(errors.New("unknown error"))

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete row from 'songs' table")
		assert.Zero(t, deleted)
	})

	t.Run("rows affected error", func(t *testing.T) {
		repo, mock := initSongRepository(t)

		mock.
			ExpectExec(`DELETE FROM songs`).
			WithArgs(fixedUUID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		deleted, err := repo.Delete(context.Background(), fixedUUID)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get number of affected rows")
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
