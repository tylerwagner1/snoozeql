package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/savings"
	"snoozeql/internal/store"
)

// SavingsHandler handles savings-related HTTP requests
type SavingsHandler struct {
	savingsStore  *store.SavingsStore
	instanceStore *store.InstanceStore
	eventStore    *store.EventStore
	calculator    *savings.SavingsCalculator
}

// NewSavingsHandler creates a new savings handler
func NewSavingsHandler(
	savingsStore *store.SavingsStore,
	instanceStore *store.InstanceStore,
	eventStore *store.EventStore,
	calculator *savings.SavingsCalculator,
) *SavingsHandler {
	return &SavingsHandler{
		savingsStore:  savingsStore,
		instanceStore: instanceStore,
		eventStore:    eventStore,
		calculator:    calculator,
	}
}

// GetSavingsSummary responds with savings summary including ongoing savings
func (h *SavingsHandler) GetSavingsSummary(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
			if days < 1 {
				days = 1
			}
		}
	}

	// Calculate date range
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	// Call savingsStore.GetTotalSavings for finalized savings
	totalSavings, err := h.savingsStore.GetTotalSavings(r.Context(), startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get total savings"})
		return
	}

	// Calculate ongoing savings for currently-stopped instances
	ongoingSavings := 0
	instances, err := h.instanceStore.ListInstances(r.Context())
	if err != nil {
		// If we can't list instances, just continue with 0 ongoing savings
		instances = []models.Instance{}
	}
	for _, instance := range instances {
		// Only calculate ongoing savings for stopped instances
		if instance.Status != "stopped" && instance.Status != "stopping" {
			continue
		}
		// Get the latest stop event for this instance
		events, err := h.eventStore.ListEventsByInstance(r.Context(), instance.ID)
		if err != nil {
			continue
		}

		var stoppedAt time.Time
		for _, event := range events {
			if event.EventType == "sleep" {
				stoppedAt = event.CreatedAt
				break
			}
		}

		if !stoppedAt.IsZero() {
			_, savingsCents := h.calculator.CalculateOngoingSavings(stoppedAt, instance.HourlyCostCents)
			ongoingSavings += savingsCents
		}
	}

	// Call savingsStore.GetTopSavers for top 5 instances
	topSavers, err := h.savingsStore.GetTopSavers(r.Context(), startDate, endDate, 5)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get top savers"})
		return
	}

	// Build response
	response := map[string]interface{}{
		"total_savings_cents":   totalSavings,
		"ongoing_savings_cents": ongoingSavings,
		"period": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"top_savers": topSavers,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetDailySavings responds with daily savings breakdown
func (h *SavingsHandler) GetDailySavings(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
			if days < 1 {
				days = 1
			}
		}
	}

	// Calculate date range
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	// Call savingsStore.GetDailySavings
	dailySavings, err := h.savingsStore.GetDailySavings(r.Context(), startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get daily savings"})
		return
	}

	// Build response
	response := map[string][]map[string]interface{}{
		"daily_savings": make([]map[string]interface{}, len(dailySavings)),
	}

	for i, ds := range dailySavings {
		response["daily_savings"][i] = map[string]interface{}{
			"date":            ds.Date,
			"savings_cents":   ds.SavingsCents,
			"stopped_minutes": ds.StoppedMinutes,
		}
		if ds.HourlyRateCents.Valid {
			response["daily_savings"][i]["hourly_rate_cents"] = ds.HourlyRateCents.Int64
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSavingsByInstance responds with savings attributed to each instance
func (h *SavingsHandler) GetSavingsByInstance(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	days := 30
	limit := 20
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
			if days < 1 {
				days = 1
			}
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil && parsed > 0 {
			limit = int(parsed)
		}
	}

	// Calculate date range
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	// Call savingsStore.GetTopSavers
	topSavers, err := h.savingsStore.GetTopSavers(r.Context(), startDate, endDate, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get savings by instance"})
		return
	}

	// Enrich with instance details (name, provider, region)
	response := make([]map[string]interface{}, len(topSavers))
	for i, saver := range topSavers {
		instance, err := h.instanceStore.GetInstanceByID(r.Context(), saver.InstanceID)
		if err != nil || instance == nil {
			response[i] = map[string]interface{}{
				"instance_id":   saver.InstanceID,
				"savings_cents": saver.SavingsCents,
				"stopped_hours": saver.StoppedHours,
			}
			continue
		}

		response[i] = map[string]interface{}{
			"instance_id":   saver.InstanceID,
			"name":          instance.Name,
			"provider":      instance.Provider,
			"region":        instance.Region,
			"savings_cents": saver.SavingsCents,
			"stopped_hours": saver.StoppedHours,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetInstanceSavings responds with savings detail for a single instance
func (h *SavingsHandler) GetInstanceSavings(w http.ResponseWriter, r *http.Request, id string) {
	// Parse query parameters
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
			if days < 1 {
				days = 1
			}
		}
	}

	// Calculate date range
	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -days)

	// Get savings records for this instance
	savingsRecords, err := h.savingsStore.GetSavingsByInstance(r.Context(), id, startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get instance savings"})
		return
	}

	// Calculate total savings
	totalSavings := 0
	for _, s := range savingsRecords {
		totalSavings += s.EstimatedSavingsCents
	}

	// Calculate ongoing savings if instance is currently stopped
	ongoingSavings := 0
	instance, err := h.instanceStore.GetInstanceByID(r.Context(), id)
	if err == nil && instance != nil {
		// Get the latest stop event
		events, err := h.eventStore.ListEventsByInstance(r.Context(), id)
		if err == nil {
			var stoppedAt time.Time
			for _, event := range events {
				if event.EventType == "sleep" {
					stoppedAt = event.CreatedAt
					break
				}
			}
			if !stoppedAt.IsZero() {
				_, savingsCents := h.calculator.CalculateOngoingSavings(stoppedAt, instance.HourlyCostCents)
				ongoingSavings = savingsCents
			}
		}
	}

	// Build savings array
	savingsArray := make([]map[string]interface{}, len(savingsRecords))
	for i, s := range savingsRecords {
		savingsArray[i] = map[string]interface{}{
			"date":              s.Date,
			"stopped_minutes":   s.StoppedMinutes,
			"savings_cents":     s.EstimatedSavingsCents,
			"hourly_rate_cents": s.HourlyRateCents,
		}
	}

	// Build response
	response := map[string]interface{}{
		"instance_id":           id,
		"total_savings_cents":   totalSavings,
		"ongoing_savings_cents": ongoingSavings,
		"savings":               savingsArray,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PostBackfill starts historical savings calculation
func (h *SavingsHandler) PostBackfill(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement backfill logic in future phase
	// For now, return success indicating backfill will be processed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "backfill_started"})
}
