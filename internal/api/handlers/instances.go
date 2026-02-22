// API handlers for instances - to be implemented in Phase 3

package handlers

import (
	"encoding/json"
	"net/http"

	"snoozeql/internal/models"
	"snoozeql/internal/store"
)

// InstanceHandler handles instance-related HTTP requests
type InstanceHandler struct {
	store *store.Postgres
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(store *store.Postgres) *InstanceHandler {
	return &InstanceHandler{store: store}
}

// GetAllInstances returns all instances
func (h *InstanceHandler) GetAllInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]models.Instance{})
}

// GetInstance returns a single instance by ID
func (h *InstanceHandler) GetInstance(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.Instance{})
}

// CreateInstance creates a new instance
func (h *InstanceHandler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Instance{})
}

// UpdateInstance updates an existing instance
func (h *InstanceHandler) UpdateInstance(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.Instance{})
}

// DeleteInstance deletes an instance
func (h *InstanceHandler) DeleteInstance(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// StartInstance starts an instance
func (h *InstanceHandler) StartInstance(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "starting", "instance_id": id})
}

// StopInstance stops an instance
func (h *InstanceHandler) StopInstance(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping", "instance_id": id})
}
