// Package provider registry for SnoozeQL
// Manages multiple cloud providers and provides a unified interface

package provider

import (
	"context"
	"fmt"
	"sync"

	"snoozeql/internal/models"
)

// Registry manages multiple providers
type Registry struct {
	mu        sync.RWMutex
	Providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		Providers: make(map[string]Provider),
	}
}

// Register registers a provider with the given name
func (r *Registry) Register(name string, provider Provider) {
	r.Providers[name] = provider
}

// Unregister removes a provider by name
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Providers, name)
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	provider, exists := r.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not registered", name)
	}
	return provider, nil
}

// ListAllDatabases lists databases from all registered providers
func (r *Registry) ListAllDatabases(ctx context.Context) ([]models.Instance, error) {
	var allInstances []models.Instance

	for providerName, provider := range r.Providers {
		instances, err := provider.ListDatabases(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list databases from %s: %w", providerName, err)
		}

		// Mark instances with their provider
		for i := range instances {
			instances[i].Provider = providerName
		}

		allInstances = append(allInstances, instances...)
	}

	return allInstances, nil
}

// StartDatabase starts a database by provider-specific ID
func (r *Registry) StartDatabase(ctx context.Context, providerName string, id string) error {
	provider, err := r.Get(providerName)
	if err != nil {
		return err
	}
	return provider.StartDatabase(ctx, id)
}

// StopDatabase stops a database by provider-specific ID
func (r *Registry) StopDatabase(ctx context.Context, providerName string, id string) error {
	provider, err := r.Get(providerName)
	if err != nil {
		return err
	}
	return provider.StopDatabase(ctx, id)
}

// GetProvider returns the provider that manages the given instance
func (r *Registry) GetProvider(instance models.Instance) (Provider, error) {
	return r.Get(instance.Provider)
}

// GetMetrics returns metrics for an instance from its provider
func (r *Registry) GetMetrics(ctx context.Context, providerName string, id string, period string) (map[string]any, error) {
	provider, err := r.Get(providerName)
	if err != nil {
		return nil, err
	}
	return provider.GetMetrics(ctx, providerName, id, period)
}
