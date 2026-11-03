package entity

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// Handler handles HTTP requests for entity endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new entity HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all entity API endpoints on the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/entities", h.ListEntities)
	mux.HandleFunc("GET /api/v1/entities/{id}", h.GetEntity)
	mux.HandleFunc("POST /api/v1/entities", h.CreateEntity)
	mux.HandleFunc("PUT /api/v1/entities/{id}", h.UpdateEntity)
	mux.HandleFunc("DELETE /api/v1/entities/{id}", h.DeleteEntity)
	mux.HandleFunc("POST /api/v1/entities/seed", h.SeedEntities)
	mux.HandleFunc("GET /api/v1/health", h.HealthCheck)
	mux.HandleFunc("GET /health", h.HealthCheck)
}

// ListEntities retrieves all entities with optional status or type filtering.
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	typeStr := r.URL.Query().Get("type")

	var typePtr *int
	if typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			typePtr = &t
		}
	}

	entities, err := h.service.GetAllEntities(r.Context(), status, typePtr)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse[any]{
			Success: false,
			Error:   "Gagal mengambil daftar entity: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[[]Entity]{
		Success: true,
		Message: "Daftar entity berhasil diambil",
		Data:    entities,
	})
}

// GetEntity retrieves a single entity by its numerical ID.
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse[any]{
			Success: false,
			Error:   "ID entity harus berupa angka valid",
		})
		return
	}

	entity, err := h.service.GetEntityByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrEntityNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse[any]{
				Success: false,
				Error:   "Entity tidak ditemukan",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIResponse[any]{
			Success: false,
			Error:   "Gagal mengambil data entity: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*Entity]{
		Success: true,
		Message: "Detail entity berhasil diambil",
		Data:    entity,
	})
}

// CreateEntity creates a new entity after validating payload inputs.
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	var req CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse[any]{
			Success: false,
			Error:   "Format JSON request tidak valid: " + err.Error(),
		})
		return
	}

	created, err := h.service.CreateEntity(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse[any]{
			Success: false,
			Error:   FormatValidationError(err),
		})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse[*Entity]{
		Success: true,
		Message: "Entity berhasil ditambahkan ke sistem",
		Data:    created,
	})
}

// UpdateEntity updates an existing entity's properties.
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse[any]{
			Success: false,
			Error:   "ID entity harus berupa angka valid",
		})
		return
	}

	var req UpdateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse[any]{
			Success: false,
			Error:   "Format JSON request tidak valid: " + err.Error(),
		})
		return
	}

	updated, err := h.service.UpdateEntity(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, ErrEntityNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse[any]{
				Success: false,
				Error:   "Entity dengan ID tersebut tidak ditemukan",
			})
			return
		}
		writeJSON(w, http.StatusUnprocessableEntity, APIResponse[any]{
			Success: false,
			Error:   FormatValidationError(err),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*Entity]{
		Success: true,
		Message: "Entity berhasil diperbarui",
		Data:    updated,
	})
}

// DeleteEntity removes an entity by ID.
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse[any]{
			Success: false,
			Error:   "ID entity harus berupa angka valid",
		})
		return
	}

	if err := h.service.DeleteEntity(r.Context(), id); err != nil {
		if errors.Is(err, ErrEntityNotFound) {
			writeJSON(w, http.StatusNotFound, APIResponse[any]{
				Success: false,
				Error:   "Entity tidak ditemukan",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIResponse[any]{
			Success: false,
			Error:   "Gagal menghapus entity: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[any]{
		Success: true,
		Message: "Entity berhasil dihapus dari sistem",
	})
}

// SeedEntities populates the database with realistic tactical entities if empty.
func (h *Handler) SeedEntities(w http.ResponseWriter, r *http.Request) {
	entities, err := h.service.SeedInitialEntities(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse[any]{
			Success: false,
			Error:   "Gagal mengisi seed data: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[[]Entity]{
		Success: true,
		Message: "Initial seed data berhasil disiapkan",
		Data:    entities,
	})
}

// HealthCheck verifies the server is running and responsive.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	count, _ := h.service.GetEntityCount(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "UP",
		"service":       "be-explore-maptalks",
		"version":       "1.0.0",
		"author":        "Fredli Fourqoni",
		"total_entities": count,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
