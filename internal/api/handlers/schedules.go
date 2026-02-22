// API handlers for schedules

package handlers

import (
	"encoding/json"
	"net/http"

	"snoozeql/internal/models"
	"snoozeql/internal/store"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	store *store.ScheduleStore
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(store *store.ScheduleStore) *ScheduleHandler {
	return &ScheduleHandler{store: store}
}

// GetAllSchedules returns all schedules
func (h *ScheduleHandler) GetAllSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.store.ListSchedules()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list schedules"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedules)
}

// GetSchedule returns a single schedule by ID
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := h.store.GetSchedule(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// CreateSchedule creates a new schedule
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var schedule models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if err := h.store.CreateSchedule(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create schedule"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// UpdateSchedule updates an existing schedule
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request, id string) {
	var schedule models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Ensure the ID from the URL is used
	schedule.ID = id

	if err := h.store.UpdateSchedule(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// DeleteSchedule deletes a schedule
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.store.DeleteSchedule(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnableSchedule enables a schedule
func (h *ScheduleHandler) EnableSchedule(w http.ResponseWriter, r *http.Request, id string) {
	// Update the schedule to enable it
	var schedule models.Schedule
	if existing, err := h.store.GetSchedule(id); err == nil {
		schedule = *existing
		schedule.Enabled = true

		if updateErr := h.store.UpdateSchedule(&schedule); updateErr != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled", "schedule_id": id})
}

// DisableSchedule disables a schedule
func (h *ScheduleHandler) DisableSchedule(w http.ResponseWriter, r *http.Request, id string) {
	// Update the schedule to disable it
	var schedule models.Schedule
	if existing, err := h.store.GetSchedule(id); err == nil {
		schedule = *existing
		schedule.Enabled = false

		if updateErr := h.store.UpdateSchedule(&schedule); updateErr != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "schedule_id": id})
}
