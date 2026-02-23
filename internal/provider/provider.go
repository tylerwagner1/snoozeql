// Package provider provides interfaces for cloud database providers

package provider

import (
	"context"

	"snoozeql/internal/models"
)

// Provider is the interface that all cloud providers must implement
type Provider interface {
	// ListDatabases returns all databases in the configured regions
	ListDatabases(ctx context.Context) ([]models.Instance, error)

	// StartDatabase starts a stopped database
	StartDatabase(ctx context.Context, id string) error

	// StopDatabase stops a running database
	StopDatabase(ctx context.Context, id string) error

	// GetDatabaseStatus returns the current status of a database
	GetDatabaseStatus(ctx context.Context, id string) (string, error)

	// GetMetrics returns activity metrics for a database
	GetMetrics(ctx context.Context, providerName string, id string, period string) (map[string]any, error)

	// GetDatabaseByID returns a database by its ID
	GetDatabaseByID(ctx context.Context, id string) (*models.Instance, error)
}
