package gcp

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	cloudsql "google.golang.org/api/sqladmin/v1"

	"snoozeql/internal/models"
)

// CloudSQLProvider implements the Provider interface for GCP Cloud SQL
type CloudSQLProvider struct {
	sqlAdminService *cloudsql.Service
	projectID       string
	region          string
	managedTags     []string
}

// NewCloudSQLProvider creates a new GCP Cloud SQL provider
func NewCloudSQLProvider(projectID, region string, managedTags []string, serviceAccountJSON string) (*CloudSQLProvider, error) {
	var service *cloudsql.Service
	var err error

	if serviceAccountJSON != "" {
		service, err = cloudsql.NewService(context.Background(),
			option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	} else {
		// Fall back to ADC (Application Default Credentials)
		service, err = cloudsql.NewService(context.Background())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud SQL Admin client: %w", err)
	}

	return &CloudSQLProvider{
		sqlAdminService: service,
		projectID:       projectID,
		region:          region,
		managedTags:     managedTags,
	}, nil
}

// TestConnection tests if the GCP credentials are valid
func (p *CloudSQLProvider) TestConnection(ctx context.Context) error {
	_, err := p.sqlAdminService.Instances.List(p.projectID).Context(ctx).MaxResults(1).Do()
	if err != nil {
		return fmt.Errorf("failed to test GCP connection: %w", err)
	}
	return nil
}

// ListDatabases returns all Cloud SQL instances
func (p *CloudSQLProvider) ListDatabases(ctx context.Context) ([]models.Instance, error) {
	var instances []models.Instance

	result, err := p.sqlAdminService.Instances.List(p.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud SQL instances: %w", err)
	}

	for _, db := range result.Items {
		instance, err := p.instanceToModel(db)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// StartDatabase starts a stopped Cloud SQL instance
func (p *CloudSQLProvider) StartDatabase(ctx context.Context, id string) error {
	instance := &cloudsql.DatabaseInstance{
		Settings: &cloudsql.Settings{
			ActivationPolicy: "ALWAYS",
		},
	}

	_, err := p.sqlAdminService.Instances.Patch(p.projectID, id, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to start Cloud SQL instance %s: %w", id, err)
	}

	return nil
}

// StopDatabase stops a running Cloud SQL instance
func (p *CloudSQLProvider) StopDatabase(ctx context.Context, id string) error {
	instance := &cloudsql.DatabaseInstance{
		Settings: &cloudsql.Settings{
			ActivationPolicy: "NEVER",
		},
	}

	_, err := p.sqlAdminService.Instances.Patch(p.projectID, id, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to stop Cloud SQL instance %s: %w", id, err)
	}

	return nil
}

// GetDatabaseStatus returns the current status of a database
func (p *CloudSQLProvider) GetDatabaseStatus(ctx context.Context, id string) (string, error) {
	result, err := p.sqlAdminService.Instances.Get(p.projectID, id).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get Cloud SQL instance %s: %w", id, err)
	}

	return result.State, nil
}

// GetMetrics returns activity metrics for a database
func (p *CloudSQLProvider) GetMetrics(ctx context.Context, providerName string, id string, period string) (map[string]any, error) {
	metrics := make(map[string]any)

	duration, err := parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("invalid period: %w", err)
	}

	_ = duration

	metrics["cpu"] = map[string]any{
		"error": "GCP Cloud Monitoring not yet implemented",
	}
	metrics["connections"] = map[string]any{
		"error": "GCP Cloud Monitoring not yet implemented",
	}
	metrics["read_iops"] = map[string]any{
		"error": "GCP Cloud Monitoring not yet implemented",
	}
	metrics["write_iops"] = map[string]any{
		"error": "GCP Cloud Monitoring not yet implemented",
	}

	return metrics, nil
}

// GetDatabaseByID returns a database by its ID
func (p *CloudSQLProvider) GetDatabaseByID(ctx context.Context, id string) (*models.Instance, error) {
	result, err := p.sqlAdminService.Instances.Get(p.projectID, id).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get Cloud SQL instance %s: %w", id, err)
	}

	inst, err := p.instanceToModel(result)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (p *CloudSQLProvider) instanceToModel(db *cloudsql.DatabaseInstance) (models.Instance, error) {
	tags := make(map[string]string)
	if db.Settings != nil && db.Settings.UserLabels != nil {
		for k, v := range db.Settings.UserLabels {
			tags[k] = v
		}
	}

	return models.Instance{
		Provider:        "gcp",
		ProviderID:      fmt.Sprintf("projects/%s/instances/%s", p.projectID, db.Name),
		Name:            db.Name,
		Region:          db.Region,
		InstanceType:    "unknown",
		Engine:          db.DatabaseVersion,
		Status:          db.State,
		Managed:         p.isManaged(tags),
		Tags:            tags,
		HourlyCostCents: 50,
	}, nil
}

func (p *CloudSQLProvider) isManaged(tags map[string]string) bool {
	if len(p.managedTags) == 0 {
		return true
	}
	for _, tag := range p.managedTags {
		if _, exists := tags[tag]; exists {
			return true
		}
	}
	return false
}

func parsePeriod(period string) (time.Duration, error) {
	switch period {
	case "1h", "1 hour":
		return time.Hour, nil
	case "24h", "1d", "1 day":
		return 24 * time.Hour, nil
	case "7d", "7 day", "7 days":
		return 7 * 24 * time.Hour, nil
	case "30d", "30 day", "30 days":
		return 30 * 24 * time.Hour, nil
	default:
		return time.Hour, fmt.Errorf("unknown period: %s", period)
	}
}
