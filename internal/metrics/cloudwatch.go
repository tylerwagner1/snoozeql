package metrics

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"snoozeql/internal/models"
)

// CloudWatchClient wraps the AWS CloudWatch client for RDS metrics
type CloudWatchClient struct {
	client *cloudwatch.Client
	region string
}

// NewCloudWatchClient creates a new CloudWatch client with credentials
func NewCloudWatchClient(region, accessKey, secretKey string) (*CloudWatchClient, error) {
	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &CloudWatchClient{
		client: cloudwatch.NewFromConfig(cfg),
		region: region,
	}, nil
}

// RDSMetrics holds the collected metrics for an RDS instance
type RDSMetrics struct {
	InstanceID  string
	Timestamp   time.Time
	CPU         *MetricValue
	Connections *MetricValue
	ReadIOPS    *MetricValue
	WriteIOPS   *MetricValue
	FreeMemory  *MetricValue
}

// MetricValue holds a single metric's statistics
type MetricValue struct {
	Avg float64
	Max float64
	Min float64
}

// RDSMetricDatapoint holds metrics for a single 5-minute interval
type RDSMetricDatapoint struct {
	Timestamp   time.Time
	CPU         *MetricValue
	Connections *MetricValue
	ReadIOPS    *MetricValue
	WriteIOPS   *MetricValue
	FreeMemory  *MetricValue
}

// GetRDSMetrics fetches all relevant metrics for an RDS instance
// Returns metrics for the last hour, aggregated
// Returns an error if ALL metrics fail to fetch (no data available from CloudWatch)
func (c *CloudWatchClient) GetRDSMetrics(ctx context.Context, dbInstanceID string) (*RDSMetrics, error) {
	return c.GetRDSMetricsForHour(ctx, dbInstanceID, time.Now().UTC().Add(-1*time.Hour))
}

// GetRDSMetricsForHour fetches all relevant metrics for an RDS instance for a specific hour
// Returns metrics for the given hour, aggregated
// Returns an error if ALL metrics fail to fetch (no data available from CloudWatch)
func (c *CloudWatchClient) GetRDSMetricsForHour(ctx context.Context, dbInstanceID string, hour time.Time) (*RDSMetrics, error) {
	startTime := hour
	endTime := hour.Add(1 * time.Hour)

	metrics := &RDSMetrics{
		InstanceID: dbInstanceID,
		Timestamp:  hour.Truncate(time.Hour),
	}

	var metricsCollected int

	// Fetch each metric type
	log.Printf("DEBUG: GetRDSMetricsForHour starting to fetch metrics for %s at hour %s", dbInstanceID, hour.Format(time.RFC3339))
	cpu, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricCPUUtilization, startTime, endTime)
	if err == nil {
		metrics.CPU = cpu
		metricsCollected++
	} else {
		log.Printf("CloudWatch: no datapoints for %s CPUUtilization - checking if this is expected", dbInstanceID)
	}

	conns, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricDatabaseConnections, startTime, endTime)
	if err == nil {
		metrics.Connections = conns
		metricsCollected++
	} else {
		log.Printf("CloudWatch: no datapoints for %s DatabaseConnections - checking if this is expected", dbInstanceID)
	}

	readIOPS, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricReadIOPS, startTime, endTime)
	if err == nil {
		metrics.ReadIOPS = readIOPS
		metricsCollected++
	} else {
		log.Printf("CloudWatch: no datapoints for %s ReadIOPS - checking if this is expected", dbInstanceID)
	}

	writeIOPS, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricWriteIOPS, startTime, endTime)
	if err == nil {
		metrics.WriteIOPS = writeIOPS
		metricsCollected++
	} else {
		log.Printf("CloudWatch: no datapoints for %s WriteIOPS - checking if this is expected", dbInstanceID)
	}

	freeMemory, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricFreeableMemory, startTime, endTime)
	if err == nil {
		metrics.FreeMemory = freeMemory
		metricsCollected++
	} else {
		log.Printf("CloudWatch: no datapoints for %s FreeableMemory - checking if this is expected", dbInstanceID)
	}

	// If no metrics were collected, CloudWatch might not have data for this instance
	if metricsCollected == 0 {
		return nil, fmt.Errorf("no CloudWatch datapoints available for instance %s - check CloudWatch has data or instance is running", dbInstanceID)
	}

	// Return metrics even if some are nil - the collector will store what's available
	return metrics, nil
}

