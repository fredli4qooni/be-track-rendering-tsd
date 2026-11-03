package entity

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Service defines the business logic methods for entities.
type Service interface {
	GetAllEntities(ctx context.Context, status string, entityType *int) ([]Entity, error)
	GetEntityByID(ctx context.Context, id int64) (*Entity, error)
	CreateEntity(ctx context.Context, req *CreateEntityRequest) (*Entity, error)
	UpdateEntity(ctx context.Context, id int64, req *UpdateEntityRequest) (*Entity, error)
	DeleteEntity(ctx context.Context, id int64) error
	SeedInitialEntities(ctx context.Context) ([]Entity, error)
	GetEntityCount(ctx context.Context) (int64, error)
}

type entityService struct {
	repo Repository
}

// NewService creates a new entity service.
func NewService(repo Repository) Service {
	return &entityService{repo: repo}
}

func (s *entityService) GetAllEntities(ctx context.Context, status string, entityType *int) ([]Entity, error) {
	return s.repo.GetAll(ctx, status, entityType)
}

func (s *entityService) GetEntityByID(ctx context.Context, id int64) (*Entity, error) {
	if id <= 0 {
		return nil, ErrEntityNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *entityService) CreateEntity(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	status, _ := ValidateStatus(req.Status)
	callsign := strings.TrimSpace(req.Callsign)
	if callsign == "" {
		callsign = generateCallsign(req.Type)
	}

	entity := &Entity{
		Name:      strings.TrimSpace(req.Name),
		Callsign:  callsign,
		Type:      req.Type,
		Status:    status,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Altitude:  req.Altitude,
		Course:    req.Course,
		Speed:     req.Speed,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *entityService) UpdateEntity(ctx context.Context, id int64, req *UpdateEntityRequest) (*Entity, error) {
	if id <= 0 {
		return nil, ErrEntityNotFound
	}
	if err := ValidateUpdateRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Callsign != nil {
		existing.Callsign = strings.TrimSpace(*req.Callsign)
	}
	if req.Type != nil {
		existing.Type = *req.Type
	}
	if req.Status != nil {
		status, _ := ValidateStatus(*req.Status)
		existing.Status = status
	}
	if req.Latitude != nil {
		existing.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		existing.Longitude = *req.Longitude
	}
	if req.Altitude != nil {
		existing.Altitude = *req.Altitude
	}
	if req.Course != nil {
		existing.Course = *req.Course
	}
	if req.Speed != nil {
		existing.Speed = *req.Speed
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *entityService) DeleteEntity(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrEntityNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *entityService) GetEntityCount(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *entityService) SeedInitialEntities(ctx context.Context) ([]Entity, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return s.repo.GetAll(ctx, "", nil)
	}

	// Default seed entities positioned realistically around Sunda Strait / Java Sea
	seedData := []Entity{
		{
			Name:      "KRI Raden Eddy Martadinata-331",
			Callsign:  "REMI-331",
			Type:      0, // Friendly Suspected / Active
			Status:    StatusActive,
			Latitude:  -5.912,
			Longitude: 106.012,
			Course:    45.0,
			Speed:     22.5,
		},
		{
			Name:      "Hostile Surface Combatant 01",
			Callsign:  "BOGEY-RED-1",
			Type:      1, // Hostile Confirmed
			Status:    StatusEngaged,
			Latitude:  -5.820,
			Longitude: 106.180,
			Course:    220.0,
			Speed:     28.0,
		},
		{
			Name:      "Unidentified Fast Craft",
			Callsign:  "SKIFF-UNCONF",
			Type:      2, // Hostile Unconfirmed
			Status:    StatusActive,
			Latitude:  -5.750,
			Longitude: 106.090,
			Course:    185.0,
			Speed:     32.0,
		},
		{
			Name:      "Tactical Supply Depot Merak",
			Callsign:  "DEPOT-MERAK",
			Type:      3, // Support Point
			Status:    StatusStandby,
			Latitude:  -5.930,
			Longitude: 105.990,
			Course:    0.0,
			Speed:     0.0,
		},
		{
			Name:      "Commercial Tanker AIS-Unknown",
			Callsign:  "AIS-992144",
			Type:      4, // Unknown Surface
			Status:    StatusActive,
			Latitude:  -5.870,
			Longitude: 106.110,
			Course:    90.0,
			Speed:     14.2,
		},
		{
			Name:      "Checkpoint Alpha Bravo",
			Callsign:  "WP-ALPHA-01",
			Type:      5, // Waypoint
			Status:    StatusStandby,
			Latitude:  -5.850,
			Longitude: 106.050,
			Course:    0.0,
			Speed:     0.0,
		},
	}

	if err := s.repo.BatchCreate(ctx, seedData); err != nil {
		return nil, fmt.Errorf("failed to seed initial entities: %w", err)
	}

	return s.repo.GetAll(ctx, "", nil)
}

func generateCallsign(entityType int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(900) + 100
	switch entityType {
	case 0:
		return fmt.Sprintf("FRD-%03d", num)
	case 1, 2:
		return fmt.Sprintf("HST-%03d", num)
	case 3:
		return fmt.Sprintf("SUP-%03d", num)
	case 4:
		return fmt.Sprintf("UNK-%03d", num)
	case 5:
		return fmt.Sprintf("WP-%03d", num)
	default:
		return fmt.Sprintf("TRK-%03d", num)
	}
}
