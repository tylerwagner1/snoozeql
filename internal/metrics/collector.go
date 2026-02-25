package metrics

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"snoozeql/internal/models"
	"snoozeql/internal/store"
)

// MetricsCollector manages background metric collection from CloudWatch
type MetricsCollector struct {
	metricsStore  *MetricsStore
	instanceStore *store.InstanceStore
	accountStore  *store.CloudAccountStore
	interval      time.Duration
	clients       map[string]*CloudWatchClient // accountID_region -> client
	clientsMu     sync.RWMutex
	enabled       bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(
	metricsStore *MetricsStore,
	instanceStore *store.InstanceStore,
	accountStore *store.CloudAccountStore,
	intervalMinutes int,
) *MetricsCollector {
	return &MetricsCollector{
		metricsStore:  metricsStore,
		instanceStore: instanceStore,
		accountStore:  accountStore,
		interval:      time.Duration(intervalMinutes) * time.Minute,
		clients:       make(map[string]*CloudWatchClient),
		enabled:       true,
	}
}

// RunContinuous runs the metrics collection on the configured interval
func (c *MetricsCollector) RunContinuous(ctx context.Context) {
	if !c.enabled {
		return
	}

	// Run immediately on startup
	if err := c.CollectAll(ctx); err != nil {
		log.Printf("Initial metrics collection failed: %v", err)
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Metrics collector shutting down")
			return
		case <-ticker.C:
			// Recover from panics to prevent goroutine from dying
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Metrics collection panic recovered: %v", r)
					}
				}()

				if err := c.CollectAll(ctx); err != nil {
					log.Printf("Metrics collection failed: %v", err)
				}
			}()
		}
	}
}

// CollectAll collects metrics for all running instances
func (c *MetricsCollector) CollectAll(ctx context.Context) error {
	log.Println("Starting metrics collection cycle...")

	instances, err := c.instanceStore.ListInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	var collected, skipped, failed int

	for _, instance := range instances {
		// Store zeros for stopped instances (shows "asleep" state in metrics) for all providers
		if instance.Status != "available" && instance.Status != "running" {
			if err := c.storeZeroMetrics(ctx, instance); err != nil {
				log.Printf("Failed to store zero metrics for stopped instance %s: %v", instance.Name, err)
				failed++
				continue
			}
			collected++
			continue
		}

		// Skip non-AWS instances for active collection (GCP metrics not yet implemented)
		if instance.Provider != "aws" {
			log.Printf("Skipping active metrics collection for non-AWS instance %s (provider: %s)", instance.Name, instance.Provider)
			skipped++
			continue
		}

		// Get or create CloudWatch client for this account/region
		client, err := c.getClient(ctx, instance)
		if err != nil {
			log.Printf("Failed to get CloudWatch client for %s: %v", instance.Name, err)
			failed++
			continue
		}

		// Collect metrics for this instance
		if err := c.collectInstance(ctx, client, instance); err != nil {
			log.Printf("Failed to collect metrics for %s: %v", instance.Name, err)
			failed++
			continue
		}

		collected++
	}

	log.Printf("Metrics collection complete: collected=%d, skipped=%d, failed=%d", collected, skipped, failed)
	return nil
}

