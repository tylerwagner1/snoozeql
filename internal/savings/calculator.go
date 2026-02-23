package savings

import (
	"time"
)

// SavingsCalculator provides savings calculation logic
type SavingsCalculator struct {
	// No external dependencies - pure calculation logic
}

// NewSavingsCalculator creates a new savings calculator
func NewSavingsCalculator() *SavingsCalculator {
	return &SavingsCalculator{}
}

// MaxStoppedDuration is the maximum duration that can be considered for savings calculation.
// This is based on AWS's 7-day auto-restart limit for stopped instances.
// After 7 days, instances may be automatically restarted by AWS.
const MaxStoppedDuration = 7 * 24 * time.Hour

// CalculateSavings calculates savings based on stopped time and hourly cost.
// It caps the duration at MaxStoppedDuration (7 days) to reflect AWS's auto-restart limits.
//
// Parameters:
//   - stoppedAt: When the instance was stopped (start of the period)
//   - startedAt: When the instance was started (end of the period)
//   - hourlyCostCents: Hourly cost of the instance in cents
//
// Returns:
//   - stoppedMinutes: Total stopped time in minutes (capped at 7 days)
//   - savingsCents: Estimated savings in cents
func (c *SavingsCalculator) CalculateSavings(stoppedAt time.Time, startedAt time.Time, hourlyCostCents int) (int, int) {
	// Calculate the duration
	duration := startedAt.Sub(stoppedAt)

	// Cap at maximum stopped duration (7 days)
	if duration > MaxStoppedDuration {
		duration = MaxStoppedDuration
	}

	// Convert to minutes using integer math
	stoppedMinutes := int(duration.Minutes())

	// Calculate savings: (minutes * hourly_cost) / 60
	// Using integer math to maintain cents precision
	savingsCents := (stoppedMinutes * hourlyCostCents) / 60

	return stoppedMinutes, savingsCents
}

// CalculateOngoingSavings calculates savings for an instance that is currently stopped.
// It uses the current time as the end time and caps the duration at MaxStoppedDuration.
//
// Parameters:
//   - stoppedAt: When the instance was stopped
//   - hourlyCostCents: Hourly cost of the instance in cents
//
// Returns:
//   - stoppedMinutes: Total stopped time in minutes (capped at 7 days)
//   - savingsCents: Estimated ongoing savings in cents
func (c *SavingsCalculator) CalculateOngoingSavings(stoppedAt time.Time, hourlyCostCents int) (int, int) {
	// Use current time as the end time
	now := time.Now()

	// Calculate duration from stoppedAt to now
	duration := now.Sub(stoppedAt)

	// Cap at maximum stopped duration (7 days)
	if duration > MaxStoppedDuration {
		duration = MaxStoppedDuration
	}

	// Convert to minutes using integer math
	stoppedMinutes := int(duration.Minutes())

	// Calculate savings: (minutes * hourly_cost) / 60
	// Using integer math to maintain cents precision
	savingsCents := (stoppedMinutes * hourlyCostCents) / 60

	return stoppedMinutes, savingsCents
}

// DailySaving represents a split portion of a stop period on a specific day
type DailySaving struct {
	Date            time.Time
	StoppedMinutes  int
	SavingsCents    int
	HourlyRateCents int
}

// SplitByDay splits a stop period across multiple days, returning daily breakdowns.
// This is needed because a single stop period may span multiple calendar days,
// and we need daily aggregation for accurate reporting.
//
// Parameters:
//   - stoppedAt: When the instance was stopped
//   - startedAt: When the instance was started
//   - hourlyCostCents: Hourly cost of the instance in cents
//
// Returns:
//   - Slice of DailySaving structs, one for each calendar day in the period
func (c *SavingsCalculator) SplitByDay(stoppedAt time.Time, startedAt time.Time, hourlyCostCents int) []DailySaving {
	var dailySavings []DailySaving

	// Cap duration at 7 days
	duration := startedAt.Sub(stoppedAt)
	if duration > MaxStoppedDuration {
		duration = MaxStoppedDuration
		startedAt = stoppedAt.Add(duration)
	}

	// Get the start and end days
	startOfDay := time.Date(stoppedAt.Year(), stoppedAt.Month(), stoppedAt.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(startedAt.Year(), startedAt.Month(), startedAt.Day(), 23, 59, 59, 999999999, time.UTC)

	// Iterate day by day
	currentDay := startOfDay
	for currentDay.Before(endOfDay) || currentDay.Equal(endOfDay) {
		// Calculate the start and end of this day in the period
		dayStart := currentDay
		dayEnd := currentDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)

		// Find the actual overlap with our period
		actualStart := dayStart
		if dayStart.Before(stoppedAt) {
			actualStart = stoppedAt
		}
		actualEnd := dayEnd
		if dayEnd.After(startedAt) {
			actualEnd = startedAt
		}

		// If there's no overlap, skip this day
		if actualStart.After(actualEnd) {
			currentDay = currentDay.Add(24 * time.Hour)
			continue
		}

		// Calculate minutes for this day
		dayDuration := actualEnd.Sub(actualStart)
		dayMinutes := int(dayDuration.Minutes())

		// Calculate savings for this day using integer math
		daySavings := (dayMinutes * hourlyCostCents) / 60

		dailySavings = append(dailySavings, DailySaving{
			Date:            currentDay,
			StoppedMinutes:  dayMinutes,
			SavingsCents:    daySavings,
			HourlyRateCents: hourlyCostCents,
		})

		// Move to next day
		currentDay = currentDay.Add(24 * time.Hour)
	}

	return dailySavings
}
