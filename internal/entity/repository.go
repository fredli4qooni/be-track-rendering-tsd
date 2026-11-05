package entity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	GetWithinRadius(ctx context.Context, centerLat, centerLon, radiusNM float64) ([]Entity, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL + PostGIS backed Entity repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetAll(ctx context.Context, status string, entityType *int) ([]Entity, error) {
	query := `
		SELECT id, name, callsign, type, status, 
		       ST_Y(geom::geometry) AS latitude, 
		       ST_X(geom::geometry) AS longitude, 
		       altitude, course, speed, created_at, updated_at 
		FROM entities
	`
	var whereClauses []string
	var args []any
	argIdx := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if entityType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *entityType)
		argIdx++
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
		var createdAt, updatedAt any

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
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan entity: %w", err)
		}
		e.Status = EntityStatus(statusStr)
		e.CreatedAt = toTime(createdAt)
		e.UpdatedAt = toTime(updatedAt)
		entities = append(entities, e)
	}

	return entities, rows.Err()
}

func (r *postgresRepository) GetByID(ctx context.Context, id int64) (*Entity, error) {
	query := `
		SELECT id, name, callsign, type, status, 
		       ST_Y(geom::geometry) AS latitude, 
		       ST_X(geom::geometry) AS longitude, 
		       altitude, course, speed, created_at, updated_at
		FROM entities
		WHERE id = $1
	`
	var e Entity
	var statusStr string
	var createdAt, updatedAt any

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
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEntityNotFound
		}
		return nil, fmt.Errorf("get entity by id: %w", err)
	}

	e.Status = EntityStatus(statusStr)
	e.CreatedAt = toTime(createdAt)
	e.UpdatedAt = toTime(updatedAt)

	return &e, nil
}

func (r *postgresRepository) Create(ctx context.Context, entity *Entity) error {
	now := time.Now().UTC()
	entity.CreatedAt = now
	entity.UpdatedAt = now

	// ST_MakePoint takes (longitude, latitude)
	query := `
		INSERT INTO entities (name, callsign, type, status, geom, altitude, course, speed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($6, $5), 4326), $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		entity.Name,
		entity.Callsign,
		entity.Type,
		string(entity.Status),
		entity.Latitude,
		entity.Longitude,
		entity.Altitude,
		entity.Course,
		entity.Speed,
		now,
		now,
	).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)

	if err != nil {
		return fmt.Errorf("insert entity into postgres postgis: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, entity *Entity) error {
	entity.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE entities
		SET name = $1, callsign = $2, type = $3, status = $4, 
		    geom = ST_SetSRID(ST_MakePoint($6, $5), 4326), 
		    altitude = $7, course = $8, speed = $9, updated_at = $10
		WHERE id = $11
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
		entity.UpdatedAt,
		entity.ID,
	)
	if err != nil {
		return fmt.Errorf("update entity in postgres postgis: %w", err)
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

func (r *postgresRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM entities WHERE id = $1"
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

func (r *postgresRepository) BatchCreate(ctx context.Context, entities []Entity) error {
	if len(entities) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO entities (name, callsign, type, status, geom, altitude, course, speed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($6, $5), 4326), $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
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

func (r *postgresRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entities").Scan(&count)
	return count, err
}

// GetWithinRadius queries tactical entities within radius in Nautical Miles using PostGIS ST_DWithin and GIST spatial index.
func (r *postgresRepository) GetWithinRadius(ctx context.Context, centerLat, centerLon, radiusNM float64) ([]Entity, error) {
	// 1 Nautical Mile = 1852 meters in PostGIS geography calculation
	radiusMeters := radiusNM * 1852.0

	query := `
		SELECT id, name, callsign, type, status, 
		       ST_Y(geom::geometry) AS latitude, 
		       ST_X(geom::geometry) AS longitude, 
		       altitude, course, speed, created_at, updated_at
		FROM entities
		WHERE ST_DWithin(geom, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography, $3)
		ORDER BY ST_Distance(geom, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography) ASC
	`
	rows, err := r.db.QueryContext(ctx, query, centerLat, centerLon, radiusMeters)
	if err != nil {
		return nil, fmt.Errorf("spatial query within radius: %w", err)
	}
	defer rows.Close()

	results := make([]Entity, 0)
	for rows.Next() {
		var e Entity
		var statusStr string
		var createdAt, updatedAt any

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
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan spatial entity: %w", err)
		}
		e.Status = EntityStatus(statusStr)
		e.CreatedAt = toTime(createdAt)
		e.UpdatedAt = toTime(updatedAt)
		results = append(results, e)
	}

	return results, rows.Err()
}

func toTime(val any) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		layouts := []string{
			time.RFC3339,
			"2006-01-02 15:04:05-07",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t
			}
		}
	}
	return time.Now().UTC()
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
