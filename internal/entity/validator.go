package entity

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrNameEmpty        = errors.New("nama entity tidak boleh kosong")
	ErrNameTooShort     = errors.New("nama entity minimal 2 karakter")
	ErrNameTooLong      = errors.New("nama entity maksimal 60 karakter")
	ErrInvalidLatitude  = errors.New("latitude harus berada dalam rentang -90° s/d +90°")
	ErrInvalidLongitude = errors.New("longitude harus berada dalam rentang -180° s/d +180°")
	ErrInvalidCourse    = errors.New("course / haluan harus berada antara 0° s/d 360°")
	ErrInvalidSpeed     = errors.New("kecepatan harus berada antara 0 s/d 120 knots")
	ErrInvalidType      = errors.New("tipe entity tidak valid (harus berada dalam rentang 0 s/d 5)")
	ErrInvalidStatus    = errors.New("status entity tidak valid (ACTIVE, STANDBY, ENGAGED, INACTIVE, DESTROYED)")
)

// ValidateName checks that the entity name conforms to length requirements.
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrNameEmpty
	}
	if len(trimmed) < 2 {
		return ErrNameTooShort
	}
	if len(trimmed) > 60 {
		return ErrNameTooLong
	}
	return nil
}

// ValidateCoordinates ensures latitude is between -90 and 90, and longitude is between -180 and 180.
func ValidateCoordinates(lat, lon float64) error {
	if math.IsNaN(lat) || lat < -90.0 || lat > 90.0 {
		return ErrInvalidLatitude
	}
	if math.IsNaN(lon) || lon < -180.0 || lon > 180.0 {
		return ErrInvalidLongitude
	}
	return nil
}

// ValidateCourse ensures course angle is between 0 and 360 degrees.
func ValidateCourse(course float64) error {
	if math.IsNaN(course) || course < 0.0 || course > 360.0 {
		return ErrInvalidCourse
	}
	return nil
}

// ValidateSpeed ensures velocity in knots is between 0 and 120.
func ValidateSpeed(speed float64) error {
	if math.IsNaN(speed) || speed < 0.0 || speed > 120.0 {
		return ErrInvalidSpeed
	}
	return nil
}

// ValidateType checks if the tactical type code is valid (0 to 5).
func ValidateType(t int) error {
	if t < 0 || t > 5 {
		return ErrInvalidType
	}
	return nil
}

// ValidateStatus ensures the status string matches one of the operational statuses.
func ValidateStatus(status string) (EntityStatus, error) {
	if status == "" {
		return StatusActive, nil
	}
	upper := strings.ToUpper(strings.TrimSpace(status))
	switch EntityStatus(upper) {
	case StatusActive, StatusStandby, StatusEngaged, StatusInactive, StatusDestroyed:
		return EntityStatus(upper), nil
	default:
		return "", ErrInvalidStatus
	}
}

// ValidateCreateRequest performs comprehensive input validation on a CreateEntityRequest.
func ValidateCreateRequest(req *CreateEntityRequest) error {
	if req == nil {
		return errors.New("request payload cannot be empty")
	}
	if err := ValidateName(req.Name); err != nil {
		return err
	}
	if err := ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return err
	}
	if err := ValidateCourse(req.Course); err != nil {
		return err
	}
	if err := ValidateSpeed(req.Speed); err != nil {
		return err
	}
	if err := ValidateType(req.Type); err != nil {
		return err
	}
	if _, err := ValidateStatus(req.Status); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateRequest validates partial updates on an existing entity.
func ValidateUpdateRequest(req *UpdateEntityRequest) error {
	if req == nil {
		return errors.New("request payload cannot be empty")
	}
	if req.Name != nil {
		if err := ValidateName(*req.Name); err != nil {
			return err
		}
	}
	if req.Latitude != nil && (math.IsNaN(*req.Latitude) || *req.Latitude < -90.0 || *req.Latitude > 90.0) {
		return ErrInvalidLatitude
	}
	if req.Longitude != nil && (math.IsNaN(*req.Longitude) || *req.Longitude < -180.0 || *req.Longitude > 180.0) {
		return ErrInvalidLongitude
	}
	if req.Course != nil && (math.IsNaN(*req.Course) || *req.Course < 0.0 || *req.Course > 360.0) {
		return ErrInvalidCourse
	}
	if req.Speed != nil && (math.IsNaN(*req.Speed) || *req.Speed < 0.0 || *req.Speed > 120.0) {
		return ErrInvalidSpeed
	}
	if req.Type != nil {
		if err := ValidateType(*req.Type); err != nil {
			return err
		}
	}
	if req.Status != nil {
		if _, err := ValidateStatus(*req.Status); err != nil {
			return err
		}
	}
	return nil
}

// FormatValidationError formats validation errors into a user-friendly string.
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Validasi gagal: %s", err.Error())
}

