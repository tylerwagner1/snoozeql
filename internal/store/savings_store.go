package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"snoozeql/internal/models"
)

// SavingsStore provides savings CRUD operations
type SavingsStore struct {
	db *Postgres
}

// NewSavingsStore creates a new savings store
func NewSavingsStore(db *Postgres) *SavingsStore {
	return &SavingsStore{db: db}
}

// UpsertDailySaving inserts or updates a daily saving record
// Uses ON CONFLICT on (instance_id, date) to accumulate savings for the same day
func (s *SavingsStore) UpsertDailySaving(ctx context.Context, instanceID string, date time.Time, stoppedMinutes int, estimatedSavingsCents int, hourlyRateCents int) error {
	query := `
		INSERT INTO savings (instance_id, date, stopped_minutes, estimated_savings_cents, hourly_rate_cents)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (instance_id, date) DO UPDATE SET
			stopped_minutes = savings.stopped_minutes + EXCLUDED.stopped_minutes,
			estimated_savings_cents = savings.estimated_savings_cents + EXCLUDED.estimated_savings_cents,
			hourly_rate_cents = EXCLUDED.hourly_rate_cents
		RETURNING id`
	var id string
	err := s.db.db.QueryRowContext(ctx, query, instanceID, date, stoppedMinutes, estimatedSavingsCents, hourlyRateCents).Scan(&id)
	return err
}

// GetSavingsByInstance returns savings records for a specific instance within a date range
func (s *SavingsStore) GetSavingsByInstance(ctx context.Context, instanceID string, startDate time.Time, endDate time.Time) ([]models.Saving, error) {
	query := `
		SELECT id, instance_id, date, stopped_minutes, estimated_savings_cents, hourly_rate_cents
		FROM savings
		WHERE instance_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC`

	rows, err := s.db.db.QueryContext(ctx, query, instanceID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query savings by instance: %w", err)
	}
	defer rows.Close()

	var savings []models.Saving
	for rows.Next() {
		var saving models.Saving
		var hourlyRateCents sql.NullInt64
		err := rows.Scan(
			&saving.ID,
			&saving.InstanceID,
			&saving.Date,
			&saving.StoppedMinutes,
			&saving.EstimatedSavingsCents,
			&hourlyRateCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saving: %w", err)
		}
		if hourlyRateCents.Valid {
			saving.HourlyRateCents = int(hourlyRateCents.Int64)
		}
		savings = append(savings, saving)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return savings, nil
}

// GetTotalSavings returns the total estimated savings in cents within a date range
func (s *SavingsStore) GetTotalSavings(ctx context.Context, startDate time.Time, endDate time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(estimated_savings_cents), 0)
		FROM savings
		WHERE date >= $1 AND date <= $2`

	var total int
	err := s.db.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to query total savings: %w", err)
	}

	return total, nil
}

// GetDailySavings returns daily savings breakdown for a time range
type DailySaving struct {
	Date            string
	SavingsCents    int
	StoppedMinutes  int
	HourlyRateCents sql.NullInt64
}

func (s *SavingsStore) GetDailySavings(ctx context.Context, startDate time.Time, endDate time.Time) ([]DailySaving, error) {
	query := `
		SELECT date::text, estimated_savings_cents, stopped_minutes, hourly_rate_cents
		FROM savings
		WHERE date >= $1 AND date <= $2
		ORDER BY date DESC`

	rows, err := s.db.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily savings: %w", err)
	}
	defer rows.Close()

	var savings []DailySaving
	for rows.Next() {
		var d DailySaving
		err := rows.Scan(&d.Date, &d.SavingsCents, &d.StoppedMinutes, &d.HourlyRateCents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily saving: %w", err)
		}
		savings = append(savings, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return savings, nil
}

// GetTopSavers returns instances with highest savings in a time range
type TopSaver struct {
	InstanceID      string
	Name            string
	SavingsCents    int
	StoppedMinutes  int
	StoppedHours    float64
	HourlyRateCents int
}

func (s *SavingsStore) GetTopSavers(ctx context.Context, startDate time.Time, endDate time.Time, limit int) ([]TopSaver, error) {
	query := `
		SELECT 
			s.instance_id,
			i.name,
			SUM(s.estimated_savings_cents) as total_savings,
			SUM(s.stopped_minutes) as total_stopped_minutes,
			AVG(s.hourly_rate_cents)::int as avg_hourly_rate
		FROM savings s
		JOIN instances i ON s.instance_id = i.id
		WHERE s.date >= $1 AND s.date <= $2
		GROUP BY s.instance_id, i.name
		ORDER BY total_savings DESC
		LIMIT $3`

	rows, err := s.db.db.QueryContext(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top savers: %w", err)
	}
	defer rows.Close()

	var savers []TopSaver
	for rows.Next() {
		var saver TopSaver
		err := rows.Scan(&saver.InstanceID, &saver.Name, &saver.SavingsCents, &saver.StoppedMinutes, &saver.HourlyRateCents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top saver: %w", err)
		}
		saver.StoppedHours = float64(saver.StoppedMinutes) / 60.0
		savers = append(savers, saver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return savers, nil
}

// RefreshSavingsSummary refreshes the materialized view for updated aggregations
func (s *SavingsStore) RefreshSavingsSummary(ctx context.Context) error {
	_, err := s.db.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW savings_summary")
	if err != nil {
		return fmt.Errorf("failed to refresh savings summary: %w", err)
	}
	return nil
}
