package entity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEntityNotFound = errors.New("entity tidak ditemukan")
)

// Repository defines the interface for entity data operations.
type Repository interface {
	GetAll(ctx context.Context, status string, entityType *int) ([]Entity, error)
	GetByID(ctx context.Context, id int64) (*Entity, error)
	Create(ctx context.Context, entity *Entity) error
	Update(ctx context.Context, entity *Entity) error
	Delete(ctx context.Context, id int64) error
	BatchCreate(ctx context.Context, entities []Entity) error
	Count(ctx context.Context) (int64, error)
}

type sqliteRepository struct {
	db *sql.DB
}

// NewRepository creates a new SQLite-backed Entity repository.
func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) GetAll(ctx context.Context, status string, entityType *int) ([]Entity, error) {
	query := "SELECT id, name, callsign, type, status, latitude, longitude, altitude, course, speed, created_at, updated_at FROM entities"
	var whereClauses []string
	var args []any

	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}
	if entityType != nil {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, *entityType)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY id ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query entities: %w", err)
	}
	defer rows.Close()

	entities := make([]Entity, 0)
	for rows.Next() {
		var e Entity
		var statusStr string
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Callsign,
			&e.Type,
			&statusStr,
			&e.Latitude,
			&e.Longitude,
			&e.Altitude,
			&e.Course,
			&e.Speed,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("scan entity: %w", err)
		}
		e.Status = EntityStatus(statusStr)
		e.CreatedAt, _ = parseTime(createdAtStr)
		e.UpdatedAt, _ = parseTime(updatedAtStr)
		entities = append(entities, e)
	}

	return entities, rows.Err()
}

func (r *sqliteRepository) GetByID(ctx context.Context, id int64) (*Entity, error) {
	query := `
		SELECT id, name, callsign, type, status, latitude, longitude, altitude, course, speed, created_at, updated_at
		FROM entities
		WHERE id = ?
	`
	var e Entity
	var statusStr string
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID,
		&e.Name,
		&e.Callsign,
		&e.Type,
		&statusStr,
		&e.Latitude,
		&e.Longitude,
		&e.Altitude,
		&e.Course,
		&e.Speed,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEntityNotFound
		}
		return nil, fmt.Errorf("get entity by id: %w", err)
	}

	e.Status = EntityStatus(statusStr)
	e.CreatedAt, _ = parseTime(createdAtStr)
	e.UpdatedAt, _ = parseTime(updatedAtStr)

	return &e, nil
}

func (r *sqliteRepository) Create(ctx context.Context, entity *Entity) error {
	now := time.Now().UTC()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	query := `
		INSERT INTO entities (name, callsign, type, status, latitude, longitude, altitude, course, speed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query,
		entity.Name,
		entity.Callsign,
		entity.Type,
		string(entity.Status),
		entity.Latitude,
		entity.Longitude,
		entity.Altitude,
		entity.Course,
		entity.Speed,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert entity: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		entity.ID = id
	}
	return nil
}

func (r *sqliteRepository) Update(ctx context.Context, entity *Entity) error {
	entity.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE entities
		SET name = ?, callsign = ?, type = ?, status = ?, latitude = ?, longitude = ?, altitude = ?, course = ?, speed = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		entity.Name,
		entity.Callsign,
		entity.Type,
		string(entity.Status),
		entity.Latitude,
		entity.Longitude,
		entity.Altitude,
		entity.Course,
		entity.Speed,
		entity.UpdatedAt.Format(time.RFC3339),
		entity.ID,
	)
	if err != nil {
		return fmt.Errorf("update entity: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (r *sqliteRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM entities WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (r *sqliteRepository) BatchCreate(ctx context.Context, entities []Entity) error {
	if len(entities) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO entities (name, callsign, type, status, latitude, longitude, altitude, course, speed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, e := range entities {
		_, err := stmt.ExecContext(ctx,
			e.Name,
			e.Callsign,
			e.Type,
			string(e.Status),
			e.Latitude,
			e.Longitude,
			e.Altitude,
			e.Course,
			e.Speed,
			now,
			now,
		)
		if err != nil {
			return fmt.Errorf("exec batch item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *sqliteRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entities").Scan(&count)
	return count, err
}

func parseTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unable to parse time string")
}
