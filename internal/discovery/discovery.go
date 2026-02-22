package discovery

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/provider"
	"snoozeql/internal/store"
)

// CloudAccountProvider represents a cloud account with its provider
type CloudAccountProvider struct {
	Name        string
	Provider    string
	Credentials map[string]string
	Regions     []string
}

// DiscoveryService manages database instance discovery
type DiscoveryService struct {
	registry      *provider.Registry
	enabled       bool
	interval      time.Duration
	tags          []string
	lastSync      time.Time
	lastError     error
	instanceStore *store.InstanceStore
	accountStore  *store.CloudAccountStore
	eventStore    *store.EventStore
	mu            sync.RWMutex
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(registry *provider.Registry, instanceStore *store.InstanceStore, accountStore *store.CloudAccountStore, eventStore *store.EventStore, enabled bool, interval int, tags []string) *DiscoveryService {
	return &DiscoveryService{
		registry:      registry,
		enabled:       enabled,
		interval:      time.Duration(interval) * time.Second,
		tags:          tags,
		instanceStore: instanceStore,
		accountStore:  accountStore,
		eventStore:    eventStore,
	}
}

// IsEnabled returns whether discovery is enabled
func (d *DiscoveryService) IsEnabled() bool {
	return d.enabled
}

// GetLastSync returns the last sync time
func (d *DiscoveryService) GetLastSync() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastSync
}

// GetLastError returns the last error
func (d *DiscoveryService) GetLastError() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastError
}

// RunOnce triggers a single discovery run (for manual refresh)
func (d *DiscoveryService) RunOnce(ctx context.Context) error {
	return d.Run(ctx)
}

// ListAllDatabases lists all databases from all providers
func (d *DiscoveryService) ListAllDatabases(ctx context.Context) ([]models.Instance, error) {
	return d.registry.ListAllDatabases(ctx)
}

// Run runs the discovery process - fetches instances from all providers
func (d *DiscoveryService) Run(ctx context.Context) error {
	if !d.enabled {
		return fmt.Errorf("discovery is not enabled")
	}

	d.mu.Lock()
	d.lastError = nil
	d.mu.Unlock()

	// Load accounts for ID mapping
	var accountsByID map[string]*models.CloudAccount
	if d.accountStore != nil {
		allAccounts, err := d.accountStore.ListCloudAccounts()
		if err == nil {
			accountsByID = make(map[string]*models.CloudAccount)
			for _, account := range allAccounts {
				accountsByID[account.ID] = &account
			}
		}
	}

	// Update all accounts to "syncing" status before discovery
	if d.accountStore != nil {
		accounts, err := d.accountStore.ListCloudAccounts()
		if err == nil {
			for _, account := range accounts {
				if err := d.accountStore.UpdateConnectionStatus(ctx, account.ID, "syncing", nil); err != nil {
					fmt.Printf("Warning: Failed to update account %s to syncing status: %v\n", account.ID, err)
				}
			}
		} else {
			fmt.Printf("Warning: Failed to list accounts for status update: %v\n", err)
		}
	}

	instances, err := d.registry.ListAllDatabases(ctx)
	if err != nil {
		d.mu.Lock()
		d.lastError = err
		d.mu.Unlock()

		// Update accounts to "failed" status
		if d.accountStore != nil {
			accounts, _ := d.accountStore.ListCloudAccounts()
			for _, account := range accounts {
				errStr := err.Error()
				if err := d.accountStore.UpdateConnectionStatus(ctx, account.ID, "failed", &errStr); err != nil {
					fmt.Printf("Warning: Failed to update account %s to failed status: %v\n", account.ID, err)
				}
			}
		}

		return fmt.Errorf("failed to list databases: %w", err)
	}

	// Set CloudAccountID for each instance based on provider name
	// Provider format: aws_{accountID}_{region} or gcp_{accountID}
	if accountsByID != nil {
		for i := range instances {
			instance := &instances[i]
			// Use AccountID that was set by ProviderRegistry
			accountID := instance.AccountID
			if accountID != "" {
				if account, exists := accountsByID[accountID]; exists {
					instance.CloudAccountID = account.ID
				}
			}
			// Clear AccountID as it's only used for mapping
			instance.AccountID = ""
			fmt.Printf("DEBUG: Provider=%s, ProviderName=%s, Instance=%s, ProviderID=%s, Status=%s, CloudAccountID=%s\n", instance.Provider, instance.ProviderName, instance.Name, instance.ProviderID, instance.Status, instance.CloudAccountID)
		}
	}

	// Sync instances to database
	syncCount := 0
	var syncErrors []error
	if d.instanceStore != nil {
		for _, instance := range instances {
			fmt.Printf("DEBUG: Syncing instance: Name=%s, Provider=%s, ProviderName=%s, ProviderID=%s, Status=%s\n", instance.Name, instance.Provider, instance.ProviderName, instance.ProviderID, instance.Status)
			if err := d.instanceStore.UpsertInstance(ctx, &instance); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("failed to sync instance %s (%s): %w", instance.Name, instance.ProviderID, err))
				fmt.Printf("DEBUG: Failed to sync instance %s: %v\n", instance.Name, err)
			} else {
				syncCount++
			}
		}
	}

	// Update accounts to "connected" status
	if d.accountStore != nil {
		accounts, _ := d.accountStore.ListCloudAccounts()
		now := time.Now()
		for _, account := range accounts {
			if err := d.accountStore.UpdateConnectionStatus(ctx, account.ID, "connected", nil); err != nil {
				fmt.Printf("Warning: Failed to update account %s to connected status: %v\n", account.ID, err)
			}
			account.LastSyncAt = &now
		}
	}

	// Log sync results
	fmt.Printf("Synced %d instances to database", syncCount)
	if len(syncErrors) > 0 {
		fmt.Printf(" with %d errors", len(syncErrors))
	}
	fmt.Println()

	d.mu.Lock()
	d.lastSync = time.Now()
	d.mu.Unlock()

	return nil
}

