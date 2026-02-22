package scheduler

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/provider"
)

// Scheduler manages database schedules
type Scheduler struct {
	store    Store
	registry *provider.Registry
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
func NewScheduler(store Store, registry *provider.Registry) *Scheduler {
	return &Scheduler{
		store:    store,
		registry: registry,
	}
}

// Run executes all active schedules for all instances
func (s *Scheduler) Run(ctx context.Context) error {
	schedules, err := s.store.ListSchedules()
	if err != nil {
		return fmt.Errorf("failed to list schedules: %w", err)
	}

	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		instances, err := s.getMatchingInstances(ctx, schedule)
		if err != nil {
			fmt.Printf("Warning: Failed to get matching instances for schedule %s: %v\n", schedule.Name, err)
			continue
		}

		action := s.determineAction(schedule, time.Now())
		for _, instance := range instances {
			if hasActiveOverride(instance) {
				fmt.Printf("Skipping %s - active override exists\n", instance.Name)
				continue
			}

			switch action {
			case "stop":
				if err := s.registry.StopDatabase(ctx, instance.Provider, instance.ProviderID); err != nil {
					fmt.Printf("Failed to stop %s: %v\n", instance.Name, err)
				} else {
					fmt.Printf("Stopped %s\n", instance.Name)
				}
			case "start":
				if err := s.registry.StartDatabase(ctx, instance.Provider, instance.ProviderID); err != nil {
					fmt.Printf("Failed to start %s: %v\n", instance.Name, err)
				} else {
					fmt.Printf("Started %s\n", instance.Name)
				}
			}
		}
	}

	return nil
}

// getMatchingInstances finds instances that match the schedule's selectors
func (s *Scheduler) getMatchingInstances(ctx context.Context, schedule models.Schedule) ([]models.Instance, error) {
	instances, err := s.registry.ListAllDatabases(ctx)
	if err != nil {
		return nil, err
	}

	var matching []models.Instance
	for _, instance := range instances {
		if matchesSelector(instance, schedule.Selectors) {
			matching = append(matching, instance)
		}
	}

	return matching, nil
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
	// Parse times from cron expressions
	// Simplified for now - always returns stop
	return "stop"
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
