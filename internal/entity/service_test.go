package entity_test

import (
	"context"
	"testing"

	"be-explore-maptalks/internal/entity"
)

func TestServiceValidationAndSeed(t *testing.T) {
	repo := setupTestDB(t)
	svc := entity.NewService(repo)
	ctx := context.Background()

	// 1. Test SeedInitialEntities
	seeded, err := svc.SeedInitialEntities(ctx)
	if err != nil {
		t.Fatalf("SeedInitialEntities() error = %v", err)
	}
	if len(seeded) == 0 {
		t.Fatal("expected at least 1 seeded entity")
	}

	count, err := svc.GetEntityCount(ctx)
	if err != nil {
		t.Fatalf("GetEntityCount() error = %v", err)
	}
	if count != int64(len(seeded)) {
		t.Errorf("count %d does not match seeded length %d", count, len(seeded))
	}

	// 2. Test CreateEntity validation failure
	invalidReq := &entity.CreateEntityRequest{
		Name:      "X", // too short
		Latitude:  -5.0,
		Longitude: 106.0,
	}
	_, err = svc.CreateEntity(ctx, invalidReq)
	if err == nil {
		t.Error("expected error for too short name, got nil")
	}

	// 3. Test CreateEntity with auto-generated callsign
	validReq := &entity.CreateEntityRequest{
		Name:      "Custom Patrol Vessel",
		Type:      0, // Friendly
		Latitude:  -5.85,
		Longitude: 106.12,
		Course:    120.0,
		Speed:     15.0,
	}
	created, err := svc.CreateEntity(ctx, validReq)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}
	if created.Callsign == "" {
		t.Error("expected auto-generated callsign, got empty")
	}

	// 4. Test UpdateEntity
	newName := "Updated Patrol Vessel Alpha"
	newSpeed := 24.5
	updateReq := &entity.UpdateEntityRequest{
		Name:  &newName,
		Speed: &newSpeed,
	}
	updated, err := svc.UpdateEntity(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Speed != newSpeed {
		t.Errorf("expected speed %v, got %v", newSpeed, updated.Speed)
	}

	// 5. Test DeleteEntity
	if err := svc.DeleteEntity(ctx, created.ID); err != nil {
		t.Fatalf("DeleteEntity() error = %v", err)
	}
}
