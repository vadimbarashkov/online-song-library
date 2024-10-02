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
// This struct is used internally within the repository to map SQL query results.
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

// SongRepository provides methods for interacting with the 'songs' table in the database.
// It abstracts the details of SQL operations (insert, update, delete, etc.) and provides
// a clean interface for managing song records.
type SongRepository struct {
	db *sqlx.DB
}

// NewSongRepository creates a new instance of SongRepository and accepts a sqlx.DB object.
// This repository can be used to interact with the 'songs' table.
func NewSongRepository(db *sqlx.DB) *SongRepository {
	return &SongRepository{db: db}
}

// entityToRow converts an entity.Song object to a songRow. This helper function is used
// internally to prepare the song entity for database insertion or updates.
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

// entityToMap converts an entity.Song object to a map of field names and values,
// which is used to dynamically generate SQL UPDATE clauses. Fields that are empty or zero-valued
// are omitted from the resulting map, ensuring that only non-empty fields are updated.
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

// rowToEntity converts a songRow object (retrieved from the database) to an entity.Song object.
// This function helps to map database rows to domain entities, making the data accessible to the application.
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
// This function is useful when retrieving multiple rows from the database and transforming them
// into a format suitable for application-level use.
func (r *SongRepository) rowsToEntities(rows []songRow) []*entity.Song {
	songs := make([]*entity.Song, 0, len(rows))

	for _, row := range rows {
		songs = append(songs, r.rowToEntity(row))
	}

	return songs
}

// setFilterConditions adds SQL WHERE conditions to the query builder (squirrel.SelectBuilder) based on the
// provided SongFilter. It allows filtering the results by group name, song title, release year/date, and text content.
func (r *SongRepository) setFilterConditions(sb sq.SelectBuilder, filter entity.SongFilter) sq.SelectBuilder {
	if filter.Group != "" {
		sb = sb.Where("group_name ILIKE ?", fmt.Sprint("%", filter.Group, "%"))
	}
	if filter.Song != "" {
		sb = sb.Where("song ILIKE ?", fmt.Sprint("%", filter.Song, "%"))
	}
	if filter.ReleaseYear != 0 {
		sb = sb.Where("EXTRACT(YEAR FROM release_date) = ?", filter.ReleaseYear)
	}
	if !filter.ReleaseDate.IsZero() {
		sb = sb.Where(sq.Eq{"release_date": filter.ReleaseDate})
	}
	if !filter.ReleaseDateAfter.IsZero() {
		sb = sb.Where("release_date > ?", filter.ReleaseDateAfter)
	}
	if !filter.ReleaseDateBefore.IsZero() {
		sb = sb.Where("release_date < ?", filter.ReleaseDateBefore)
	}
	if filter.Text != "" {
		sb = sb.Where("text ILIKE ?", fmt.Sprint("%", filter.Text, "%"))
	}

	return sb
}

// Save inserts a new song record into the 'songs' table. If any required fields are missing, it returns an error.
// It returns the saved song entity if successful or an error if the operation fails.
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

// GetAll retrieves all song records that match the provided filter conditions and pagination settings.
// It uses the SongFilter struct for filtering the results and Pagination for controlling the result set size.
// The function returns a slice of song entities or an error if the operation fails.
func (r *SongRepository) GetAll(ctx context.Context, filter entity.SongFilter, pagination entity.Pagination) ([]*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.GetAll"

	if pagination.Page < 1 {
		pagination.Page = entity.DefaultPage
	}
	if pagination.Limit < 1 {
		pagination.Limit = entity.DefaultLimit
	}

	sb := sq.
		Select("*").From("songs").
		Offset(uint64(pagination.GetOffset())).
		Limit(uint64(pagination.Limit)).
		PlaceholderFormat(sq.Dollar)

	sb = r.setFilterConditions(sb, filter)

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var rows []songRow

	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("%s: failed to get rows from 'songs' table: %w", op, err)
	}

	return r.rowsToEntities(rows), nil
}

// GetByID retrieves a song by its ID from the 'songs' table.
// It returns the corresponding entity.Song object or an error if the song is not found.
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

// Update modifies an existing song record in the database, identified by its ID.
// It uses the entityToMap function to dynamically generate the SET clauses in the SQL update query.
// The method returns the updated song entity or an error if the song is not found or the update fails.
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

// Delete removes a song record from the 'songs' table, identified by its ID.
// The method returns the number of rows affected by the delete operation.
// If no rows were deleted, it returns an ErrSongNotFound error.
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
