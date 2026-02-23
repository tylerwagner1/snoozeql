package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"snoozeql/internal/metrics"
	"snoozeql/internal/models"
	"snoozeql/internal/provider"
)

// Analyzer manages activity analysis and pattern detection
type Analyzer struct {
	provider     *provider.Registry
	store        Store
	metricsStore *metrics.MetricsStore
	threshold    ThresholdConfig
}

// Store interface for recommendation persistence
type Store interface {
	GetRecommendation(id string) (*models.Recommendation, error)
	ListRecommendations(status string) ([]models.Recommendation, error)
	CreateRecommendation(recommendation *models.Recommendation) error
	UpdateRecommendation(recommendation *models.Recommendation) error
	ListRecommendationsByStatus(ctx context.Context, status string) ([]map[string]interface{}, error)
}

// ThresholdConfig defines detection thresholds
type ThresholdConfig struct {
	DefaultInactivityHours int
	DetectionDays          int
	ConfidenceMinimum      float64
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(provider *provider.Registry, store Store, metricsStore *metrics.MetricsStore, threshold ThresholdConfig) *Analyzer {
	return &Analyzer{
		provider:     provider,
		store:        store,
		metricsStore: metricsStore,
		threshold:    threshold,
	}
}

// RunAnalysis performs activity analysis on all managed instances
func (a *Analyzer) RunAnalysis(ctx context.Context) error {
	instances, err := a.provider.ListAllDatabases(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	for _, instance := range instances {
		if !instance.Managed {
			continue
		}

		metrics, err := a.provider.GetMetrics(ctx, instance.Provider, instance.ProviderID, "7d")
		if err != nil {
			fmt.Printf("Warning: Failed to get metrics for %s: %v\n", instance.Name, err)
			continue
		}

		pattern, err := a.detectPattern(metrics, instance)
		if err != nil {
			fmt.Printf("Warning: Failed to detect pattern for %s: %v\n", instance.Name, err)
			continue
		}

		if pattern != nil && pattern.Confidence >= a.threshold.ConfidenceMinimum {
			recommendation := a.generateRecommendation(instance, pattern)
			if err := a.store.CreateRecommendation(recommendation); err != nil {
				fmt.Printf("Warning: Failed to create recommendation for %s: %v\n", instance.Name, err)
			}
		}
	}

	return nil
}

// detectPattern analyzes metrics to find inactivity patterns
func (a *Analyzer) detectPattern(metrics map[string]any, instance models.Instance) (*DetectedPattern, error) {
	pattern := &DetectedPattern{
		InstanceID:    instance.ProviderID,
		Confidence:    0.0,
		DetectionDays: a.threshold.DetectionDays,
	}

	// Analyze connections metric for inactivity patterns
	if connData, ok := metrics["connections"]; ok {
		if connections, ok := connData.(map[string]any); ok {
			pattern = a.analyzeConnectionPattern(connections, instance)
		}
	}

	// Calculate confidence based on pattern consistency
	pattern.Confidence = a.calculateConfidence(pattern)

	return pattern, nil
}

// analyzeConnectionPattern analyzes connection metrics to detect inactive periods
func (a *Analyzer) analyzeConnectionPattern(connections map[string]any, instance models.Instance) *DetectedPattern {
	pattern := &DetectedPattern{
		InstanceID: instance.ProviderID,
	}

	// Check for zero connections pattern (completely inactive)
	if avg, ok := connections["average"].(float64); ok {
		if avg == 0 {
			pattern.InactiveStart = "00:00"
			pattern.InactiveEnd = "23:59"
			pattern.Days = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
			pattern.Confidence = 0.9
			return pattern
		}

		// Check for low average with spiking pattern (night inactivity)
		if avg < 1 {
			pattern.InactiveStart = "23:00"
			pattern.InactiveEnd = "07:00"
			pattern.Days = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
			pattern.Confidence = 0.85
			return pattern
		}
	}

	return pattern
}

// calculateConfidence calculates a confidence score for the pattern
func (a *Analyzer) calculateConfidence(pattern *DetectedPattern) float64 {
	confidence := pattern.Confidence

	// Increase confidence for longer detection periods
	if pattern.DetectionDays >= 7 {
		confidence += 0.1
	}

	// Decrease confidence if pattern is too broad
	if len(pattern.Days) == 7 && pattern.InactiveStart == "00:00" && pattern.InactiveEnd == "23:59" {
		confidence -= 0.2
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// generateRecommendation generates a schedule recommendation from a pattern
func (a *Analyzer) generateRecommendation(instance models.Instance, pattern *DetectedPattern) *models.Recommendation {
	recommendation := &models.Recommendation{
		InstanceID:      instance.ProviderID,
		ConfidenceScore: pattern.Confidence,
		Status:          "pending",
	}

	// Generate suggested schedule from pattern
	recommendation.SuggestedSchedule = a.generateScheduleFromPattern(pattern, instance)

	// Store detected pattern
	recommendation.DetectedPattern = a.encodePattern(pattern)

	return recommendation
}

// generateScheduleFromPattern creates a cron schedule from detected pattern
func (a *Analyzer) generateScheduleFromPattern(pattern *DetectedPattern, instance models.Instance) []byte {
	schedule := map[string]interface{}{
		"timezone":   "America/New_York",
		"sleep_cron": "0 19 * * 1-5",
		"wake_cron":  "0 7 * * 1-5",
	}

	// Enhance based on instance tags
	if env, ok := instance.Tags["Environment"]; ok {
		switch env {
		case "staging":
			schedule["timezone"] = "UTC"
			schedule["sleep_cron"] = "0 22 * * 1-5"
			schedule["wake_cron"] = "0 6 * * 1-5"
		case "development":
			schedule["sleep_cron"] = "0 20 * * 1-5"
			schedule["wake_cron"] = "0 8 * * 1-5"
		}
	}

	b, _ := json.Marshal(schedule)
	return b
}

// encodePattern encodes a pattern to JSON for storage
func (a *Analyzer) encodePattern(pattern *DetectedPattern) []byte {
	data := map[string]interface{}{
		"pattern_id":     "pattern-" + pattern.InstanceID,
		"instance_id":    pattern.InstanceID,
		"confidence":     pattern.Confidence,
		"inactive_start": pattern.InactiveStart,
		"inactive_end":   pattern.InactiveEnd,
		"detection_days": pattern.DetectionDays,
	}
	b, _ := json.Marshal(data)
	return b
}

// ThresholdConfig returns the current threshold configuration
func (a *Analyzer) ThresholdConfig() ThresholdConfig {
	return a.threshold
}

// AnalyzeInstanceActivity analyzes an instance's activity patterns from stored metrics
func (a *Analyzer) AnalyzeInstanceActivity(ctx context.Context, instanceID string) (*ActivityPattern, error) {
	// Check for sufficient data first
	hasSufficient, err := a.metricsStore.HasSufficientData(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check data sufficiency: %w", err)
	}

	if !hasSufficient {
		return &ActivityPattern{
			InstanceID:        instanceID,
			HasSufficientData: false,
			AnalyzedAt:        time.Now(),
		}, nil
	}

	// Get last 14 days of metrics
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -14)

	metrics, err := a.metricsStore.GetMetricsByInstance(ctx, instanceID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Analyze patterns using default thresholds from CONTEXT.md
	thresholds := DefaultThresholds()
	pattern := AnalyzeActivityPattern(metrics, thresholds)

	return pattern, nil
}

// AnalyzeAllInstances analyzes activity patterns for all instances with sufficient data
func (a *Analyzer) AnalyzeAllInstances(ctx context.Context) (map[string]*ActivityPattern, error) {
	instances, err := a.provider.ListAllDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	patterns := make(map[string]*ActivityPattern)

	for _, instance := range instances {
		if !instance.Managed {
			continue
		}

		pattern, err := a.AnalyzeInstanceActivity(ctx, instance.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to analyze %s: %v\n", instance.Name, err)
			continue
		}

		if len(pattern.IdleWindows) > 0 {
			patterns[instance.ID] = pattern
		}
	}

	return patterns, nil
}

// GenerateRecommendations generates recommendations from analyzed patterns
func (a *Analyzer) GenerateRecommendations(ctx context.Context) ([]models.Recommendation, error) {
	instances, err := a.provider.ListAllDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	patterns, err := a.AnalyzeAllInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze instances: %w", err)
	}

	var recommendations []models.Recommendation

	for instanceID, pattern := range patterns {
		if len(pattern.IdleWindows) == 0 {
			continue
		}

		// Find instance in the list
		var instance *models.Instance
		for i := range instances {
			if instances[i].ID == instanceID {
				instance = &instances[i]
				break
			}
		}
		if instance == nil {
			fmt.Printf("Warning: Could not find instance %s for recommendation\n", instanceID)
			continue
		}

		// Skip if already has pending recommendation
		existing, _ := a.store.ListRecommendationsByStatus(ctx, "pending")
		hasPending := false
		for _, rec := range existing {
			if rec["instance_id"] == instanceID {
				hasPending = true
				break
			}
		}
		if hasPending {
			continue
		}

		// Convert best IdleWindow to Recommendation
		window := pattern.IdleWindows[0] // Already sorted by confidence
		rec := idleWindowToRecommendation(instance, window)
		recommendations = append(recommendations, *rec)
	}

	return recommendations, nil
}

// DetectedPattern represents the detected pattern for storage
type DetectedPattern struct {
	PatternID       string   `json:"pattern_id"`
	InstanceID      string   `json:"instance_id"`
	Confidence      float64  `json:"confidence"`
	InactiveStart   string   `json:"inactive_start"`
	InactiveEnd     string   `json:"inactive_end"`
	DetectionDays   int      `json:"detection_days"`
	Days            []string `json:"days,omitempty"`
	ConfidenceScore float64  `json:"confidence_score"`
}
