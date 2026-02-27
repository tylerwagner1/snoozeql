package scheduler

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/gorhill/cronexpr"
	"snoozeql/internal/models"
	"snoozeql/internal/provider"
	"snoozeql/internal/store"
)

// Scheduler manages database schedules
type Scheduler struct {
	store         Store
	registry      *provider.Registry
	instanceStore *store.InstanceStore
	eventStore    *store.EventStore
	lastExecuted  map[string]time.Time // "scheduleID_wake" or "scheduleID_sleep" -> last execution time
	mu            sync.Mutex
}

// Store interface for schedule persistence
type Store interface {
	GetSchedule(id string) (*models.Schedule, error)
	ListSchedules() ([]models.Schedule, error)
	CreateSchedule(schedule *models.Schedule) error
	UpdateSchedule(schedule *models.Schedule) error
	DeleteSchedule(id string) error
	GetMatchingSchedules(instance models.Instance) ([]models.Schedule, error)
}

// NewScheduler creates a new scheduler
func NewScheduler(store Store, registry *provider.Registry, instanceStore *store.InstanceStore, eventStore *store.EventStore) *Scheduler {
	return &Scheduler{
		store:         store,
		registry:      registry,
		instanceStore: instanceStore,
		eventStore:    eventStore,
		lastExecuted:  make(map[string]time.Time),
	}
}

// RunContinuous runs the scheduler evaluation on a 1-minute interval
func (s *Scheduler) RunContinuous(ctx context.Context) {
	log.Printf("Scheduler daemon starting (1-minute interval)")

	// Run immediately on startup to catch any missed schedules
	if err := s.Run(ctx); err != nil {
		log.Printf("Initial scheduler run error: %v", err)
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Scheduler daemon stopping")
			return
		case <-ticker.C:
			if err := s.Run(ctx); err != nil {
				log.Printf("Scheduler run error: %v", err)
			}
		}
	}
}

// shouldExecute checks if we should execute this schedule's action
// Returns true only if we haven't already executed in the current minute
func (s *Scheduler) shouldExecute(scheduleID, action string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := scheduleID + "_" + action
	lastExec, exists := s.lastExecuted[key]

	// If executed within the same minute, skip
	currentMinute := now.Truncate(time.Minute)
	if exists && lastExec.Truncate(time.Minute).Equal(currentMinute) {
		return false
	}

	// Mark as executed
	s.lastExecuted[key] = now
	return true
}

// Run executes all active schedules for all instances
func (s *Scheduler) Run(ctx context.Context) error {
	schedules, err := s.store.ListSchedules()
	if err != nil {
		return fmt.Errorf("failed to list schedules: %w", err)
	}

	now := time.Now()

	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		action := s.determineAction(schedule, now)
		if action == "none" {
			continue
		}

		// Check if we already executed this action for this schedule in this minute
		if !s.shouldExecute(schedule.ID, action, now) {
			continue
		}

		log.Printf("Schedule '%s' triggered action: %s", schedule.Name, action)

		// Get matching instances from database (not from provider registry)
		instances, err := s.instanceStore.ListInstances(ctx)
		if err != nil {
			log.Printf("Warning: Failed to list instances for schedule %s: %v", schedule.Name, err)
			continue
		}

		// Filter to matching instances
		var matchingInstances []models.Instance
		for _, instance := range instances {
			if matchesSelector(instance, schedule.Selectors) {
				matchingInstances = append(matchingInstances, instance)
			}
		}

		if len(matchingInstances) == 0 {
			log.Printf("Schedule '%s': No matching instances found", schedule.Name)
			continue
		}

		log.Printf("Schedule '%s': Found %d matching instances", schedule.Name, len(matchingInstances))

		for _, instance := range matchingInstances {
			// Skip if instance is already in target state
			if action == "start" && (instance.Status == "available" || instance.Status == "starting" || instance.Status == "running") {
				log.Printf("Skipping start for %s - already %s", instance.Name, instance.Status)
				continue
			}
			if action == "stop" && (instance.Status == "stopped" || instance.Status == "stopping") {
				log.Printf("Skipping stop for %s - already %s", instance.Name, instance.Status)
				continue
			}

			// Check for active override
			if hasActiveOverride(instance) {
				log.Printf("Skipping %s - active override exists", instance.Name)
				continue
			}

			// Determine new status for event logging
			var newStatus string
			var eventType string
			if action == "start" {
				newStatus = "starting"
				eventType = "wake"
			} else {
				newStatus = "stopping"
				eventType = "sleep"
			}

			// Log event BEFORE attempting action
			if s.eventStore != nil {
				event := &models.Event{
					InstanceID:     instance.ID,
					EventType:      eventType,
					TriggeredBy:    "schedule",
					PreviousStatus: instance.Status,
					NewStatus:      newStatus,
				}
				if err := s.eventStore.CreateEvent(ctx, event); err != nil {
					log.Printf("Warning: Failed to create event for %s: %v", instance.Name, err)
				}
			}

			// Execute action using ProviderName (e.g., "aws_uuid_us-west-2")
			switch action {
			case "stop":
				if err := s.registry.StopDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
					log.Printf("Failed to stop %s: %v", instance.Name, err)
				} else {
					log.Printf("Stopped %s (schedule: %s)", instance.Name, schedule.Name)
				}
			case "start":
				if err := s.registry.StartDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
					log.Printf("Failed to start %s: %v", instance.Name, err)
				} else {
					log.Printf("Started %s (schedule: %s)", instance.Name, schedule.Name)
				}
			}
		}
	}

	return nil
}

