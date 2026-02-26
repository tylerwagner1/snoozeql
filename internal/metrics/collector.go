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

// CollectInstance collects metrics for a single instance on-demand
// This is the public API for manual/triggered collection
func (c *MetricsCollector) CollectInstance(ctx context.Context, instance models.Instance) error {
	// For stopped instances, store zeros
	if instance.Status != "available" && instance.Status != "running" {
		return c.storeZeroMetrics(ctx, instance)
	}

	// Only AWS supported for active collection
	if instance.Provider != "aws" {
		return fmt.Errorf("metrics collection not supported for provider: %s", instance.Provider)
	}

	client, err := c.getClient(ctx, instance)
	if err != nil {
		return err
	}

	return c.collectInstance(ctx, client, instance)
}

// collectInstance collects and stores metrics for a single instance
// Returns error only if all metrics fail AND we can't store zero metrics as fallback
func (c *MetricsCollector) collectInstance(ctx context.Context, client *CloudWatchClient, instance models.Instance) error {
	// Calculate timestamps for 15-minute window (3 datapoints at 5-minute intervals)
	now := time.Now().UTC()
	startTime := now.Add(-15 * time.Minute)
	endTime := now

	// Fetch all 5-minute datapoints
	metrics, err := client.GetRDSMetricsMultiple(ctx, instance.ProviderID, startTime, endTime)
	if err != nil {
		log.Printf("CloudWatch unavailable for %s: %v - storing zero metrics as fallback", instance.Name, err)
		// Store zero metrics as fallback when CloudWatch is unavailable
		return c.storeZeroMetrics(ctx, instance)
	}

	// Store each metric type for each datapoint (3 datapoints per 15-minute cycle)
	storedCount := 0
	for _, dp := range metrics {
		if dp.CPU != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricCPUUtilization, dp.Timestamp, dp.CPU); err != nil {
				log.Printf("Failed to store CPU metric for %s: %v", instance.Name, err)
			} else {
				storedCount++
			}
		}

		if dp.Connections != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricDatabaseConnections, dp.Timestamp, dp.Connections); err != nil {
				log.Printf("Failed to store Connections metric for %s: %v", instance.Name, err)
			} else {
				storedCount++
			}
		}

		if dp.ReadIOPS != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricReadIOPS, dp.Timestamp, dp.ReadIOPS); err != nil {
				log.Printf("Failed to store ReadIOPS metric for %s: %v", instance.Name, err)
			} else {
				storedCount++
			}
		}

		if dp.WriteIOPS != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricWriteIOPS, dp.Timestamp, dp.WriteIOPS); err != nil {
				log.Printf("Failed to store WriteIOPS metric for %s: %v", instance.Name, err)
			} else {
				storedCount++
			}
		}

		if dp.FreeMemory != nil {
			// Calculate memory percentage from bytes
			pct := CalculateMemoryPercentage(instance.InstanceType, dp.FreeMemory.Avg)
			if pct != nil {
				memValue := &MetricValue{Avg: *pct, Max: *pct, Min: *pct}
				if err := c.storeMetric(ctx, instance.ID, models.MetricFreeableMemory, dp.Timestamp, memValue); err != nil {
					log.Printf("Failed to store FreeableMemory metric for %s: %v", instance.Name, err)
				} else {
					storedCount++
				}
			} else {
				log.Printf("Unknown instance class %s for %s - skipping memory percentage", instance.InstanceType, instance.Name)
			}
		}
	}

	// If we stored no metrics at all, consider this a failure
	if storedCount == 0 {
		return fmt.Errorf("no metrics stored for %s", instance.Name)
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
// Generates 3 zero entries for current 15-minute window (one per 5-minute interval)
func (c *MetricsCollector) storeZeroMetrics(ctx context.Context, instance models.Instance) error {
	zeroValue := &MetricValue{Avg: 0, Max: 0, Min: 0}
	now := time.Now().UTC()
	const fiveMinute = 5 * time.Minute

	for i := 0; i < 3; i++ {
		// Timestamps at 5-minute intervals going backward
		timestamp := now.Truncate(fiveMinute).Add(-time.Duration(i*5) * time.Minute)

		for _, metricName := range []string{
			models.MetricCPUUtilization,
			models.MetricDatabaseConnections,
			models.MetricFreeableMemory,
		} {
			if err := c.storeMetric(ctx, instance.ID, metricName, timestamp, zeroValue); err != nil {
				return fmt.Errorf("failed to store zero %s: %w", metricName, err)
			}
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
	var accessKey, secretKey string
	accessKeyIface, ok := account.Credentials["aws_access_key_id"]
	if !ok {
		log.Printf("DEBUG: access_key_id NOT FOUND in credentials for account %s", account.Name)
	} else {
		var ok bool
		accessKey, ok = accessKeyIface.(string)
		if !ok {
			log.Printf("DEBUG: access_key_id is not a string, type=%T for account %s", accessKeyIface, account.Name)
		}
		log.Printf("DEBUG: access_key_id found: %v", accessKey[:10]+"...")
	}

	secretKeyIface, ok := account.Credentials["aws_secret_access_key"]
	if !ok {
		log.Printf("DEBUG: secret_access_key NOT FOUND in credentials for account %s", account.Name)
	} else {
		var ok bool
		secretKey, ok = secretKeyIface.(string)
		if !ok {
			log.Printf("DEBUG: secret_access_key is not a string, type=%T for account %s", secretKeyIface, account.Name)
		}
		log.Printf("DEBUG: secret_access_key found: %v", secretKey[:10]+"...")
	}

	if accessKey == "" || secretKey == "" {
		log.Printf("DEBUG: Final check - accessKey=%q, secretKey=%q, credentials len=%d", accessKey, secretKey, len(account.Credentials))
		return nil, fmt.Errorf("missing AWS credentials for account %s", account.Name)
	}

	client, err = NewCloudWatchClient(instance.Region, accessKey, secretKey)
	if err != nil {
		log.Printf("ERROR: failed to create CloudWatch client: %v", err)
		return nil, fmt.Errorf("failed to create CloudWatch client: %w", err)
	}
	log.Printf("DEBUG: CloudWatch client created successfully for %s (%s)", instance.Name, instance.CloudAccountID)

	c.clientsMu.Lock()
	c.clients[key] = client
	c.clientsMu.Unlock()

	return client, nil
}

// SetEnabled enables or disables the collector
func (c *MetricsCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// BackfillMetrics collects historical CloudWatch metrics for a specific instance
// over a given number of days, collecting metrics at hourly granularity.
// Returns the count of hours backfilled and any error encountered.
// The method self-throttles between hours to prevent CloudWatch rate limit errors.
func (c *MetricsCollector) BackfillMetrics(ctx context.Context, instance models.Instance, days int) (int, error) {
	// Cap days at 7 (CloudWatch free tier limitation)
	if days > 7 {
		days = 7
	}
	if days < 1 {
		days = 1
	}

	// Only AWS instances are supported
	if instance.Provider != "aws" {
		return 0, fmt.Errorf("backfill not supported for non-AWS instances (provider: %s)", instance.Provider)
	}

	// Get CloudWatch client
	client, err := c.getClient(ctx, instance)
	if err != nil {
		return 0, fmt.Errorf("failed to get CloudWatch client: %w", err)
	}

	// Calculate start and end times
	now := time.Now().UTC()
	endHour := now.Truncate(time.Hour)
	startHour := endHour.Add(-time.Duration(days*24) * time.Hour)

	log.Printf("BackfillMetrics: collecting %d days of metrics for %s (from %s to %s)",
		days, instance.Name, startHour.Format(time.RFC3339), endHour.Format(time.RFC3339))

	hoursBackfilled := 0

	// Iterate backward hour-by-hour
	currentHour := endHour
	for currentHour.Compare(startHour) >= 0 {
		// Check if this hour already has data
		hasData := true
		for _, metricName := range []string{
			models.MetricCPUUtilization,
			models.MetricDatabaseConnections,
			models.MetricFreeableMemory,
		} {
			exists, err := c.metricsStore.HourHasData(ctx, instance.ID, metricName, currentHour)
			if err != nil {
				log.Printf("Warning: failed to check if hour %s has data: %v", currentHour.Format(time.RFC3339), err)
			}
			if exists {
				hasData = true
			} else {
				hasData = false
				break
			}
		}

		// Skip hours that already have data
		if hasData {
			log.Printf("Skipping hour %s - already has data", currentHour.Format(time.RFC3339))
			currentHour = currentHour.Add(-1 * time.Hour)
			continue
		}

		// Fetch metrics for this specific hour
		metrics, err := client.GetRDSMetricsForHour(ctx, instance.ProviderID, currentHour)
		if err != nil {
			log.Printf("Failed to fetch metrics for %s at hour %s: %v", instance.Name, currentHour.Format(time.RFC3339), err)
			currentHour = currentHour.Add(-1 * time.Hour)
			continue
		}

		// Store each metric type
		metricStored := false
		if metrics.CPU != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricCPUUtilization, currentHour, metrics.CPU); err != nil {
				log.Printf("Failed to store CPU metric for %s at %s: %v", instance.Name, currentHour.Format(time.RFC3339), err)
			} else {
				metricStored = true
			}
		}

		if metrics.Connections != nil {
			if err := c.storeMetric(ctx, instance.ID, models.MetricDatabaseConnections, currentHour, metrics.Connections); err != nil {
				log.Printf("Failed to store Connections metric for %s at %s: %v", instance.Name, currentHour.Format(time.RFC3339), err)
			} else {
				metricStored = true
			}
		}

		if metrics.FreeMemory != nil {
			pct := CalculateMemoryPercentage(instance.InstanceType, metrics.FreeMemory.Avg)
			if pct != nil {
				memValue := &MetricValue{Avg: *pct, Max: *pct, Min: *pct}
				if err := c.storeMetric(ctx, instance.ID, models.MetricFreeableMemory, currentHour, memValue); err != nil {
					log.Printf("Failed to store FreeableMemory metric for %s at %s: %v", instance.Name, currentHour.Format(time.RFC3339), err)
				} else {
					metricStored = true
				}
			} else {
				log.Printf("Unknown instance class %s for %s - skipping memory percentage", instance.InstanceType, instance.Name)
			}
		}

		if metricStored {
			hoursBackfilled++
		}

		// Progress logging every 24 hours
		hoursRemaining := int(endHour.Sub(currentHour).Hours())
		if hoursRemaining > 0 && hoursRemaining%24 == 0 {
			log.Printf("Backfill progress: %d hours backfilled, %d hours remaining", hoursBackfilled, hoursRemaining)
		}

		// Self-throttle between hours to prevent rate limit errors
		select {
		case <-ctx.Done():
			return hoursBackfilled, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		// Move to previous hour
		currentHour = currentHour.Add(-1 * time.Hour)
	}

	log.Printf("BackfillMetrics complete: %d hours backfilled for %s", hoursBackfilled, instance.Name)
	return hoursBackfilled, nil
}
