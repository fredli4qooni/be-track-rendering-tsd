package entity_test

import (
	"math"
	"testing"

	"be-explore-maptalks/internal/entity"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"single character", "A", true},
		{"valid short name", "AB", false},
		{"valid normal name", "KRI Raden Eddy Martadinata", false},
		{"name exactly 60 chars", "123456789012345678901234567890123456789012345678901234567890", false},
		{"name exceeds 60 chars", "1234567890123456789012345678901234567890123456789012345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entity.ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{"valid equator/meridian", 0.0, 0.0, false},
		{"valid Jakarta/Sunda Strait", -5.9, 106.0, false},
		{"valid extremes", 90.0, 180.0, false},
		{"valid negative extremes", -90.0, -180.0, false},
		{"latitude too high", 90.1, 100.0, true},
		{"latitude too low", -90.1, 100.0, true},
		{"longitude too high", 0.0, 180.1, true},
		{"longitude too low", 0.0, -180.1, true},
		{"NaN latitude", math.NaN(), 100.0, true},
		{"NaN longitude", 0.0, math.NaN(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entity.ValidateCoordinates(tt.lat, tt.lon)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCoordinates(%v, %v) err = %v, wantErr = %v", tt.lat, tt.lon, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCourseAndSpeed(t *testing.T) {
	// Test Course
	courseTests := []struct {
		course  float64
		wantErr bool
	}{
		{0.0, false},
		{180.5, false},
		{360.0, false},
		{-0.1, true},
		{360.1, true},
		{math.NaN(), true},
	}
	for _, tt := range courseTests {
		err := entity.ValidateCourse(tt.course)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateCourse(%v) err = %v, wantErr = %v", tt.course, err, tt.wantErr)
		}
	}

	// Test Speed
	speedTests := []struct {
		speed   float64
		wantErr bool
	}{
		{0.0, false},
		{25.4, false},
		{120.0, false},
		{-1.0, true},
		{120.1, true},
		{math.NaN(), true},
	}
	for _, tt := range speedTests {
		err := entity.ValidateSpeed(tt.speed)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateSpeed(%v) err = %v, wantErr = %v", tt.speed, err, tt.wantErr)
		}
	}
}

func TestValidateTypeAndStatus(t *testing.T) {
	// Test Type
	for i := 0; i <= 5; i++ {
		if err := entity.ValidateType(i); err != nil {
			t.Errorf("ValidateType(%d) unexpected error: %v", i, err)
		}
	}
	if err := entity.ValidateType(-1); err == nil {
		t.Error("ValidateType(-1) expected error, got nil")
	}
	if err := entity.ValidateType(6); err == nil {
		t.Error("ValidateType(6) expected error, got nil")
	}

	// Test Status
	validStatuses := []string{"ACTIVE", "active", "STANDBY", "ENGAGED", "INACTIVE", "DESTROYED", ""}
	for _, s := range validStatuses {
		if _, err := entity.ValidateStatus(s); err != nil {
			t.Errorf("ValidateStatus(%q) unexpected error: %v", s, err)
		}
	}
	if _, err := entity.ValidateStatus("UNKNOWN_STATUS"); err == nil {
		t.Error("ValidateStatus('UNKNOWN_STATUS') expected error, got nil")
	}
}

func TestValidateCreateRequest(t *testing.T) {
	validReq := &entity.CreateEntityRequest{
		Name:      "KRI Diponegoro-365",
		Callsign:  "DIPO-365",
		Type:      0,
		Status:    "ACTIVE",
		Latitude:  -6.10,
		Longitude: 106.85,
		Course:    90.0,
		Speed:     18.0,
	}

	if err := entity.ValidateCreateRequest(validReq); err != nil {
		t.Errorf("ValidateCreateRequest() valid request failed: %v", err)
	}

	if err := entity.ValidateCreateRequest(nil); err == nil {
		t.Error("ValidateCreateRequest(nil) expected error, got nil")
	}

	invalidReq := *validReq
	invalidReq.Name = ""
	if err := entity.ValidateCreateRequest(&invalidReq); err == nil {
		t.Error("ValidateCreateRequest() with empty name expected error, got nil")
	}
}

