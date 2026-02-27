package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

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

// PatternSignature represents a grouping key for recommendations
type PatternSignature struct {
	StartBucket string
	EndBucket   string
	DayType     string
}

// RecommendationGroup represents a group of recommendations with similar patterns
type RecommendationGroup struct {
	PatternDescription string           `json:"pattern_description"`
	PatternKey         string           `json:"pattern_key"`
	TotalDailySavings  float64          `json:"total_daily_savings"`
	InstanceCount      int              `json:"instance_count"`
	Recommendations    []map[string]any `json:"recommendations"`
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
		instance, err := h.instanceStore.GetInstanceByID(r.Context(), rec.InstanceID)
		if err != nil {
			log.Printf("DEBUG: Failed to get instance %s for recommendation %s: %v", rec.InstanceID, rec.ID, err)
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

	// Local functions for grouping
	hourToBucket := func(hour int) string {
		switch {
		case hour >= 6 && hour < 10:
			return "early-morning"
		case hour >= 10 && hour < 14:
			return "midday"
		case hour >= 14 && hour < 18:
			return "afternoon"
		case hour >= 18 && hour < 22:
			return "evening"
		default:
			return "night"
		}
	}

	daysToType := func(days []interface{}) string {
		if len(days) >= 7 {
			return "daily"
		}
		weekdays := 0
		weekends := 0
		for _, d := range days {
			day := d.(string)
			if day == "Saturday" || day == "Sunday" {
				weekends++
			} else {
				weekdays++
			}
		}
		if weekdays >= 4 && weekends == 0 {
			return "weekdays"
		}
		if weekends >= 2 && weekdays == 0 {
			return "weekends"
		}
		return "mixed"
	}

	generatePatternSignature := func(pattern map[string]interface{}) PatternSignature {
		startHour := int(pattern["idle_start_hour"].(float64))
		endHour := int(pattern["idle_end_hour"].(float64))
		daysOfWeek := pattern["days_of_week"].([]interface{})

		return PatternSignature{
			StartBucket: hourToBucket(startHour),
			EndBucket:   hourToBucket(endHour),
			DayType:     daysToType(daysOfWeek),
		}
	}

	formatHour := func(hour int) string {
		if hour == 0 {
			return "midnight"
		}
		if hour == 12 {
			return "noon"
		}
		if hour < 12 {
			return fmt.Sprintf("%dam", hour)
		}
		return fmt.Sprintf("%dpm", hour-12)
	}

	describePattern := func(pattern map[string]interface{}) string {
		startHour := int(pattern["idle_start_hour"].(float64))
		endHour := int(pattern["idle_end_hour"].(float64))
		daysOfWeek := pattern["days_of_week"].([]interface{})

		startTime := formatHour(startHour)
		endTime := formatHour(endHour)
		timeRange := fmt.Sprintf("%s to %s", startTime, endTime)

		dayType := daysToType(daysOfWeek)
		var dayDesc string
		switch dayType {
		case "weekdays":
			dayDesc = "weekdays"
		case "weekends":
			dayDesc = "weekends"
		case "daily":
			dayDesc = "daily"
		default:
			dayDesc = fmt.Sprintf("%d days/week", len(daysOfWeek))
		}

		return fmt.Sprintf("Idle %s, %s", timeRange, dayDesc)
	}

	groupRecommendations := func(recs []enrichedRec) []RecommendationGroup {
		groupMap := make(map[string]*RecommendationGroup)

		for _, rec := range recs {
			sig := generatePatternSignature(rec.DetectedPattern)
			key := fmt.Sprintf("%s-%s-%s", sig.StartBucket, sig.EndBucket, sig.DayType)

			if groupMap[key] == nil {
				groupMap[key] = &RecommendationGroup{
					PatternDescription: describePattern(rec.DetectedPattern),
					PatternKey:         key,
					Recommendations:    []map[string]any{},
				}
			}

			recMap := map[string]any{
				"id":                      rec.ID,
				"instance_id":             rec.InstanceID,
				"instance_name":           rec.InstanceName,
				"provider":                rec.Provider,
				"region":                  rec.Region,
				"engine":                  rec.Engine,
				"hourly_cost_cents":       rec.HourlyCostCents,
				"detected_pattern":        rec.DetectedPattern,
				"suggested_schedule":      rec.SuggestedSchedule,
				"confidence_score":        rec.ConfidenceScore,
				"estimated_daily_savings": rec.EstimatedDailySavings,
				"status":                  rec.Status,
				"created_at":              rec.CreatedAt,
			}
			groupMap[key].Recommendations = append(groupMap[key].Recommendations, recMap)
			groupMap[key].TotalDailySavings += rec.EstimatedDailySavings
			groupMap[key].InstanceCount++
		}

		var groups []RecommendationGroup
		for _, g := range groupMap {
			// Sort recommendations within group by savings
			sort.Slice(g.Recommendations, func(i, j int) bool {
				return g.Recommendations[i]["estimated_daily_savings"].(float64) > g.Recommendations[j]["estimated_daily_savings"].(float64)
			})
			groups = append(groups, *g)
		}

		// Sort groups by total savings (already sorted, but explicit for clarity)
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].TotalDailySavings > groups[j].TotalDailySavings
		})

		return groups
	}

	// Group recommendations by pattern signature
	groups := groupRecommendations(enriched)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": groups,
	})
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

	// Check for 24 hours of data requirement first with clear error message
	if err := h.checkDataSufficiency(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":      "Insufficient data",
			"message":    err.Error(),
			"suggestion": "GenerateRecommendations requires at least 24 hours of activity data per instance. Wait for data collection or check metrics_hourly table.",
		})
		return
	}

	recs, err := h.analyzer.GenerateRecommendations(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Failed to generate recommendations",
			"message": err.Error(),
		})
		return
	}

	// Store new recommendations
	created := 0
	for _, rec := range recs {
		if err := h.store.CreateRecommendation(&rec); err == nil {
			created++
		}
	}

	// Check if any recommendations were generated
	if created == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"created": 0,
			"message": "No new recommendations generated. All instances either have insufficient data or already have pending recommendations.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": created,
		"message": fmt.Sprintf("Generated %d new recommendations", created),
	})
}

// checkDataSufficiency verifies that at least one instance has 24+ hours of data
func (h *RecommendationHandler) checkDataSufficiency(ctx context.Context) error {
	instances, err := h.analyzer.GetInstanceIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	// Check each instance for sufficient data
	dataError := ""
	for _, instanceID := range instances {
		hasSufficient, err := h.analyzer.HasDataSufficient(ctx, instanceID)
		if err != nil {
			// Track the error for the final message
			if dataError == "" {
				dataError = err.Error()
			}
			// Continue checking other instances even if one fails
			continue
		}
		if hasSufficient {
			// At least one instance has enough data
			return nil
		}
	}

	// If we got here, either no instances had enough data or there was a data access error
	if dataError != "" {
		return fmt.Errorf("metrics data unavailable: %s", dataError)
	}

	return fmt.Errorf("no instances have 24+ hours of activity data. Please wait for metrics collection to accumulate data.")
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