// getMetricWithRetry fetches a single metric with 3 retry attempts
func (c *CloudWatchClient) getMetricWithRetry(ctx context.Context, dbInstanceID, metricName string, start, end time.Time) (*MetricValue, error) {
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		fmt.Printf("DEBUG: getMetricWithRetry calling getMetric for %s, attempt %d\n", metricName, attempt)
		value, err := c.getMetric(ctx, dbInstanceID, metricName, start, end)
		if err == nil {
			return value, nil
		}
		lastErr = err

		// Check for throttling - exponential backoff
		var limitErr *types.LimitExceededException
		if errors.As(err, &limitErr) {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		// For other errors, check if retryable
		if !isRetryableError(err) {
			return nil, err
		}
		time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed after 3 retries: %w", lastErr)
}

// getMetric fetches a single CloudWatch metric
func (c *CloudWatchClient) getMetric(ctx context.Context, dbInstanceID, metricName string, start, end time.Time) (*MetricValue, error) {
	// Write to file to verify function is called
	f, err := os.OpenFile("/tmp/cloudwatch_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "DEBUG: getMetric called for %s DBInstanceID: %s\n", metricName, dbInstanceID)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("DEBUG: getMetric called for %s DBInstanceID: %s", metricName, dbInstanceID)

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("DBInstanceIdentifier"),
				Value: aws.String(dbInstanceID),
			},
		},
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		Period:    aws.Int32(3600), // 1 hour
		Statistics: []types.Statistic{
			types.StatisticAverage,
			types.StatisticMaximum,
			types.StatisticMinimum,
		},
	}

	log.Printf("DEBUG: GetMetricStatistics for %s: Namespace=%s, MetricName=%s, DBInstanceID=%s, Period=%ds, Start=%s, End=%s",
		metricName, *input.Namespace, *input.MetricName, dbInstanceID, *input.Period, start.Format(time.RFC3339), end.Format(time.RFC3339))

	output, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		log.Printf("ERROR: GetMetricStatistics failed for %s: %v", metricName, err)
		return nil, fmt.Errorf("GetMetricStatistics failed for %s: %w", metricName, err)
	}

	log.Printf("DEBUG: GetMetricStatistics returned %d datapoints for %s", len(output.Datapoints), metricName)
	if len(output.Datapoints) == 0 {
		return nil, fmt.Errorf("no datapoints for %s", metricName)
	}

	// Use the most recent datapoint
	dp := output.Datapoints[0]
	for _, d := range output.Datapoints[1:] {
		if d.Timestamp.After(*dp.Timestamp) {
			dp = d
		}
	}

	return &MetricValue{
		Avg: aws.ToFloat64(dp.Average),
		Max: aws.ToFloat64(dp.Maximum),
		Min: aws.ToFloat64(dp.Minimum),
	}, nil
}

