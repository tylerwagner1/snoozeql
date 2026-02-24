package metrics

import (
	"context"
	"errors"
	"fmt"
	"log"
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

// GetRDSMetrics fetches all relevant metrics for an RDS instance
// Returns metrics for the last hour, aggregated
// Returns an error if ALL metrics fail to fetch (no data available from CloudWatch)
func (c *CloudWatchClient) GetRDSMetrics(ctx context.Context, dbInstanceID string) (*RDSMetrics, error) {
	endTime := time.Now().UTC()
	startTime := endTime.Add(-1 * time.Hour)

	metrics := &RDSMetrics{
		InstanceID: dbInstanceID,
		Timestamp:  endTime.Truncate(time.Hour),
	}

	var metricsCollected int

	// Fetch each metric type
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

	output, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("GetMetricStatistics failed for %s: %w", metricName, err)
	}

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
