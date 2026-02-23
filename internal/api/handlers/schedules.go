// API handlers for schedules

package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"snoozeql/internal/models"
	"snoozeql/internal/scheduler"
	"snoozeql/internal/store"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	scheduleStore *store.ScheduleStore
	instanceStore *store.InstanceStore
	eventStore    *store.EventStore
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleStore *store.ScheduleStore, instanceStore *store.InstanceStore, eventStore *store.EventStore) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleStore: scheduleStore,
		instanceStore: instanceStore,
		eventStore:    eventStore,
	}
}

// CreateEvent logs a schedule operation event
func (h *ScheduleHandler) CreateEvent(ctx context.Context, eventType, scheduleName, prevStatus, newStatus string) {
	if h.eventStore == nil {
		return
	}

	// Marshal metadata to JSON bytes
	metadata, _ := json.Marshal(map[string]string{"schedule_name": scheduleName})

	event := &models.Event{
		InstanceID:     "", // Schedule operations use the schedule ID
		EventType:      eventType,
		TriggeredBy:    "manual",
		PreviousStatus: prevStatus,
		NewStatus:      newStatus,
		Metadata:       metadata,
	}

	// Note: This will be stored in the events table
	// For schedule-specific events, we could add a schedule_id field to Event model
	// For now, we log to the events table but without instance association
	_ = h.eventStore.CreateEvent(ctx, event)
}

// GetAllSchedules returns all schedules
func (h *ScheduleHandler) GetAllSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.scheduleStore.ListSchedules()
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
	schedule, err := h.scheduleStore.GetSchedule(id)
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

	if err := h.scheduleStore.CreateSchedule(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create schedule"})
		return
	}

	// Log the event
	h.CreateEvent(r.Context(), "schedule_create", schedule.Name, "", "created")

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

	// Get the existing schedule name for logging
	var existingName string
	if existing, err := h.scheduleStore.GetSchedule(id); err == nil {
		existingName = existing.Name
	}

	// Ensure the ID from the URL is used
	schedule.ID = id

	if err := h.scheduleStore.UpdateSchedule(&schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	// Log the event
	h.CreateEvent(r.Context(), "schedule_update", existingName, "updated", "updated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// DeleteSchedule deletes a schedule
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request, id string) {
	// Get the schedule name for logging
	var scheduleName string
	if existing, err := h.scheduleStore.GetSchedule(id); err == nil {
		scheduleName = existing.Name
	}

	if err := h.scheduleStore.DeleteSchedule(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Schedule not found"})
		return
	}

	// Log the event
	h.CreateEvent(r.Context(), "schedule_delete", scheduleName, "deleted", "deleted")

	w.WriteHeader(http.StatusNoContent)
}

// EnableSchedule enables a schedule
func (h *ScheduleHandler) EnableSchedule(w http.ResponseWriter, r *http.Request, id string) {
	// Get the schedule name for logging
	var scheduleName string
	if existing, err := h.scheduleStore.GetSchedule(id); err == nil {
		scheduleName = existing.Name
		existing.Enabled = true // Modify the existing pointer directly

		if updateErr := h.scheduleStore.UpdateSchedule(existing); updateErr != nil {
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

	// Log the event
	h.CreateEvent(r.Context(), "schedule_enable", scheduleName, "enabled", "enabled")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled", "schedule_id": id})
}

// DisableSchedule disables a schedule
func (h *ScheduleHandler) DisableSchedule(w http.ResponseWriter, r *http.Request, id string) {
	// Get the schedule name for logging
	var scheduleName string
	if existing, err := h.scheduleStore.GetSchedule(id); err == nil {
		scheduleName = existing.Name
		existing.Enabled = false // Modify the existing pointer directly

		if updateErr := h.scheduleStore.UpdateSchedule(existing); updateErr != nil {
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

	// Log the event
	h.CreateEvent(r.Context(), "schedule_disable", scheduleName, "disabled", "disabled")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "schedule_id": id})
}

// PreviewFilter returns instances matching the given selectors
// POST /api/v1/schedules/preview-filter
func (h *ScheduleHandler) PreviewFilter(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Selectors []models.Selector `json:"selectors"`
		Operator  string            `json:"operator"` // "and" or "or", default "and"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Default to "and" if not specified
	if req.Operator == "" {
		req.Operator = "and"
	}

	// Validate operator
	if req.Operator != "and" && req.Operator != "or" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Operator must be 'and' or 'or'"})
		return
	}

	// Validate selectors
	if errMsg := scheduler.ValidateSelectors(req.Selectors); errMsg != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	// Get all instances
	instances, err := h.instanceStore.ListInstances(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list instances"})
		return
	}

	// Filter instances
	var matched []models.Instance
	for _, inst := range instances {
		if scheduler.MatchInstance(&inst, req.Selectors, req.Operator) {
			matched = append(matched, inst)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"matched_count": len(matched),
		"total_count":   len(instances),
		"instances":     matched,
	})
}
