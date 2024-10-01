package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vadimbarashkov/online-song-library/internal/entity"

	sq "github.com/Masterminds/squirrel"
)

// songRow represents a row in the 'songs' table of the database.
type songRow struct {
	ID          uuid.UUID `db:"id"`
	GroupName   string    `db:"group_name"`
	Song        string    `db:"song"`
	ReleaseDate time.Time `db:"release_date"`
	Text        string    `db:"text"`
	Link        string    `db:"link"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// SongRepository provides methods to interact with the songs table in the database.
type SongRepository struct {
	db *sqlx.DB
}

// NewSongRepository creates a new instance of SongRepository with the provided database connection.
func NewSongRepository(db *sqlx.DB) *SongRepository {
	return &SongRepository{db: db}
}

// entityToRow converts an entity.Song object to a songRow suitable for database operations.
func (r *SongRepository) entityToRow(song entity.Song) songRow {
	return songRow{
		ID:          song.ID,
		GroupName:   song.Group,
		Song:        song.Song,
		ReleaseDate: song.SongDetail.ReleaseDate,
		Text:        song.SongDetail.Text,
		Link:        song.SongDetail.Link,
		CreatedAt:   song.CreatedAt,
		UpdatedAt:   song.UpdatedAt,
	}
}

// entityToMap converts an entity.Song object to a map of database fields for updates.
func (r *SongRepository) entityToMap(song entity.Song) map[string]any {
	clauses := make(map[string]any)

	if song.Group != "" {
		clauses["group_name"] = song.Group
	}
	if song.Song != "" {
		clauses["song"] = song.Song
	}
	if !song.SongDetail.ReleaseDate.IsZero() {
		clauses["release_date"] = song.SongDetail.ReleaseDate
	}
	if song.SongDetail.Text != "" {
		clauses["text"] = song.SongDetail.Text
	}
	if song.SongDetail.Link != "" {
		clauses["link"] = song.SongDetail.Link
	}

	return clauses
}

// rowToEntity converts a songRow from the database to an entity.Song object.
func (r *SongRepository) rowToEntity(row songRow) *entity.Song {
	return &entity.Song{
		ID:    row.ID,
		Group: row.GroupName,
		Song:  row.Song,
		SongDetail: entity.SongDetail{
			ReleaseDate: row.ReleaseDate,
			Text:        row.Text,
			Link:        row.Link,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// rowsToEntities converts a slice of songRow objects to a slice of entity.Song pointers.
func (r *SongRepository) rowsToEntities(rows []songRow) []*entity.Song {
	songs := make([]*entity.Song, 0, len(rows))

	for _, row := range rows {
		songs = append(songs, r.rowToEntity(row))
	}

	return songs
}

// Save inserts a new song into the database and returns the saved entity.Song object.
// It returns an error if required fields are missing or if the insert operation fails.
func (r *SongRepository) Save(ctx context.Context, song entity.Song) (*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.Save"

	row := r.entityToRow(song)
	if row.GroupName == "" ||
		row.Song == "" ||
		row.ReleaseDate.IsZero() ||
		row.Text == "" ||
		row.Link == "" {

		return nil, fmt.Errorf("%s: missing required fields for saving song", op)
	}

	query, args, err := sq.
		Insert("songs").Columns("group_name", "song", "release_date", "text", "link").
		Values(row.GroupName, row.Song, row.ReleaseDate, row.Text, row.Link).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var savedRow songRow

	if err := r.db.GetContext(ctx, &savedRow, query, args...); err != nil {
		return nil, fmt.Errorf("%s: failed to insert row into 'songs' table: %w", op, err)
	}

	return r.rowToEntity(savedRow), nil
}

// GetAll retrieves all songs from the database and returns them as a slice of entity.Song pointers.
// It returns an error if the retrieval operation fails.
func (r *SongRepository) GetAll(ctx context.Context) ([]*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.GetAll"

	query, args, err := sq.
		Select("*").From("songs").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var rows []songRow

	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("%s: failed to get rows from 'songs' table: %w", op, err)
	}

	return r.rowsToEntities(rows), nil
}

// GetByID retrieves a song by its ID from the database.
// It returns the corresponding entity.Song object or an error if the song is not found or if the retrieval fails.
func (r *SongRepository) GetByID(ctx context.Context, songID uuid.UUID) (*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.GetByID"

	query, args, err := sq.
		Select("*").From("songs").
		Where(sq.Eq{"id": songID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var row songRow

	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrSongNotFound)
		}

		return nil, fmt.Errorf("%s: failed to get row from 'songs' table: %w", op, err)
	}

	return r.rowToEntity(row), nil
}

// Update modifies an existing song in the database identified by its ID.
// It returns the updated entity.Song object or an error if the song is not found or if the update fails.
func (r *SongRepository) Update(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.Update"

	clauses := r.entityToMap(song)
	if len(clauses) == 0 {
		return nil, fmt.Errorf("%s: no fields provided for update", op)
	}

	ub := sq.
		Update("songs").
		SetMap(clauses).
		Where(sq.Eq{"id": songID}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	query, args, err := ub.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var updatedRow songRow

	if err := r.db.GetContext(ctx, &updatedRow, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, entity.ErrSongNotFound)
		}

		return nil, fmt.Errorf("%s: failed to update row from 'songs' table: %w", op, err)
	}

	return r.rowToEntity(updatedRow), nil
}

// Delete removes a song from the database identified by its ID.
// It returns the number of deleted rows or an error if the deletion fails or the song is not found.
func (r *SongRepository) Delete(ctx context.Context, songID uuid.UUID) (int64, error) {
	const op = "adapter.repository.postgres.SongRepository.Delete"

	query, args, err := sq.
		Delete("songs").
		Where(sq.Eq{"id": songID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to delete row from 'songs' table: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get number of affected rows: %w", op, err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("%s: %w", op, entity.ErrSongNotFound)
	}

	return rowsAffected, nil
}