// getMetricMultiple fetches all CloudWatch datapoints for a metric within a time range
// Returns all datapoints (not just the most recent) for use with 5-minute periods
func (c *CloudWatchClient) getMetricMultiple(ctx context.Context, dbInstanceID, metricName string, start, end time.Time) ([]MetricValue, error) {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("DBInstanceIdentifier"),
				Value: aws.String(dbInstanceID),
			},
		},
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		Period:    aws.Int32(300), // 5 minutes
		Statistics: []types.Statistic{
			types.StatisticAverage,
			types.StatisticMaximum,
			types.StatisticMinimum,
		},
	}

	output, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("GetMetricStatistics failed for %s: %w", metricName, err)
	}

	if len(output.Datapoints) == 0 {
		return nil, fmt.Errorf("no datapoints for %s", metricName)
	}

	// Return all datapoints sorted by timestamp
	var datapoints []MetricValue
	for _, dp := range output.Datapoints {
		datapoints = append(datapoints, MetricValue{
			Avg: aws.ToFloat64(dp.Average),
			Max: aws.ToFloat64(dp.Maximum),
			Min: aws.ToFloat64(dp.Minimum),
		})
	}

	// Sort by timestamp ascending
	for i := 0; i < len(datapoints); i++ {
		for j := i + 1; j < len(datapoints); j++ {
			if output.Datapoints[i].Timestamp.Before(*output.Datapoints[j].Timestamp) {
				output.Datapoints[i], output.Datapoints[j] = output.Datapoints[j], output.Datapoints[i]
				datapoints[i], datapoints[j] = datapoints[j], datapoints[i]
			}
		}
	}

	return datapoints, nil
}

// GetRDSMetricsMultiple fetches all relevant metrics for an RDS instance over a time range
// Returns multiple datapoints at 5-minute intervals (Period=300)
// For a 15-minute window, returns up to 3 datapoints
func (c *CloudWatchClient) GetRDSMetricsMultiple(ctx context.Context, dbInstanceID string, start, end time.Time) ([]RDSMetricDatapoint, error) {
	log.Printf("DEBUG: GetRDSMetricsMultiple fetching 5-minute metrics for %s from %s to %s",
		dbInstanceID, start.Format(time.RFC3339), end.Format(time.RFC3339))

	var allDatapoints []RDSMetricDatapoint

	// Fetch each metric type and collect all datapoints by timestamp
	metricNames := []string{
		models.MetricCPUUtilization,
		models.MetricDatabaseConnections,
		models.MetricReadIOPS,
		models.MetricWriteIOPS,
		models.MetricFreeableMemory,
	}

	// Store datapoints by timestamp for each metric
	metricData := make(map[string][]MetricValue)

	// Fetch each metric type
	for _, metricName := range metricNames {
		dp, err := c.getMetricMultiple(ctx, dbInstanceID, metricName, start, end)
		if err != nil {
			log.Printf("CloudWatch: no datapoints for %s - checking if this is expected", metricName)
			continue
		}
		metricData[metricName] = dp
	}

	// Merge all datapoints by timestamp
	// First, iterate through CPU datapoints (most reliable) and build the base
	if cpus, ok := metricData[models.MetricCPUUtilization]; ok {
		for _, cpu := range cpus {
			dp := &RDSMetricDatapoint{}
			dp.CPU = &cpu
			allDatapoints = append(allDatapoints, *dp)
		}
	}

	// Now merge other metrics into existing datapoints
	if conns, ok := metricData[models.MetricDatabaseConnections]; ok {
		for i, conn := range conns {
			if i < len(allDatapoints) {
				allDatapoints[i].Connections = &conn
			}
		}
	}

	if readIOPS, ok := metricData[models.MetricReadIOPS]; ok {
		for i, v := range readIOPS {
			if i < len(allDatapoints) {
				allDatapoints[i].ReadIOPS = &v
			}
		}
	}

	if writeIOPS, ok := metricData[models.MetricWriteIOPS]; ok {
		for i, v := range writeIOPS {
			if i < len(allDatapoints) {
				allDatapoints[i].WriteIOPS = &v
			}
		}
	}

	if freeMem, ok := metricData[models.MetricFreeableMemory]; ok {
		for i, v := range freeMem {
			if i < len(allDatapoints) {
				allDatapoints[i].FreeMemory = &v
			}
		}
	}

	if len(allDatapoints) == 0 {
		return nil, fmt.Errorf("no CloudWatch datapoints available for instance %s across all metrics", dbInstanceID)
	}

	return allDatapoints, nil
}

// isRetryableError checks if an error is transient and should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Network errors, timeouts, and server errors are retryable
	errStr := err.Error()
	return contains(errStr, "timeout") ||
		contains(errStr, "connection") ||
		contains(errStr, "503") ||
		contains(errStr, "500")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
