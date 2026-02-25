package metrics

import (
	"context"
	"log"
	"time"

	"snoozeql/internal/store"
)

const (
	retentionDays    = 7
	cleanupBatchSize = 1000
	startupDelay     = 7 * time.Minute
	cleanupInterval  = 24 * time.Hour
	settingsKey      = "metrics_retention_last_run"
)

// RetentionCleaner manages automatic cleanup of old metrics
type RetentionCleaner struct {
	metricsStore *MetricsStore
	db           *store.Postgres
}

// NewRetentionCleaner creates a new retention cleaner
func NewRetentionCleaner(metricsStore *MetricsStore, db *store.Postgres) *RetentionCleaner {
	return &RetentionCleaner{
		metricsStore: metricsStore,
		db:           db,
	}
}

// RunContinuous runs the retention cleanup on the configured interval
func (r *RetentionCleaner) RunContinuous(ctx context.Context) {
	// Wait for startup delay with context awareness
	select {
	case <-ctx.Done():
		log.Println("Retention cleaner shutting down before startup")
		return
	case <-time.After(startupDelay):
		// Continue to cleanup
	}

	// Run immediately after delay
	if err := r.runCleanup(ctx); err != nil {
		log.Printf("Retention cleanup failed: %v", err)
	}

	// Then every 24 hours
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Retention cleaner shutting down")
			return
		case <-ticker.C:
			if err := r.runCleanup(ctx); err != nil {
				log.Printf("Retention cleanup failed: %v", err)
			}
		}
	}
}

// runCleanup performs a single cleanup cycle
func (r *RetentionCleaner) runCleanup(ctx context.Context) error {
	// Check if we already ran within 24 hours
	lastRun, err := r.getLastRunTime(ctx)
	if err == nil && time.Since(lastRun) < cleanupInterval {
		return nil // Skip - already ran recently
	}

	// Calculate cutoff time (7 days ago in UTC)
	cutoff := time.Now().UTC().Add(-retentionDays * 24 * time.Hour)

	// Perform batched deletion
	if err := r.deleteInBatches(ctx, cutoff); err != nil {
		return err
	}

	// Update last run timestamp on success
	return r.setLastRunTime(ctx, time.Now())
}

// deleteInBatches deletes old metrics in batches to avoid table locking
func (r *RetentionCleaner) deleteInBatches(ctx context.Context, cutoff time.Time) error {
	for {
		// Delete in batches using subquery with LIMIT
		deleted, err := r.metricsStore.DeleteOldMetrics(ctx, cutoff, cleanupBatchSize)
		if err != nil {
			return err
		}

		if deleted < int64(cleanupBatchSize) {
			break // No more rows to delete
		}

		// Brief pause between batches with context check
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return nil
}

// getLastRunTime retrieves the last run timestamp from settings
func (r *RetentionCleaner) getLastRunTime(ctx context.Context) (time.Time, error) {
	var timestampStr string
	query := `SELECT value->>'timestamp' FROM settings WHERE key = $1`
	err := r.db.QueryRow(ctx, query, settingsKey).Scan(&timestampStr)
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

// setLastRunTime stores the last run timestamp in settings
func (r *RetentionCleaner) setLastRunTime(ctx context.Context, t time.Time) error {
	query := `
		INSERT INTO settings (key, value, scope, updated_at)
		VALUES ($1, jsonb_build_object('timestamp', $2::text), 'global', NOW())
		ON CONFLICT (key) DO UPDATE SET
			value = jsonb_build_object('timestamp', $2::text),
			updated_at = NOW()`
	_, err := r.db.Exec(ctx, query, settingsKey, t.UTC().Format(time.RFC3339))
	return err
}
