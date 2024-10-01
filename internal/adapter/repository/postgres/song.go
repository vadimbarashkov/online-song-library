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

type songRow struct {
	ID          uuid.UUID `db:"id"`
	Group       string    `db:"group_name"`
	Song        string    `db:"song"`
	ReleaseDate time.Time `db:"release_date"`
	Text        string    `db:"text"`
	Link        string    `db:"link"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type SongRepository struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) *SongRepository {
	return &SongRepository{db: db}
}

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

func (r *SongRepository) rowToEntity(row songRow) *entity.Song {
	return &entity.Song{
		ID:    row.ID,
		Group: row.Group,
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

func (r *SongRepository) rowsToEntity(rows []songRow) []*entity.Song {
	songs := make([]*entity.Song, 0, len(rows))

	for _, row := range rows {
		songs = append(songs, r.rowToEntity(row))
	}

	return songs
}

func (r *SongRepository) Save(ctx context.Context, song entity.Song) (*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.Save"

	query, args, err := sq.
		Insert("songs").Columns("group_name", "song", "release_date", "text", "link").
		Values(song.Group, song.Song, song.SongDetail.ReleaseDate, song.SongDetail.Text, song.SongDetail.Link).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql query: %w", op, err)
	}

	var row songRow

	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("%s: failed to insert row into 'songs' table: %w", op, err)
	}

	return r.rowToEntity(row), nil
}

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

	return r.rowsToEntity(rows), nil
}

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

func (r *SongRepository) Update(ctx context.Context, songID uuid.UUID, song entity.Song) (*entity.Song, error) {
	const op = "adapter.repository.postgres.SongRepository.Update"

	clauses := r.entityToMap(song)
	if len(clauses) == 0 {
		return nil, fmt.Errorf("%s: %w", op, entity.ErrNoFieldsToUpdate)
	}

	query, args, err := sq.
		Update("songs").
		SetMap(clauses).
		Where(sq.Eq{"id": songID}).
		Suffix("RETURNING *").
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

		return nil, fmt.Errorf("%s: failed to update row from 'songs' table: %w", op, err)
	}

	return r.rowToEntity(row), nil
}

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
