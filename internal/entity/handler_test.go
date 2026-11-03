package entity_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-explore-maptalks/internal/entity"
)

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()
	repo := setupTestDB(t)
	svc := entity.NewService(repo)
	handler := entity.NewHandler(svc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}

func TestHandlerEndpoints(t *testing.T) {
	mux := setupTestServer(t)

	// 1. Health check
	t.Run("Health check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
	})

	// 2. Create Entity
	var createdID int64
	t.Run("Create entity success", func(t *testing.T) {
		payload := entity.CreateEntityRequest{
			Name:      "Test Coastal Radar",
			Callsign:  "RADAR-01",
			Type:      3, // Support
			Status:    "ACTIVE",
			Latitude:  -5.95,
			Longitude: 106.01,
			Course:    0.0,
			Speed:     0.0,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp entity.APIResponse[*entity.Entity]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data == nil || resp.Data.ID == 0 {
			t.Fatal("expected returned entity with valid ID")
		}
		createdID = resp.Data.ID
	})

	// 3. Create Entity validation failure
	t.Run("Create entity validation fail", func(t *testing.T) {
		payload := entity.CreateEntityRequest{
			Name:      "", // Invalid name
			Latitude:  120.0, // Invalid latitude
			Longitude: 106.01,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422 Unprocessable Entity, got %d", rec.Code)
		}
	})

	// 4. Get Entity
	t.Run("Get entity by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/entities/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
	})

	// 5. Update Entity
	t.Run("Update entity", func(t *testing.T) {
		updatedName := "Renamed Coastal Radar"
		payload := entity.UpdateEntityRequest{
			Name: &updatedName,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/entities/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	// 6. Delete Entity
	t.Run("Delete entity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/entities/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}

		// Verify 404 on subsequent get
		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/entities/1", nil)
		getRec := httptest.NewRecorder()
		mux.ServeHTTP(getRec, getReq)
		if getRec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found after deletion, got %d", getRec.Code)
		}
	})
	_ = createdID
}
