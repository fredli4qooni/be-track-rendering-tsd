package entity_test

import (
	"context"
	"path/filepath"
	"testing"

	"be-explore-maptalks/internal/database"
	"be-explore-maptalks/internal/entity"
)

func setupTestDB(t *testing.T) entity.Repository {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_entities.db")

	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return entity.NewRepository(db)
}

func TestRepositoryCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// 1. Initial Count should be 0
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}

	// 2. Create Entity
	newEnt := &entity.Entity{
		Name:      "KRI Bung Tomo-357",
		Callsign:  "TOMO-357",
		Type:      0,
		Status:    entity.StatusActive,
		Latitude:  -5.912,
		Longitude: 106.012,
		Course:    45.0,
		Speed:     20.0,
	}

	if err := repo.Create(ctx, newEnt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if newEnt.ID == 0 {
		t.Fatal("expected assigned ID > 0")
	}

	// 3. GetByID
	fetched, err := repo.GetByID(ctx, newEnt.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fetched.Name != newEnt.Name {
		t.Errorf("expected name %q, got %q", newEnt.Name, fetched.Name)
	}
	if fetched.Status != entity.StatusActive {
		t.Errorf("expected status %v, got %v", entity.StatusActive, fetched.Status)
	}

	// 4. Update Entity
	fetched.Status = entity.StatusStandby
	fetched.Speed = 0.0
	fetched.Course = 90.0
	if err := repo.Update(ctx, fetched); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, newEnt.ID)
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if updated.Status != entity.StatusStandby {
		t.Errorf("expected status STANDBY, got %v", updated.Status)
	}
	if updated.Speed != 0.0 {
		t.Errorf("expected speed 0.0, got %v", updated.Speed)
	}

	// 5. GetAll with filters
	all, err := repo.GetAll(ctx, "STANDBY", nil)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 entity with STANDBY status, got %d", len(all))
	}

	// 6. Delete Entity
	if err := repo.Delete(ctx, newEnt.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 7. Verify Not Found after deletion
	_, err = repo.GetByID(ctx, newEnt.ID)
	if err != entity.ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestRepositoryBatchCreate(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	batch := []entity.Entity{
		{Name: "Target 01", Callsign: "TGT-01", Type: 1, Status: entity.StatusActive, Latitude: -5.8, Longitude: 106.1},
		{Name: "Target 02", Callsign: "TGT-02", Type: 2, Status: entity.StatusActive, Latitude: -5.7, Longitude: 106.2},
		{Name: "Target 03", Callsign: "TGT-03", Type: 3, Status: entity.StatusStandby, Latitude: -5.9, Longitude: 106.0},
	}

	if err := repo.BatchCreate(ctx, batch); err != nil {
		t.Fatalf("BatchCreate() error = %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

