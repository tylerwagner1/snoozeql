package savings

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/store"
)

// EventCreator interface matches existing EventStore methods
type EventCreator interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	ListEventsByInstance(ctx context.Context, instanceID string) ([]models.Event, error)
}

// SavingsStorer interface for persisting savings records
type SavingsStorer interface {
	UpsertDailySaving(ctx context.Context, instanceID string, date time.Time, stoppedMinutes int, estimatedSavingsCents int, hourlyRateCents int) error
}

// InstanceGetter interface for getting instance information
type InstanceGetter interface {
	GetInstanceByID(ctx context.Context, id string) (*models.Instance, error)
}

// EventStoreWithSavings decorates EventStore with automatic savings calculation
type EventStoreWithSavings struct {
	wrapped       EventCreator
	calculator    *SavingsCalculator
	savingsStore  SavingsStorer
	instanceStore InstanceGetter
	db            *store.Postgres
}

// NewEventStoreWithSavings creates a new decorated EventStore
func NewEventStoreWithSavings(
	wrapped EventCreator,
	calculator *SavingsCalculator,
	savingsStore SavingsStorer,
	instanceStore InstanceGetter,
	db *store.Postgres,
) *EventStoreWithSavings {
	return &EventStoreWithSavings{
		wrapped:       wrapped,
		calculator:    calculator,
		savingsStore:  savingsStore,
		instanceStore: instanceStore,
		db:            db,
	}
}

// CreateEvent intercepts event creation for savings calculation
// - For stop/sleep events: captures hourly_rate_cents in metadata
// - For start/wake events: calculates and persists savings
func (e *EventStoreWithSavings) CreateEvent(ctx context.Context, event *models.Event) error {
	// Always call wrapped.CreateEvent first (preserve existing behavior)
	if err := e.wrapped.CreateEvent(ctx, event); err != nil {
		return err
	}

	// Handle stop/sleep events - capture hourly rate in metadata
	if event.EventType == "sleep" || event.EventType == "stop" {
		// Get instance to capture hourly_rate_cents
		if instance, err := e.instanceStore.GetInstanceByID(ctx, event.InstanceID); err == nil {
			// Store hourly_rate_cents in event metadata
			metadata := map[string]interface{}{
				"hourly_rate_cents": instance.HourlyCostCents,
			}
			if metadataBytes, marshalErr := json.Marshal(metadata); marshalErr == nil {
				// Update the event's metadata in the database
				if updateErr := e.updateEventMetadata(ctx, event.ID, metadataBytes); updateErr != nil {
					log.Printf("Warning: Failed to update event metadata for instance %s: %v", event.InstanceID, updateErr)
				}
			} else {
				log.Printf("Warning: Failed to marshal metadata for instance %s: %v", event.InstanceID, marshalErr)
			}
		} else {
			log.Printf("Warning: Failed to get instance %s for stop event: %v", event.InstanceID, err)
		}
	}

	// Handle start/wake events - calculate and persist savings
	if event.EventType == "wake" || event.EventType == "start" {
		e.calculateAndPersistSavings(ctx, event)
	}

	return nil
}

// updateEventMetadata updates the metadata for an existing event
func (e *EventStoreWithSavings) updateEventMetadata(ctx context.Context, eventID string, metadata []byte) error {
	query := `UPDATE events SET metadata = $1 WHERE id = $2`
	_, err := e.db.Exec(ctx, query, metadata, eventID)
	return err
}

// calculateAndPersistSavings calculates savings from a stop/start event pair
func (e *EventStoreWithSavings) calculateAndPersistSavings(ctx context.Context, startEvent *models.Event) {
	// Find the most recent stop/sleep event for this instance
	stopEvents, err := e.wrapped.ListEventsByInstance(ctx, startEvent.InstanceID)
	if err != nil {
		log.Printf("Warning: Failed to list events for instance %s: %v", startEvent.InstanceID, err)
		return
	}

	// Find the most recent stop/sleep event before this start event
	var stopEvent *models.Event
	// Iterate in reverse to find the most recent stop event (latest event, first in reverse)
	for i := len(stopEvents) - 1; i >= 0; i-- {
		if stopEvents[i].EventType == "sleep" || stopEvents[i].EventType == "stop" {
			stopEvent = &stopEvents[i]
			break
		}
	}

	if stopEvent == nil {
		// No previous stop event found - instance was never stopped by SnoozeQL
		log.Printf("Info: No previous stop event found for instance %s, skipping savings calculation", startEvent.InstanceID)
		return
	}

	// Extract hourly_rate_cents from stop event metadata (or fallback to instance.HourlyCostCents)
	hourlyRateCents := 0

	// Try to get from stop event metadata
	if len(stopEvent.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(stopEvent.Metadata, &metadata); err == nil {
			if rate, ok := metadata["hourly_rate_cents"].(float64); ok {
				hourlyRateCents = int(rate)
			}
		}
	}

	// Fallback to instance.HourlyCostCents if not in metadata
	if hourlyRateCents == 0 {
		if instance, err := e.instanceStore.GetInstanceByID(ctx, startEvent.InstanceID); err == nil {
			hourlyRateCents = instance.HourlyCostCents
		}
	}

	// Skip calculation if we still couldn't get the rate
	if hourlyRateCents == 0 {
		log.Printf("Warning: Could not determine hourly rate for instance %s, skipping savings calculation", startEvent.InstanceID)
		return
	}

	// Use calculator.SplitByDay to get per-day savings
	stoppedMinutes, _ := e.calculator.CalculateSavings(
		stopEvent.CreatedAt,
		startEvent.CreatedAt,
		hourlyRateCents,
	)

	// Log calculation details (instance_id, date, stopped_minutes, rate) - satisfies AUD-01
	log.Printf(
		"AUDIT: Savings calculation - instance_id=%s, date=%s, stopped_minutes=%d, rate=%d",
		startEvent.InstanceID,
		startEvent.CreatedAt.Format("2006-01-02"),
		stoppedMinutes,
		hourlyRateCents,
	)

	// For each day, call savingsStore.UpsertDailySaving
	dailySavings := e.calculator.SplitByDay(stopEvent.CreatedAt, startEvent.CreatedAt, hourlyRateCents)
	for _, dailySaving := range dailySavings {
		if err := e.savingsStore.UpsertDailySaving(
			ctx,
			startEvent.InstanceID,
			dailySaving.Date,
			dailySaving.StoppedMinutes,
			dailySaving.SavingsCents,
			dailySaving.HourlyRateCents,
		); err != nil {
			log.Printf("Warning: Failed to persist daily saving for instance %s on %s: %v",
				startEvent.InstanceID, dailySaving.Date.Format("2006-01-02"), err)
		}
	}
}

// ListEventsByInstance delegates to wrapped
func (e *EventStoreWithSavings) ListEventsByInstance(ctx context.Context, instanceID string) ([]models.Event, error) {
	return e.wrapped.ListEventsByInstance(ctx, instanceID)
}
