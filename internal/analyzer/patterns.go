package analyzer

import (
	"sort"
	"time"

	"snoozeql/internal/models"
)

// ActivityThresholds defines the thresholds for "low activity" per CONTEXT.md
type ActivityThresholds struct {
	CPUPercent        float64 // CPU < 1%
	QueriesPerMin     float64 // Queries < 5/min (approximated from connections/IOPS)
	MinIdleHours      int     // 8+ hours of low activity
	MinDataHours      int     // 24+ hours of data required
	MinDaysConsistent int     // Require pattern on 3+ days
}

// DefaultThresholds returns the thresholds from CONTEXT.md
func DefaultThresholds() ActivityThresholds {
	return ActivityThresholds{
		CPUPercent:        1.0, // Near-zero activity (CPU < 1%)
		QueriesPerMin:     5.0, // Queries < 5/min
		MinIdleHours:      8,   // 8+ hours total low activity
		MinDataHours:      24,  // 24+ hours of data required
		MinDaysConsistent: 3,   // Pattern on at least 3 days
	}
}

// IdleWindow represents a detected period of low activity
type IdleWindow struct {
	StartHour   int            `json:"start_hour"`   // 0-23
	EndHour     int            `json:"end_hour"`     // 0-23 (can be < start for overnight)
	DaysOfWeek  []time.Weekday `json:"days_of_week"` // Which days this pattern was found
	AvgCPU      float64        `json:"avg_cpu"`      // Average CPU during window
	AvgConns    float64        `json:"avg_conns"`    // Average connections during window
	Confidence  float64        `json:"confidence"`   // 0.0-1.0 confidence score
	IsOvernight bool           `json:"is_overnight"` // True if crosses midnight
}

// ActivityPattern represents the analysis result for an instance
type ActivityPattern struct {
	InstanceID        string       `json:"instance_id"`
	IdleWindows       []IdleWindow `json:"idle_windows"`
	HasSufficientData bool         `json:"has_sufficient_data"`
	DataHours         int          `json:"data_hours"`
	AnalyzedAt        time.Time    `json:"analyzed_at"`
}

// HourBucket aggregates metrics for a specific hour of week (day + hour)
type HourBucket struct {
	DayOfWeek  time.Weekday
	Hour       int
	CPUValues  []float64
	ConnValues []float64
	IOPSValues []float64
}

// AnalyzeActivityPattern analyzes metrics to find idle patterns
func AnalyzeActivityPattern(metrics []models.HourlyMetric, thresholds ActivityThresholds) *ActivityPattern {
	pattern := &ActivityPattern{
		AnalyzedAt: time.Now(),
	}

	if len(metrics) == 0 {
		return pattern
	}

	// Set instance ID from first metric
	pattern.InstanceID = metrics[0].InstanceID

	// Check if we have sufficient data (24+ hours)
	uniqueHours := make(map[time.Time]bool)
	for _, m := range metrics {
		uniqueHours[m.Hour] = true
	}
	pattern.DataHours = len(uniqueHours)
	pattern.HasSufficientData = pattern.DataHours >= thresholds.MinDataHours

	if !pattern.HasSufficientData {
		return pattern
	}

	// Build hour buckets grouped by day of week and hour
	buckets := buildHourBuckets(metrics)

	// Find idle windows
	pattern.IdleWindows = findIdleWindows(buckets, thresholds)

	return pattern
}

// buildHourBuckets groups metrics by day of week and hour
func buildHourBuckets(metrics []models.HourlyMetric) map[time.Weekday]map[int]*HourBucket {
	buckets := make(map[time.Weekday]map[int]*HourBucket)

	for _, m := range metrics {
		dow := m.Hour.Weekday()
		hour := m.Hour.Hour()

		if buckets[dow] == nil {
			buckets[dow] = make(map[int]*HourBucket)
		}
		if buckets[dow][hour] == nil {
			buckets[dow][hour] = &HourBucket{
				DayOfWeek: dow,
				Hour:      hour,
			}
		}

		bucket := buckets[dow][hour]
		switch m.MetricName {
		case models.MetricCPUUtilization:
			bucket.CPUValues = append(bucket.CPUValues, m.AvgValue)
		case models.MetricDatabaseConnections:
			bucket.ConnValues = append(bucket.ConnValues, m.AvgValue)
		case models.MetricReadIOPS, models.MetricWriteIOPS:
			bucket.IOPSValues = append(bucket.IOPSValues, m.AvgValue)
		}
	}

	return buckets
}

// findIdleWindows finds contiguous periods of low activity
func findIdleWindows(buckets map[time.Weekday]map[int]*HourBucket, thresholds ActivityThresholds) []IdleWindow {
	var windows []IdleWindow

	// Analyze each day of week separately first
	dayWindows := make(map[time.Weekday][]idleSegment)

	for dow := time.Sunday; dow <= time.Saturday; dow++ {
		hours := buckets[dow]
		if hours == nil {
			continue
		}

		// Find idle segments for this day
		segments := findIdleSegments(hours, thresholds)
		if len(segments) > 0 {
			dayWindows[dow] = segments
		}
	}

	// Group similar windows across days
	windows = groupSimilarWindows(dayWindows, thresholds)

	return windows
}

