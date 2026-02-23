package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"snoozeql/internal/analyzer"
	"snoozeql/internal/models"
	"snoozeql/internal/provider"
	"snoozeql/internal/store"
)

// RecommendationHandler handles recommendation-related HTTP requests
type RecommendationHandler struct {
	store         *store.RecommendationStore
	instanceStore *store.InstanceStore
	scheduleStore *store.ScheduleStore
	provider      *provider.Registry
	analyzer      *analyzer.Analyzer
}

// NewRecommendationHandler creates a new recommendation handler
func NewRecommendationHandler(
	recStore *store.RecommendationStore,
	instStore *store.InstanceStore,
	schedStore *store.ScheduleStore,
	provider *provider.Registry,
	analyzer *analyzer.Analyzer,
) *RecommendationHandler {
	return &RecommendationHandler{
		store:         recStore,
		instanceStore: instStore,
		scheduleStore: schedStore,
		provider:      provider,
		analyzer:      analyzer,
	}
}

// GetAllRecommendations returns all recommendations with instance details
func (h *RecommendationHandler) GetAllRecommendations(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	recs, err := h.store.ListRecommendations(status)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list recommendations"})
		return
	}

	// Enrich each recommendation with instance details
	type enrichedRec struct {
		ID                    string         `json:"id"`
		InstanceID            string         `json:"instance_id"`
		InstanceName          string         `json:"instance_name"`
		Provider              string         `json:"provider"`
		Region                string         `json:"region"`
		Engine                string         `json:"engine"`
		HourlyCostCents       int            `json:"hourly_cost_cents"`
		DetectedPattern       map[string]any `json:"detected_pattern"`
		SuggestedSchedule     map[string]any `json:"suggested_schedule"`
		ConfidenceScore       float64        `json:"confidence_score"`
		EstimatedDailySavings float64        `json:"estimated_daily_savings"`
		Status                string         `json:"status"`
		CreatedAt             string         `json:"created_at"`
	}

	var enriched []enrichedRec
	for _, rec := range recs {
		// Parse detected pattern
		var detectedPattern map[string]any
		if len(rec.DetectedPattern) > 0 {
			json.Unmarshal(rec.DetectedPattern, &detectedPattern)
		}

		// Parse suggested schedule
		var suggestedSchedule map[string]any
		if len(rec.SuggestedSchedule) > 0 {
			json.Unmarshal(rec.SuggestedSchedule, &suggestedSchedule)
		}

		// Get instance to enrich with details
		instance, err := h.instanceStore.GetInstanceByProviderID(r.Context(), "", rec.InstanceID)
		if err != nil {
			// If not found, continue without instance details
			enriched = append(enriched, enrichedRec{
				ID:                rec.ID,
				InstanceID:        rec.InstanceID,
				DetectedPattern:   detectedPattern,
				SuggestedSchedule: suggestedSchedule,
				ConfidenceScore:   rec.ConfidenceScore,
				Status:            rec.Status,
				CreatedAt:         rec.CreatedAt.String(),
			})
			continue
		}

		// Calculate estimated daily savings from idle window
		var idleStartHour, idleEndHour int
		if detectedPattern != nil {
			if v, ok := detectedPattern["idle_start_hour"].(float64); ok {
				idleStartHour = int(v)
			}
			if v, ok := detectedPattern["idle_end_hour"].(float64); ok {
				idleEndHour = int(v)
			}
		}

		// Calculate idle hours (handling overnight windows)
		idleHours := idleEndHour - idleStartHour + 1
		if idleEndHour <= idleStartHour {
			idleHours = (24 - idleStartHour) + idleEndHour + 1
		}
		dailySavings := float64(idleHours*instance.HourlyCostCents) / 100.0

		enriched = append(enriched, enrichedRec{
			ID:                    rec.ID,
			InstanceID:            rec.InstanceID,
			InstanceName:          instance.Name,
			Provider:              instance.Provider,
			Region:                instance.Region,
			Engine:                instance.Engine,
			HourlyCostCents:       instance.HourlyCostCents,
			DetectedPattern:       detectedPattern,
			SuggestedSchedule:     suggestedSchedule,
			ConfidenceScore:       rec.ConfidenceScore,
			EstimatedDailySavings: dailySavings,
			Status:                rec.Status,
			CreatedAt:             rec.CreatedAt.String(),
		})
	}

	// Sort by estimated_daily_savings descending (highest savings first)
	for i := 0; i < len(enriched)-1; i++ {
		for j := i + 1; j < len(enriched); j++ {
			if enriched[j].EstimatedDailySavings > enriched[i].EstimatedDailySavings {
				enriched[i], enriched[j] = enriched[j], enriched[i]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(enriched)
}

// GetRecommendation returns a single recommendation by ID
func (h *RecommendationHandler) GetRecommendation(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := h.store.GetRecommendation(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Recommendation not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rec)
}

// GenerateRecommendations generates new recommendations from analyzed patterns
func (h *RecommendationHandler) GenerateRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	recs, err := h.analyzer.GenerateRecommendations(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Store new recommendations
	created := 0
	for _, rec := range recs {
		if err := h.store.CreateRecommendation(&rec); err == nil {
			created++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": created,
		"message": fmt.Sprintf("Generated %d new recommendations", created),
	})
}

// DismissRecommendation dismisses a schedule recommendation
func (h *RecommendationHandler) DismissRecommendation(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := h.store.GetRecommendation(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Recommendation not found"})
		return
	}

	rec.Status = "dismissed"
	if err := h.store.UpdateRecommendation(rec); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to dismiss recommendation"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "dismissed"})
}

// ConfirmRecommendation confirms and creates a schedule from a recommendation
func (h *RecommendationHandler) ConfirmRecommendation(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := h.store.GetRecommendation(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Recommendation not found"})
		return
	}

	// Parse suggested schedule
	var suggestedSchedule struct {
		Timezone  string `json:"timezone"`
		SleepCron string `json:"sleep_cron"`
		WakeCron  string `json:"wake_cron"`
	}
	json.Unmarshal(rec.SuggestedSchedule, &suggestedSchedule)

	// Get instance details
	instance, err := h.instanceStore.GetInstanceByProviderID(r.Context(), "", rec.InstanceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Instance not found"})
		return
	}

	// Create new schedule
	schedule := &models.Schedule{
		Name:        fmt.Sprintf("AI Suggested: %s", instance.Name),
		Description: "Auto-generated from activity pattern analysis",
		Timezone:    suggestedSchedule.Timezone,
		SleepCron:   suggestedSchedule.SleepCron,
		WakeCron:    suggestedSchedule.WakeCron,
		Selectors: []models.Selector{{
			Name: &models.Matcher{
				Pattern: instance.Name,
				Type:    "exact",
			},
		}},
		Enabled: true,
	}

	if err := h.scheduleStore.CreateSchedule(schedule); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create schedule"})
		return
	}

	// Update recommendation status to approved
	rec.Status = "approved"
	if err := h.store.UpdateRecommendation(rec); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update recommendation"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"schedule_id": schedule.ID,
		"status":      "approved",
	})
}
