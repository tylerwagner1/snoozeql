package analyzer

import (
	"encoding/json"
	"fmt"
	"time"

	"snoozeql/internal/models"
)

// idleWindowToRecommendation converts an IdleWindow to a Recommendation
func idleWindowToRecommendation(instance *models.Instance, window IdleWindow) *models.Recommendation {
	// Generate SuggestedSchedule JSONB
	suggestedSchedule := map[string]interface{}{
		"timezone":   "UTC",
		"sleep_cron": generateCronFromHour(window.StartHour, window.DaysOfWeek),
		"wake_cron":  generateCronFromHour(window.EndHour, window.DaysOfWeek),
	}
	scheduleJSON, _ := json.Marshal(suggestedSchedule)

	// Convert time.Weekday to string slice for JSONB
	daysStr := make([]string, len(window.DaysOfWeek))
	for i, d := range window.DaysOfWeek {
		daysStr[i] = d.String()
	}

	// Generate DetectedPattern JSONB
	detectedPattern := map[string]interface{}{
		"idle_start_hour": window.StartHour,
		"idle_end_hour":   window.EndHour,
		"days_of_week":    daysStr,
		"avg_cpu":         window.AvgCPU,
		"confidence":      window.Confidence,
	}
	patternJSON, _ := json.Marshal(detectedPattern)

	return &models.Recommendation{
		InstanceID:        instance.ID,
		DetectedPattern:   patternJSON,
		SuggestedSchedule: scheduleJSON,
		ConfidenceScore:   window.Confidence * 100, // Convert 0-1 to 0-100 scale
		Status:            "pending",
	}
}

// calculateEstimatedDailySavings calculates estimated daily savings
func calculateEstimatedDailySavings(hourlyCostCents int, startHour, endHour int) int {
	// Calculate idle hours (handling overnight windows)
	duration := segmentDuration(startHour, endHour)
	// Returns estimated daily savings in cents: (idle_hours * hourly_cost_cents)
	return duration * hourlyCostCents
}

// generateCronFromHour generates a CRON expression from an hour and days
func generateCronFromHour(hour int, days []time.Weekday) string {
	// Build days string (1-5 for weekdays, 0-6 for Sunday-Saturday)
	dayMap := map[time.Weekday]string{
		time.Sunday:    "0",
		time.Monday:    "1",
		time.Tuesday:   "2",
		time.Wednesday: "3",
		time.Thursday:  "4",
		time.Friday:    "5",
		time.Saturday:  "6",
	}

	var dayString string
	if len(days) == 7 {
		// Every day
		dayString = "*"
	} else if len(days) == 5 {
		// Check if it's weekdays only
		isWeekdays := true
		for _, d := range days {
			if d != time.Monday && d != time.Tuesday && d != time.Wednesday && d != time.Thursday && d != time.Friday {
				isWeekdays = false
				break
			}
		}
		if isWeekdays {
			dayString = "1-5"
		} else {
			var dayNums []string
			for _, d := range days {
				dayNums = append(dayNums, dayMap[d])
			}
			dayString = fmt.Sprintf("%v", dayNums)
		}
	} else {
		// General case - build day numbers
		var dayNums []string
		for _, d := range days {
			dayNums = append(dayNums, dayMap[d])
		}
		dayString = fmt.Sprintf("%v", dayNums)
	}

	return fmt.Sprintf("0 %d * * %s", hour, dayString)
}
