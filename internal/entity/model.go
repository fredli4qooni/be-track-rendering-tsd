package entity

import "time"

// EntityStatus represents the operational status of an entity.
type EntityStatus string

const (
	StatusActive    EntityStatus = "ACTIVE"
	StatusStandby   EntityStatus = "STANDBY"
	StatusEngaged   EntityStatus = "ENGAGED"
	StatusInactive  EntityStatus = "INACTIVE"
	StatusDestroyed EntityStatus = "DESTROYED"
)

// Entity represents a real-world or tactical tracked object on the map.
type Entity struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Callsign  string       `json:"callsign"`
	Type      int          `json:"type"` // 0: Friendly, 1: Hostile Confirmed, 2: Hostile Unconfirmed, 3: Support, 4: Unknown, 5: Waypoint
	Status    EntityStatus `json:"status"`
	Latitude  float64      `json:"lat"`
	Longitude float64      `json:"lon"`
	Altitude  float64      `json:"altitude,omitempty"`
	Course    float64      `json:"course"` // 0 - 360 degrees
	Speed     float64      `json:"speed"`  // knots (0 - 120)
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// CreateEntityRequest contains the required fields to create a new entity.
type CreateEntityRequest struct {
	Name      string  `json:"name"`
	Callsign  string  `json:"callsign,omitempty"`
	Type      int     `json:"type"`
	Status    string  `json:"status,omitempty"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Altitude  float64 `json:"altitude,omitempty"`
	Course    float64 `json:"course"`
	Speed     float64 `json:"speed"`
}

// UpdateEntityRequest contains the fields that can be updated for an existing entity.
type UpdateEntityRequest struct {
	Name      *string  `json:"name,omitempty"`
	Callsign  *string  `json:"callsign,omitempty"`
	Type      *int     `json:"type,omitempty"`
	Status    *string  `json:"status,omitempty"`
	Latitude  *float64 `json:"lat,omitempty"`
	Longitude *float64 `json:"lon,omitempty"`
	Altitude  *float64 `json:"altitude,omitempty"`
	Course    *float64 `json:"course,omitempty"`
	Speed     *float64 `json:"speed,omitempty"`
}

// APIResponse wraps all HTTP API responses in a standardized format.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