func matchesSelector(instance models.Instance, selectors []models.Selector) bool {
	if len(selectors) == 0 {
		return true
	}

	for _, selector := range selectors {
		if selectorMatchesInstance(instance, selector) {
			return true
		}
	}

	return false
}

func selectorMatchesInstance(instance models.Instance, selector models.Selector) bool {
	if selector.Name != nil {
		if !matchesMatcher(instance.Name, selector.Name) {
			return false
		}
	}

	if selector.Provider != nil {
		if instance.Provider != *selector.Provider {
			return false
		}
	}

	if selector.Region != nil {
		if !matchesMatcher(instance.Region, selector.Region) {
			return false
		}
	}

	if selector.Engine != nil {
		if !matchesMatcher(instance.Engine, selector.Engine) {
			return false
		}
	}

	if selector.Tags != nil {
		for key, matcher := range selector.Tags {
			tagValue, exists := instance.Tags[key]
			if !exists {
				return false
			}
			if !matchesMatcher(tagValue, matcher) {
				return false
			}
		}
	}

	return true
}

func matchesMatcher(value string, matcher *models.Matcher) bool {
	if matcher == nil {
		return true
	}

	pattern := matcher.Pattern

	switch models.MatchType(matcher.Type) {
	case models.MatchExact:
		return value == pattern
	case models.MatchContains:
		return contains(value, pattern)
	case models.MatchPrefix:
		return startsWith(value, pattern)
	case models.MatchSuffix:
		return endsWith(value, pattern)
	case models.MatchRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(value)
	default:
		return contains(value, pattern)
	}
}

func (s *Scheduler) determineAction(schedule models.Schedule, now time.Time) string {
	// Load timezone
	loc, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		log.Printf("Warning: Invalid timezone %s for schedule %s, using UTC", schedule.Timezone, schedule.Name)
		loc = time.UTC
	}

	// Convert now to the schedule's timezone for all comparisons
	nowInScheduleLoc := now.In(loc)
	currentMinute := nowInScheduleLoc.Truncate(time.Minute)

	// Get the minute that was the current minute when the CRON should have fired
	// We check by going back 1 minute and seeing when the CRON fired after that
	checkTime := nowInScheduleLoc.Add(-1 * time.Minute)

	// Parse wake CRON expression
	if schedule.WakeCron != "" {
		wakeCron, err := cronexpr.Parse(schedule.WakeCron)
		if err == nil {
			// Get the next fire time after checkTime
			wakeLast := wakeCron.Next(checkTime)
			wakeLastInScheduleLoc := wakeLast.In(loc)
			wakeLastMinute := wakeLastInScheduleLoc.Truncate(time.Minute)

			log.Printf("DEBUG: Schedule '%s' wake_cron=%s (in %s), now=%s, last fire minute=%s (schedule time)",
				schedule.Name, schedule.WakeCron, schedule.Timezone,
				nowInScheduleLoc.Format("15:04:05"),
				wakeLastMinute.In(loc).Format("15:04:05"))

			// Check if the last fire was in the current minute
			if wakeLastMinute.Equal(currentMinute) {
				log.Printf("DEBUG: Schedule '%s' triggered wake at %s", schedule.Name, schedule.WakeCron)
				return "start"
			}
		} else {
			log.Printf("Warning: Invalid wake_cron '%s' for schedule %s: %v", schedule.WakeCron, schedule.Name, err)
		}
	}

	// Parse sleep CRON expression
	if schedule.SleepCron != "" {
		sleepCron, err := cronexpr.Parse(schedule.SleepCron)
		if err == nil {
			// Get the next fire time after checkTime
			sleepLast := sleepCron.Next(checkTime)
			sleepLastInScheduleLoc := sleepLast.In(loc)
			sleepLastMinute := sleepLastInScheduleLoc.Truncate(time.Minute)

			log.Printf("DEBUG: Schedule '%s' sleep_cron=%s (in %s), now=%s, last fire minute=%s (schedule time)",
				schedule.Name, schedule.SleepCron, schedule.Timezone,
				nowInScheduleLoc.Format("15:04:05"),
				sleepLastMinute.In(loc).Format("15:04:05"))

			// Check if the last fire was in the current minute
			if sleepLastMinute.Equal(currentMinute) {
				log.Printf("DEBUG: Schedule '%s' triggered sleep at %s", schedule.Name, schedule.SleepCron)
				return "stop"
			}
		} else {
			log.Printf("Warning: Invalid sleep_cron '%s' for schedule %s: %v", schedule.SleepCron, schedule.Name, err)
		}
	}

	return "none"
}

func hasActiveOverride(instance models.Instance) bool {
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