// RunContinuous runs discovery on the configured interval (for background goroutine)
func (d *DiscoveryService) RunContinuous(ctx context.Context) {
	if !d.enabled {
		return
	}

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.Run(ctx); err != nil {
				fmt.Printf("Discovery run failed: %v\n", err)
			}
		}
	}
}

// StartDatabase starts a database instance
func (d *DiscoveryService) StartDatabase(ctx context.Context, providerName string, id string) error {
	provider, err := d.registry.Get(providerName)
	if err != nil {
		return err
	}

	// Try to get current status and UUID before starting
	instance, getErr := d.instanceStore.GetInstanceByProviderID(ctx, "", id)
	var prevStatus string
	var instanceUUID string
	if getErr == nil {
		prevStatus = instance.Status
		instanceUUID = instance.ID
	}

	// Log start event BEFORE attempting AWS call (so it always happens)
	if d.eventStore != nil && instanceUUID != "" {
		err := d.eventStore.CreateEvent(ctx, &models.Event{
			InstanceID:     instanceUUID,
			EventType:      "start",
			PreviousStatus: prevStatus,
			NewStatus:      "starting",
			TriggeredBy:    "manual",
		})
		if err != nil {
			log.Printf("DEBUG: Failed to create start event for %s: %v", id, err)
		} else {
			log.Printf("DEBUG: Created start event for %s (UUID: %s)", id, instanceUUID)
		}
	} else if d.eventStore == nil {
		log.Printf("DEBUG: eventStore is nil, cannot create start event for %s", id)
	} else {
		log.Printf("DEBUG: No instance UUID found for %s, cannot create start event", id)
	}

	err = provider.StartDatabase(ctx, id)
	return err
}

// StopDatabase stops a database instance
func (d *DiscoveryService) StopDatabase(ctx context.Context, providerName string, id string) error {
	provider, err := d.registry.Get(providerName)
	if err != nil {
		return err
	}

	// Try to get current status and UUID before stopping
	instance, getErr := d.instanceStore.GetInstanceByProviderID(ctx, "", id)
	var prevStatus string
	var instanceUUID string
	if getErr == nil {
		prevStatus = instance.Status
		instanceUUID = instance.ID
	}

	// Log stop event BEFORE attempting AWS call (so it always happens)
	if d.eventStore != nil && instanceUUID != "" {
		err := d.eventStore.CreateEvent(ctx, &models.Event{
			InstanceID:     instanceUUID,
			EventType:      "stop",
			PreviousStatus: prevStatus,
			NewStatus:      "stopping",
			TriggeredBy:    "manual",
		})
		if err != nil {
			log.Printf("DEBUG: Failed to create stop event for %s: %v", id, err)
		} else {
			log.Printf("DEBUG: Created stop event for %s (UUID: %s)", id, instanceUUID)
		}
	} else if d.eventStore == nil {
		log.Printf("DEBUG: eventStore is nil, cannot create stop event for %s", id)
	} else {
		log.Printf("DEBUG: No instance UUID found for %s, cannot create stop event", id)
	}

	err = provider.StopDatabase(ctx, id)
	return err
}

// SyncInstance syncs a single instance to the database
func (d *DiscoveryService) SyncInstance(ctx context.Context, instance models.Instance) error {
	// TODO: Implement database sync in Phase 3
	_ = ctx
	fmt.Printf("Syncing instance: %s (%s)\n", instance.Name, instance.Provider)
	return nil
}

// GetManagedTags returns the configured managed tags
func (d *DiscoveryService) GetManagedTags() []string {
	return d.tags
}

// ProviderRegistry implements the Provider interface for use with Registry
type ProviderRegistry struct {
	providers map[string]provider.Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]provider.Provider),
	}
}

// Register registers a provider
func (r *ProviderRegistry) Register(name string, p provider.Provider) {
	r.providers[name] = p
}

// Get retrieves a provider by name
func (r *ProviderRegistry) Get(name string) (provider.Provider, error) {
	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not registered", name)
	}
	return p, nil
}

// ListAllDatabases lists databases from all providers
func (r *ProviderRegistry) ListAllDatabases(ctx context.Context) ([]models.Instance, error) {
	var allInstances []models.Instance

	for name, p := range r.providers {
		instances, err := p.ListDatabases(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list from %s: %w", name, err)
		}
		// Extract cloud provider type from provider name (e.g., "aws_f74f8e11-42e4-4bce-b85e-6d0cbb3e86ea_us-east-1" -> "aws")
		cloudProvider := ""
		parts := strings.Split(name, "_")
		if len(parts) >= 1 {
			cloudProvider = parts[0]
		}
		for i := range instances {
			instances[i].Provider = cloudProvider
			instances[i].ProviderName = name
			fmt.Printf("DEBUG: Set Provider=%s, ProviderName=%s for instance %s\n", cloudProvider, name, instances[i].Name)
		}
		// Verify the values were set
		allInstances = append(allInstances, instances...)
	}

	return allInstances, nil
}

// StartDatabase starts a database
func (r *ProviderRegistry) StartDatabase(ctx context.Context, providerName string, id string) error {
	p, err := r.Get(providerName)
	if err != nil {
		return err
	}
	return p.StartDatabase(ctx, id)
}

// StopDatabase stops a database
func (r *ProviderRegistry) StopDatabase(ctx context.Context, providerName string, id string) error {
	p, err := r.Get(providerName)
	if err != nil {
		return err
	}
	return p.StopDatabase(ctx, id)
}
