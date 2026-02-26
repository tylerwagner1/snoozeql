package metrics

import (
	"context"
	"fmt"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/store"
)

// MetricsStore provides metrics CRUD operations
type MetricsStore struct {
	db *store.Postgres
}

// MetricPeriod constant for 5-minute granularity
const MetricPeriod = 5 * time.Minute

// TruncateToMetricPeriod truncates a time to 5-minute boundaries
func TruncateToMetricPeriod(t time.Time) time.Time {
	return t.Truncate(MetricPeriod)
}

// NewMetricsStore creates a new metrics store
func NewMetricsStore(db *store.Postgres) *MetricsStore {
	return &MetricsStore{db: db}
}

// UpsertHourlyMetric inserts or updates an hourly metric aggregate
// Uses incremental averaging for existing hour buckets
func (s *MetricsStore) UpsertHourlyMetric(ctx context.Context, m *models.HourlyMetric) error {
	query := `
        INSERT INTO metrics_hourly (instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count)
        VALUES ($1, $2, date_trunc('hour', $3::timestamptz), $4, $5, $6, $7)
        ON CONFLICT (instance_id, metric_name, hour) DO UPDATE SET
            avg_value = (metrics_hourly.avg_value * metrics_hourly.sample_count + EXCLUDED.avg_value) 
                        / (metrics_hourly.sample_count + 1),
            max_value = GREATEST(metrics_hourly.max_value, EXCLUDED.max_value),
            min_value = LEAST(metrics_hourly.min_value, EXCLUDED.min_value),
            sample_count = metrics_hourly.sample_count + 1,
            updated_at = NOW()
        RETURNING id`

	return s.db.QueryRow(ctx, query,
		m.InstanceID, m.MetricName, m.Hour,
		m.AvgValue, m.MaxValue, m.MinValue, m.SampleCount,
	).Scan(&m.ID)
}

// GetMetricsByInstance returns metrics for an instance within a time range
func (s *MetricsStore) GetMetricsByInstance(ctx context.Context, instanceID string, start, end time.Time) ([]models.HourlyMetric, error) {
	query := `
        SELECT id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count, created_at, updated_at
        FROM metrics_hourly
        WHERE instance_id = $1 AND hour >= $2 AND hour <= $3
        ORDER BY hour ASC, metric_name ASC`

	rows, err := s.db.Query(ctx, query, instanceID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []models.HourlyMetric
	for rows.Next() {
		var m models.HourlyMetric
		err := rows.Scan(
			&m.ID, &m.InstanceID, &m.MetricName, &m.Hour,
			&m.AvgValue, &m.MaxValue, &m.MinValue, &m.SampleCount,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return metrics, nil
}

// GetLatestMetrics returns the most recent metrics for an instance
func (s *MetricsStore) GetLatestMetrics(ctx context.Context, instanceID string) ([]models.HourlyMetric, error) {
	query := `
        SELECT DISTINCT ON (metric_name) 
            id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count, created_at, updated_at
        FROM metrics_hourly
        WHERE instance_id = $1
        ORDER BY metric_name, hour DESC`

	rows, err := s.db.Query(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest metrics: %w", err)
	}
	defer rows.Close()

	metrics := make([]models.HourlyMetric, 0)
	for rows.Next() {
		var m models.HourlyMetric
		err := rows.Scan(
			&m.ID, &m.InstanceID, &m.MetricName, &m.Hour,
			&m.AvgValue, &m.MaxValue, &m.MinValue, &m.SampleCount,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// DeleteOldMetrics removes metrics older than the retention period
// Called by a cleanup job to maintain 7-day retention
// Supports batched deletes with limit parameter to avoid table locking
func (s *MetricsStore) DeleteOldMetrics(ctx context.Context, before time.Time, limit int) (int64, error) {
	// Use subquery with LIMIT to batch deletes and avoid table locks
	query := `
		DELETE FROM metrics_hourly 
		WHERE id IN (
			SELECT id FROM metrics_hourly 
			WHERE hour < $1 
			LIMIT $2
		)`
	return s.db.Exec(ctx, query, before, limit)
}

// HasSufficientData checks if an instance has enough data for pattern analysis
// Returns true if there are at least 24 hours of data (per CONTEXT.md requirement)
func (s *MetricsStore) HasSufficientData(ctx context.Context, instanceID string) (bool, error) {
	query := `
        SELECT COUNT(DISTINCT hour) >= 24
        FROM metrics_hourly
        WHERE instance_id = $1`

	var sufficient bool
	err := s.db.QueryRow(ctx, query, instanceID).Scan(&sufficient)
	if err != nil {
		return false, fmt.Errorf("failed to check data sufficiency for instance %s: %w", instanceID, err)
	}
	return sufficient, nil
}

// HourHasData checks if an instance has any metrics for a specific time period
// The hour parameter should be pre-truncated to 5-minute boundaries using TruncateToMetricPeriod
func (s *MetricsStore) HourHasData(ctx context.Context, instanceID, metricName string, hour time.Time) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 FROM metrics_hourly
            WHERE instance_id = $1 AND metric_name = $2 AND hour = date_trunc('hour', $3::timestamptz)
        )`

	var exists bool
	err := s.db.QueryRow(ctx, query, instanceID, metricName, hour).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check hour data for instance %s: %w", instanceID, err)
	}
	return exists, nil
}

// GetLatestMetricTimes returns the most recent metric timestamp for each instance
// Returns map[instanceID]time.Time - batch query to avoid N queries for N instances
func (s *MetricsStore) GetLatestMetricTimes(ctx context.Context) (map[string]time.Time, error) {
	query := `
        SELECT instance_id, MAX(hour) as latest
        FROM metrics_hourly
        GROUP BY instance_id`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest metric times: %w", err)
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var instanceID string
		var latestTime time.Time
		err := rows.Scan(&instanceID, &latestTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[instanceID] = latestTime
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}

// GetMetricsAtTime returns metrics for an instance at a specific timestamp
// Used to get boundary values for interpolation
func (s *MetricsStore) GetMetricsAtTime(ctx context.Context, instanceID string, timestamp time.Time) ([]models.HourlyMetric, error) {
	query := `
        SELECT id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count, created_at, updated_at
        FROM metrics_hourly
        WHERE instance_id = $1 AND hour = date_trunc('hour', $2::timestamptz)`

	rows, err := s.db.Query(ctx, query, instanceID, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics at time: %w", err)
	}
	defer rows.Close()

	var metrics []models.HourlyMetric
	for rows.Next() {
		var m models.HourlyMetric
		err := rows.Scan(
			&m.ID, &m.InstanceID, &m.MetricName, &m.Hour,
			&m.AvgValue, &m.MaxValue, &m.MinValue, &m.SampleCount,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return metrics, nil
}