// idleSegment represents a contiguous period of low activity within a day
type idleSegment struct {
	startHour int
	endHour   int
	avgCPU    float64
	avgConns  float64
}

// findIdleSegments finds contiguous low-activity hours in a day
func findIdleSegments(hours map[int]*HourBucket, thresholds ActivityThresholds) []idleSegment {
	var segments []idleSegment

	// Track current segment
	var current *idleSegment
	var cpuSum, connSum float64
	var count int

	// Check hours 0-23, then wrap to handle overnight
	for h := 0; h < 48; h++ { // Check 48 hours to catch overnight windows
		hour := h % 24
		bucket := hours[hour]

		isIdle := false
		var cpu, conns float64

		if bucket != nil && len(bucket.CPUValues) > 0 {
			cpu = average(bucket.CPUValues)
			if len(bucket.ConnValues) > 0 {
				conns = average(bucket.ConnValues)
			}

			// Check if this hour is "idle" per CONTEXT.md thresholds
			isIdle = cpu < thresholds.CPUPercent
		}

		if isIdle {
			if current == nil {
				current = &idleSegment{startHour: hour}
				cpuSum = 0
				connSum = 0
				count = 0
			}
			current.endHour = hour
			cpuSum += cpu
			connSum += conns
			count++
		} else if current != nil {
			// End current segment
			current.avgCPU = cpuSum / float64(count)
			current.avgConns = connSum / float64(count)

			// Only keep if 8+ hours
			duration := segmentDuration(current.startHour, current.endHour)
			if duration >= thresholds.MinIdleHours {
				segments = append(segments, *current)
			}
			current = nil
		}

		// Stop after wrapping if we've gone past midnight
		if h >= 24 && current == nil {
			break
		}
	}

	// Close final segment if still open
	if current != nil {
		current.avgCPU = cpuSum / float64(count)
		current.avgConns = connSum / float64(count)
		duration := segmentDuration(current.startHour, current.endHour)
		if duration >= thresholds.MinIdleHours {
			segments = append(segments, *current)
		}
	}

	return segments
}

// segmentDuration calculates hours in a segment, handling overnight wrap
func segmentDuration(start, end int) int {
	if end >= start {
		return end - start + 1
	}
	// Overnight: e.g., 22 to 6 = (24-22) + 6 + 1 = 9 hours
	return (24 - start) + end + 1
}

// groupSimilarWindows combines similar patterns across days
func groupSimilarWindows(dayWindows map[time.Weekday][]idleSegment, thresholds ActivityThresholds) []IdleWindow {
	var windows []IdleWindow

	// Group segments by start/end hour similarity (within 1 hour tolerance)
	type windowKey struct {
		startHour int
		endHour   int
	}
	grouped := make(map[windowKey][]time.Weekday)
	avgCPU := make(map[windowKey]float64)
	avgConns := make(map[windowKey]float64)

	for dow, segments := range dayWindows {
		for _, seg := range segments {
			key := windowKey{seg.startHour, seg.endHour}

			// Find similar existing key or create new
			found := false
			for k := range grouped {
				if abs(k.startHour-seg.startHour) <= 1 && abs(k.endHour-seg.endHour) <= 1 {
					grouped[k] = append(grouped[k], dow)
					avgCPU[k] = (avgCPU[k] + seg.avgCPU) / 2
					avgConns[k] = (avgConns[k] + seg.avgConns) / 2
					found = true
					break
				}
			}
			if !found {
				grouped[key] = []time.Weekday{dow}
				avgCPU[key] = seg.avgCPU
				avgConns[key] = seg.avgConns
			}
		}
	}

	// Convert to IdleWindows, filtering by minimum day consistency
	for key, days := range grouped {
		if len(days) < thresholds.MinDaysConsistent {
			continue
		}

		// Sort days for consistent output
		sort.Slice(days, func(i, j int) bool { return days[i] < days[j] })

		window := IdleWindow{
			StartHour:   key.startHour,
			EndHour:     key.endHour,
			DaysOfWeek:  days,
			AvgCPU:      avgCPU[key],
			AvgConns:    avgConns[key],
			IsOvernight: key.endHour < key.startHour,
			Confidence:  calculateConfidence(len(days), avgCPU[key], thresholds),
		}
		windows = append(windows, window)
	}

	// Sort by confidence descending
	sort.Slice(windows, func(i, j int) bool {
		return windows[i].Confidence > windows[j].Confidence
	})

	return windows
}

// calculateConfidence computes a confidence score for an idle window
func calculateConfidence(numDays int, avgCPU float64, thresholds ActivityThresholds) float64 {
	confidence := 0.5 // Base confidence

	// Increase for more consistent days
	if numDays >= 5 {
		confidence += 0.3
	} else if numDays >= 3 {
		confidence += 0.2
	}

	// Increase for lower CPU (more clearly idle)
	if avgCPU < 0.5 {
		confidence += 0.2
	} else if avgCPU < thresholds.CPUPercent {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// Helper functions
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
