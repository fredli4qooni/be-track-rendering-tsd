package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// InitPostgres initializes the PostgreSQL database with PostGIS extension and runs schema migration.
func InitPostgres(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait up to 15 seconds for database connection
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var pingErr error
	for i := 0; i < 15; i++ {
		if pingErr = db.PingContext(ctx); pingErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", pingErr)
	}

	if err := migratePostgres(db); err != nil {
		return nil, fmt.Errorf("failed to run postgres schema migration: %w", err)
	}

	log.Printf("[Database] Connected successfully to PostgreSQL + PostGIS database.")
	return db, nil
}

func migratePostgres(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queries := []string{
		// 1. Enable PostGIS extension
		`CREATE EXTENSION IF NOT EXISTS postgis;`,

		// 2. Create entities table with PostGIS GEOGRAPHY(Point, 4326)
		`CREATE TABLE IF NOT EXISTS entities (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			callsign VARCHAR(100) NOT NULL,
			type INTEGER NOT NULL DEFAULT 4,
			status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
			geom GEOGRAPHY(Point, 4326) NOT NULL,
			altitude DOUBLE PRECISION DEFAULT 0,
			course DOUBLE PRECISION DEFAULT 0,
			speed DOUBLE PRECISION DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,

		// 3. PostGIS Spatial GIST Index for high-speed spatial queries
		`CREATE INDEX IF NOT EXISTS idx_entities_geom ON entities USING GIST (geom);`,

		// 4. Standard B-Tree indexes for status and type filtering
		`CREATE INDEX IF NOT EXISTS idx_entities_status ON entities(status);`,
		`CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("migration step failed (%s): %w", query, err)
		}
	}

	return nil
}