// collectInstance collects and stores metrics for a single instance
func (c *MetricsCollector) collectInstance(ctx context.Context, client *CloudWatchClient, instance models.Instance) error {
	// ProviderID is the DBInstanceIdentifier for RDS
	metrics, err := client.GetRDSMetrics(ctx, instance.ProviderID)
	if err != nil {
		return fmt.Errorf("GetRDSMetrics failed: %w", err)
	}

	// Store each metric type
	if metrics.CPU != nil {
		if err := c.storeMetric(ctx, instance.ID, models.MetricCPUUtilization, metrics.Timestamp, metrics.CPU); err != nil {
			log.Printf("Failed to store CPU metric for %s: %v", instance.Name, err)
		}
	}

	if metrics.Connections != nil {
		if err := c.storeMetric(ctx, instance.ID, models.MetricDatabaseConnections, metrics.Timestamp, metrics.Connections); err != nil {
			log.Printf("Failed to store Connections metric for %s: %v", instance.Name, err)
		}
	}

	if metrics.ReadIOPS != nil {
		if err := c.storeMetric(ctx, instance.ID, models.MetricReadIOPS, metrics.Timestamp, metrics.ReadIOPS); err != nil {
			log.Printf("Failed to store ReadIOPS metric for %s: %v", instance.Name, err)
		}
	}

	if metrics.WriteIOPS != nil {
		if err := c.storeMetric(ctx, instance.ID, models.MetricWriteIOPS, metrics.Timestamp, metrics.WriteIOPS); err != nil {
			log.Printf("Failed to store WriteIOPS metric for %s: %v", instance.Name, err)
		}
	}

	if metrics.FreeMemory != nil {
		// Calculate memory percentage from bytes
		pct := CalculateMemoryPercentage(instance.InstanceType, metrics.FreeMemory.Avg)
		if pct != nil {
			memValue := &MetricValue{Avg: *pct, Max: *pct, Min: *pct}
			if err := c.storeMetric(ctx, instance.ID, models.MetricFreeableMemory, metrics.Timestamp, memValue); err != nil {
				log.Printf("Failed to store FreeableMemory metric for %s: %v", instance.Name, err)
			}
		} else {
			log.Printf("Unknown instance class %s for %s - skipping memory percentage", instance.InstanceType, instance.Name)
		}
	}

	return nil
}

// storeMetric stores a single metric value
func (c *MetricsCollector) storeMetric(ctx context.Context, instanceID, metricName string, hour time.Time, value *MetricValue) error {
	m := &models.HourlyMetric{
		InstanceID:  instanceID,
		MetricName:  metricName,
		Hour:        hour,
		AvgValue:    value.Avg,
		MaxValue:    value.Max,
		MinValue:    value.Min,
		SampleCount: 1,
	}
	return c.metricsStore.UpsertHourlyMetric(ctx, m)
}

// storeZeroMetrics stores zero metrics for all metric types
// Used for stopped instances to show "asleep" state
func (c *MetricsCollector) storeZeroMetrics(ctx context.Context, instance models.Instance) error {
	zeroValue := &MetricValue{Avg: 0, Max: 0, Min: 0}
	timestamp := time.Now().UTC().Truncate(time.Hour)

	for _, metricName := range []string{
		models.MetricCPUUtilization,
		models.MetricDatabaseConnections,
		models.MetricFreeableMemory,
	} {
		if err := c.storeMetric(ctx, instance.ID, metricName, timestamp, zeroValue); err != nil {
			return fmt.Errorf("failed to store zero %s: %w", metricName, err)
		}
	}
	return nil
}

// getClient returns or creates a CloudWatch client for the instance's account/region
func (c *MetricsCollector) getClient(ctx context.Context, instance models.Instance) (*CloudWatchClient, error) {
	// Only AWS instances are supported for active metric collection
	if instance.Provider != "aws" {
		return nil, fmt.Errorf("active metrics collection not supported for provider: %s", instance.Provider)
	}

	// Key by account+region since each region needs its own client
	key := fmt.Sprintf("%s_%s", instance.CloudAccountID, instance.Region)

	c.clientsMu.RLock()
	client, exists := c.clients[key]
	c.clientsMu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client - need to get credentials from account store
	account, err := c.accountStore.GetCloudAccount(instance.CloudAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud account: %w", err)
	}

	// Extract AWS credentials
	accessKey, _ := account.Credentials["aws_access_key_id"].(string)
	secretKey, _ := account.Credentials["aws_secret_access_key"].(string)

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing AWS credentials for account %s", account.Name)
	}

	client, err = NewCloudWatchClient(instance.Region, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CloudWatch client: %w", err)
	}

	c.clientsMu.Lock()
	c.clients[key] = client
	c.clientsMu.Unlock()

	return client, nil
}

// SetEnabled enables or disables the collector
func (c *MetricsCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}
