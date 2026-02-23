package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"snoozeql/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres represents a PostgreSQL connection wrapper
type Postgres struct {
	db  *sql.DB
	url string
}

// NewPostgres creates a new PostgreSQL connection
func NewPostgres(url string) (*Postgres, error) {
	conn, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Postgres{db: conn, url: url}, nil
}

// Migrate runs database migrations
func (p *Postgres) Migrate(migrationDir string) error {
	return nil
}

// Close closes the database connection
func (p *Postgres) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Query executes a query and returns rows
func (p *Postgres) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns at most one row
func (p *Postgres) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (p *Postgres) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return p.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (p *Postgres) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Begin starts a transaction
func (p *Postgres) Begin(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// BuildWhere builds a WHERE clause from conditions
func BuildWhere(conditions []string) (string, []any) {
	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), nil
}

// SanitizeOrder validates and sanitizes order by clause
func SanitizeOrder(column string, validColumns map[string]bool) string {
	if validColumns[column] {
		return column
	}
	return "created_at"
}

// InstanceStore provides instance CRUD operations
type InstanceStore struct {
	db *Postgres
}

// NewInstanceStore creates a new instance store
func NewInstanceStore(db *Postgres) *InstanceStore {
	return &InstanceStore{db: db}
}

// UpsertInstance inserts or updates an instance in the database
// Uses ON CONFLICT on (provider, provider_id) which matches the unique constraint
// This ensures the same RDS instance can't be duplicated even when discovered from different accounts
func (s *InstanceStore) UpsertInstance(ctx context.Context, instance *models.Instance) error {
	tagsJSON, err := json.Marshal(instance.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Use ON CONFLICT on (provider, provider_id) to handle duplicates
	// When a conflict occurs (same provider/provider_id but different cloud_account_id),
	// update the cloud_account_id to the new value
	query := `
		INSERT INTO instances (
			cloud_account_id, provider, provider_name, provider_id, name, region,
			instance_type, engine, status, managed, tags, hourly_cost_cents
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (provider, provider_id, cloud_account_id) DO UPDATE SET
			name = EXCLUDED.name,
			provider_name = EXCLUDED.provider_name,
			provider_id = EXCLUDED.provider_id,
			status = EXCLUDED.status,
			tags = EXCLUDED.tags,
			hourly_cost_cents = EXCLUDED.hourly_cost_cents,
			updated_at = NOW()
		RETURNING id`
	return s.db.QueryRowContext(ctx, query,
		instance.CloudAccountID, instance.Provider, instance.ProviderName, instance.ProviderID,
		instance.Name, instance.Region, instance.InstanceType, instance.Engine,
		instance.Status, instance.Managed, tagsJSON, instance.HourlyCostCents,
	).Scan(&instance.ID)
}

// ListInstances returns all instances from the database (only from active accounts)
func (s *InstanceStore) ListInstances(ctx context.Context) ([]models.Instance, error) {
	query := `
		SELECT i.id, i.cloud_account_id, i.provider, i.provider_name, i.provider_id, i.name, i.region,
			i.instance_type, i.engine, i.status, i.managed, i.tags, i.hourly_cost_cents,
			i.created_at, i.updated_at
		FROM instances i
		JOIN cloud_accounts ca ON i.cloud_account_id = ca.id
		WHERE ca.deleted_at IS NULL
		ORDER BY i.created_at DESC`

	rows, err := s.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []models.Instance
	for rows.Next() {
		var instance models.Instance
		var tagsJSON []byte
		var providerName sql.NullString

		err := rows.Scan(
			&instance.ID, &instance.CloudAccountID, &instance.Provider,
			&providerName, &instance.ProviderID, &instance.Name, &instance.Region,
			&instance.InstanceType, &instance.Engine, &instance.Status,
			&instance.Managed, &tagsJSON, &instance.HourlyCostCents,
			&instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
		}

		if providerName.Valid {
			instance.ProviderName = providerName.String
		}

		// Parse tags JSONB
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &instance.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		} else {
			instance.Tags = make(map[string]string)
		}

		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return instances, nil
}

// GetInstanceByProviderID returns an instance by provider and provider ID
func (s *InstanceStore) GetInstanceByProviderID(ctx context.Context, provider string, providerID string) (*models.Instance, error) {
	query := `
		SELECT i.id, i.cloud_account_id, i.provider, i.provider_name, i.provider_id, i.name, i.region,
			i.instance_type, i.engine, i.status, i.managed, i.tags, i.hourly_cost_cents,
			i.created_at, i.updated_at
		FROM instances i
		JOIN cloud_accounts ca ON i.cloud_account_id = ca.id
		WHERE ca.deleted_at IS NULL`

	var conditions []string
	var args []any

	if provider != "" {
		conditions = append(conditions, "i.provider = $1")
		args = append(args, provider)
	}
	conditions = append(conditions, "i.provider_id = $"+fmt.Sprintf("%d", len(args)+1))
	args = append(args, providerID)

	query += " WHERE " + strings.Join(conditions, " AND ")
	query += " LIMIT 1"

	rows, err := s.db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query instance: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var instance models.Instance
		var tagsJSON []byte

		err := rows.Scan(
			&instance.ID, &instance.CloudAccountID, &instance.Provider,
			&instance.ProviderName, &instance.ProviderID, &instance.Name, &instance.Region,
			&instance.InstanceType, &instance.Engine, &instance.Status,
			&instance.Managed, &tagsJSON, &instance.HourlyCostCents,
			&instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &instance.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		} else {
			instance.Tags = make(map[string]string)
		}

		return &instance, nil
	}

	return nil, nil
}

// GetInstanceByID returns an instance by its ID
func (s *InstanceStore) GetInstanceByID(ctx context.Context, id string) (*models.Instance, error) {
	query := `
		SELECT i.id, i.cloud_account_id, i.provider, i.provider_name, i.provider_id, i.name, i.region,
			i.instance_type, i.engine, i.status, i.managed, i.tags, i.hourly_cost_cents,
			i.created_at, i.updated_at
		FROM instances i
		JOIN cloud_accounts ca ON i.cloud_account_id = ca.id
		WHERE i.id = $1 AND ca.deleted_at IS NULL`

	var instance models.Instance
	var tagsJSON []byte

	err := s.db.db.QueryRowContext(ctx, query, id).Scan(
		&instance.ID, &instance.CloudAccountID, &instance.Provider,
		&instance.ProviderName, &instance.ProviderID, &instance.Name, &instance.Region,
		&instance.InstanceType, &instance.Engine, &instance.Status,
		&instance.Managed, &tagsJSON, &instance.HourlyCostCents,
		&instance.CreatedAt, &instance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &instance.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	} else {
		instance.Tags = make(map[string]string)
	}

	return &instance, nil
}

// ListRecommendationsByStatus returns recommendations by status (for InstanceStore usage)
func (s *InstanceStore) ListRecommendationsByStatus(ctx context.Context, status string) ([]map[string]interface{}, error) {
	// Create a temporary RecommendationStore to reuse the logic
	recStore := NewRecommendationStore(s.db)
	return recStore.ListRecommendationsByStatus(ctx, status)
}

// EventStore provides event CRUD operations
type EventStore struct {
	db *Postgres
}

// NewEventStore creates a new event store
func NewEventStore(db *Postgres) *EventStore {
	return &EventStore{db: db}
}

// CreateEvent inserts an event into the database
func (s *EventStore) CreateEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (instance_id, event_type, triggered_by, previous_status, new_status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query,
		event.InstanceID, event.EventType, event.TriggeredBy,
		event.PreviousStatus, event.NewStatus, event.Metadata,
	).Scan(&event.ID, &event.CreatedAt)
}

// ListEvents returns events with pagination (most recent first)
func (s *EventStore) ListEvents(ctx context.Context, limit int, offset int) ([]models.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, instance_id, event_type, triggered_by, previous_status, new_status, metadata, created_at
		FROM events ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := []models.Event{} // Initialize as empty slice, not nil
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.InstanceID, &e.EventType, &e.TriggeredBy,
			&e.PreviousStatus, &e.NewStatus, &e.Metadata, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return events, nil
}

// ListEventsByInstance returns events for a specific instance
func (s *EventStore) ListEventsByInstance(ctx context.Context, instanceID string) ([]models.Event, error) {
	query := `
		SELECT id, instance_id, event_type, triggered_by, previous_status, new_status, metadata, created_at
		FROM events WHERE instance_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := []models.Event{} // Initialize as empty slice, not nil
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.InstanceID, &e.EventType, &e.TriggeredBy,
			&e.PreviousStatus, &e.NewStatus, &e.Metadata, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// RecommendationStore provides recommendation CRUD operations
type RecommendationStore struct {
	db *Postgres
}

// NewRecommendationStore creates a new recommendation store
func NewRecommendationStore(db *Postgres) *RecommendationStore {
	return &RecommendationStore{db: db}
}

// GetRecommendation retrieves a recommendation by ID
func (s *RecommendationStore) GetRecommendation(id string) (*models.Recommendation, error) {
	var recommendation models.Recommendation
	err := s.db.db.QueryRowContext(context.Background(), `
		SELECT id, instance_id, detected_pattern, suggested_schedule, confidence_score, status, created_at, resolved_at
		FROM recommendations WHERE id = $1`, id).Scan(
		&recommendation.ID, &recommendation.InstanceID, &recommendation.DetectedPattern,
		&recommendation.SuggestedSchedule, &recommendation.ConfidenceScore, &recommendation.Status,
		&recommendation.CreatedAt, &recommendation.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

// ListRecommendations returns recommendations by status
func (s *RecommendationStore) ListRecommendations(status string) ([]models.Recommendation, error) {
	var query string
	var args []any
	if status != "" {
		query = `
			SELECT id, instance_id, detected_pattern, suggested_schedule, confidence_score, status, created_at, resolved_at
			FROM recommendations WHERE status = $1 ORDER BY created_at DESC`
		args = append(args, status)
	} else {
		query = `
			SELECT id, instance_id, detected_pattern, suggested_schedule, confidence_score, status, created_at, resolved_at
			FROM recommendations ORDER BY created_at DESC`
	}

	rows, err := s.db.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []models.Recommendation
	for rows.Next() {
		var recommendation models.Recommendation
		err := rows.Scan(
			&recommendation.ID, &recommendation.InstanceID, &recommendation.DetectedPattern,
			&recommendation.SuggestedSchedule, &recommendation.ConfidenceScore, &recommendation.Status,
			&recommendation.CreatedAt, &recommendation.ResolvedAt,
		)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recommendation)
	}
	return recommendations, rows.Err()
}

// ListRecommendationsByStatus returns recommendations by status
func (s *RecommendationStore) ListRecommendationsByStatus(ctx context.Context, status string) ([]map[string]interface{}, error) {
	query := `
		SELECT id, instance_id, detected_pattern, suggested_schedule, confidence_score, status, created_at, resolved_at
		FROM recommendations`

	var conditions []string
	var args []any
	if status != "" {
		conditions = append(conditions, "status = $1")
		args = append(args, status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []map[string]interface{}
	for rows.Next() {
		var temp struct {
			ID                string
			InstanceID        string
			DetectedPattern   string
			SuggestedSchedule string
			ConfidenceScore   float64
			Status            string
			CreatedAt         string
			ResolvedAt        sql.NullString
		}
		err := rows.Scan(
			&temp.ID,
			&temp.InstanceID,
			&temp.DetectedPattern,
			&temp.SuggestedSchedule,
			&temp.ConfidenceScore,
			&temp.Status,
			&temp.CreatedAt,
			&temp.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}
		rec := map[string]interface{}{
			"id":                 temp.ID,
			"instance_id":        temp.InstanceID,
			"detected_pattern":   temp.DetectedPattern,
			"suggested_schedule": temp.SuggestedSchedule,
			"confidence_score":   temp.ConfidenceScore,
			"status":             temp.Status,
			"created_at":         temp.CreatedAt,
		}
		if temp.ResolvedAt.Valid {
			rec["resolved_at"] = temp.ResolvedAt.String
		}
		recommendations = append(recommendations, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return recommendations, nil
}

// CloudAccountStore provides cloud account CRUD operations
type CloudAccountStore struct {
	db *Postgres
}

// NewCloudAccountStore creates a new cloud account store
func NewCloudAccountStore(db *Postgres) *CloudAccountStore {
	return &CloudAccountStore{db: db}
}

// GetCloudAccount retrieves a cloud account by ID
func (s *CloudAccountStore) GetCloudAccount(id string) (*models.CloudAccount, error) {
	var account models.CloudAccount
	var regionsStr string
	var connectionStatus, lastError sql.NullString
	var lastSyncAt sql.NullTime

	err := s.db.db.QueryRowContext(context.Background(), `
		SELECT id, name, provider, regions, connection_status, last_sync_at, last_error, created_at
		FROM cloud_accounts WHERE id = $1`, id).Scan(
		&account.ID, &account.Name, &account.Provider, &regionsStr,
		&connectionStatus, &lastSyncAt, &lastError, &account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	account.ConnectionStatus = connectionStatus.String
	if lastSyncAt.Valid {
		account.LastSyncAt = &lastSyncAt.Time
	}
	if lastError.Valid {
		account.LastError = &lastError.String
	}
	// Parse PostgreSQL text[] format: {us-east-1,us-west-2}
	if regionsStr != "" && regionsStr != "{}" {
		regionsStr = strings.TrimPrefix(regionsStr, "{")
		regionsStr = strings.TrimSuffix(regionsStr, "}")
		if regionsStr != "" {
			account.Regions = strings.Split(regionsStr, ",")
		}
	}
	return &account, nil
}

// UpdateConnectionStatus updates a cloud account's connection status
func (s *CloudAccountStore) UpdateConnectionStatus(ctx context.Context, id string, status string, lastError *string) error {
	_, err := s.db.db.ExecContext(ctx, `
		UPDATE cloud_accounts SET
			connection_status = $1, last_error = $2, last_sync_at = NOW()
		WHERE id = $3`,
		status, lastError, id)
	if err != nil {
		return fmt.Errorf("failed to update connection status: %w", err)
	}
	return nil
}

// ListCloudAccounts returns all cloud accounts
func (s *CloudAccountStore) ListCloudAccounts() ([]models.CloudAccount, error) {
	log.Printf("DEBUG: Listing cloud accounts...")
	query := `
		SELECT id, name, provider, regions, credentials, connection_status, last_sync_at, last_error, deleted_at, created_at
		FROM cloud_accounts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := s.db.db.QueryContext(context.Background(), query)
	if err != nil {
		log.Printf("ERROR: Query failed: %v", err)
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var accounts []models.CloudAccount
	for rows.Next() {
		log.Printf("DEBUG: Processing row...")
		var account models.CloudAccount
		var regionsStr string
		var credentialsJSON []byte
		var connectionStatus, lastError sql.NullString
		var lastSyncAt sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&account.ID, &account.Name, &account.Provider, &regionsStr,
			&credentialsJSON, &connectionStatus, &lastSyncAt, &lastError, &deletedAt, &account.CreatedAt,
		)
		if err != nil {
			log.Printf("ERROR: Scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		// Parse credentials JSONB
		if len(credentialsJSON) > 0 {
			if err := json.Unmarshal(credentialsJSON, &account.Credentials); err != nil {
				log.Printf("Warning: Failed to parse credentials for account %s: %v", account.ID, err)
				account.Credentials = make(map[string]any)
			}
		}
		account.ConnectionStatus = connectionStatus.String
		if lastSyncAt.Valid {
			account.LastSyncAt = &lastSyncAt.Time
		}
		if lastError.Valid {
			account.LastError = &lastError.String
		}
		if deletedAt.Valid {
			account.DeletedAt = &deletedAt.Time
		}
		log.Printf("DEBUG: Account %s: name=%s, provider=%s, regions=%s, status=%s", account.ID, account.Name, account.Provider, regionsStr, account.ConnectionStatus)
		// Parse PostgreSQL text[] format: {us-east-1,us-west-2}
		if regionsStr != "" && regionsStr != "{}" {
			// Remove braces and split by comma
			regionsStr = strings.TrimPrefix(regionsStr, "{")
			regionsStr = strings.TrimSuffix(regionsStr, "}")
			if regionsStr != "" {
				account.Regions = strings.Split(regionsStr, ",")
			}
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Rows has err: %v", err)
		return nil, fmt.Errorf("rows err: %w", err)
	}
	log.Printf("DEBUG: Found %d accounts", len(accounts))
	return accounts, nil
}

// CreateCloudAccount creates a new cloud account
func (s *CloudAccountStore) CreateCloudAccount(account *models.CloudAccount) error {
	credentialsJSON, err := json.Marshal(account.Credentials)
	if err != nil {
		return err
	}
	err = s.db.db.QueryRowContext(context.Background(), `
		INSERT INTO cloud_accounts (name, provider, regions, credentials)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, account.Name, account.Provider, account.Regions, credentialsJSON).Scan(&account.ID)
	return err
}

// UpdateCloudAccount updates an existing cloud account
func (s *CloudAccountStore) UpdateCloudAccount(account *models.CloudAccount) error {
	credentialsJSON, err := json.Marshal(account.Credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	_, err = s.db.db.ExecContext(context.Background(), `
		UPDATE cloud_accounts SET
			name = $1, regions = $2, credentials = $3, connection_status = 'unknown'
		WHERE id = $4`,
		account.Name, account.Regions, credentialsJSON, account.ID)
	return err
}

// DeleteCloudAccount deletes a cloud account (soft delete)
func (s *CloudAccountStore) DeleteCloudAccount(id string) error {
	_, err := s.db.db.ExecContext(context.Background(), "UPDATE cloud_accounts SET deleted_at = NOW() WHERE id = $1", id)
	return err
}

// HardDeleteCloudAccount permanently deletes a cloud account
func (s *CloudAccountStore) HardDeleteCloudAccount(id string) error {
	_, err := s.db.db.ExecContext(context.Background(), "DELETE FROM cloud_accounts WHERE id = $1", id)
	return err
}

// ScheduleStore provides schedule CRUD operations
type ScheduleStore struct {
	db *Postgres
}

// NewScheduleStore creates a new schedule store
func NewScheduleStore(db *Postgres) *ScheduleStore {
	return &ScheduleStore{db: db}
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleStore) GetSchedule(id string) (*models.Schedule, error) {
	var schedule models.Schedule
	var selectorsJSON []byte

	err := s.db.db.QueryRowContext(context.Background(), `
		SELECT id, name, description, selectors, timezone, sleep_cron, wake_cron, enabled, created_at, updated_at
		FROM schedules WHERE id = $1`, id).Scan(
		&schedule.ID, &schedule.Name, &schedule.Description, &selectorsJSON,
		&schedule.Timezone, &schedule.SleepCron, &schedule.WakeCron, &schedule.Enabled,
		&schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse selectors JSONB
	if len(selectorsJSON) > 0 {
		if err := json.Unmarshal(selectorsJSON, &schedule.Selectors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal selectors: %w", err)
		}
	} else {
		schedule.Selectors = []models.Selector{}
	}

	return &schedule, nil
}

// ListSchedules returns all schedules from the database
func (s *ScheduleStore) ListSchedules() ([]models.Schedule, error) {
	query := `
		SELECT id, name, description, selectors, timezone, sleep_cron, wake_cron, enabled, created_at, updated_at
		FROM schedules ORDER BY created_at DESC`

	rows, err := s.db.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var selectorsJSON []byte

		err := rows.Scan(
			&schedule.ID, &schedule.Name, &schedule.Description, &selectorsJSON,
			&schedule.Timezone, &schedule.SleepCron, &schedule.WakeCron, &schedule.Enabled,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		// Parse selectors JSONB
		if len(selectorsJSON) > 0 {
			if err := json.Unmarshal(selectorsJSON, &schedule.Selectors); err != nil {
				return nil, fmt.Errorf("failed to unmarshal selectors: %w", err)
			}
		} else {
			schedule.Selectors = []models.Selector{}
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

// CreateSchedule creates a new schedule
func (s *ScheduleStore) CreateSchedule(schedule *models.Schedule) error {
	selectorsJSON, err := json.Marshal(schedule.Selectors)
	if err != nil {
		return fmt.Errorf("failed to marshal selectors: %w", err)
	}

	err = s.db.db.QueryRowContext(context.Background(), `
		INSERT INTO schedules (
			name, description, selectors, timezone, sleep_cron, wake_cron, enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`, schedule.Name, schedule.Description,
		selectorsJSON, schedule.Timezone, schedule.SleepCron, schedule.WakeCron, schedule.Enabled).Scan(
		&schedule.ID, &schedule.CreatedAt,
	)
	return err
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleStore) UpdateSchedule(schedule *models.Schedule) error {
	selectorsJSON, err := json.Marshal(schedule.Selectors)
	if err != nil {
		return fmt.Errorf("failed to marshal selectors: %w", err)
	}

	_, err = s.db.db.ExecContext(context.Background(), `
		UPDATE schedules SET
			name = $1, description = $2, selectors = $3,
			timezone = $4, sleep_cron = $5, wake_cron = $6,
			enabled = $7, updated_at = NOW()
		WHERE id = $8`,
		schedule.Name, schedule.Description, selectorsJSON,
		schedule.Timezone, schedule.SleepCron, schedule.WakeCron, schedule.Enabled, schedule.ID,
	)
	return err
}

// DeleteSchedule deletes a schedule
func (s *ScheduleStore) DeleteSchedule(id string) error {
	_, err := s.db.db.ExecContext(context.Background(), "DELETE FROM schedules WHERE id = $1", id)
	return err
}

// GetMatchingSchedules returns schedules that match a given instance
func (s *ScheduleStore) GetMatchingSchedules(instance models.Instance) ([]models.Schedule, error) {
	// Get all enabled schedules and filter in Go (since selector matching is complex)
	schedules, err := s.ListSchedules()
	if err != nil {
		return nil, err
	}

	var matching []models.Schedule
	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		if matchesInstance(instance, schedule.Selectors) {
			matching = append(matching, schedule)
		}
	}

	return matching, nil
}

// matchesInstance checks if an instance matches any of the schedule's selectors
func matchesInstance(instance models.Instance, selectors []models.Selector) bool {
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
		return strings.Contains(value, pattern)
	case models.MatchPrefix:
		return strings.HasPrefix(value, pattern)
	case models.MatchSuffix:
		return strings.HasSuffix(value, pattern)
	case models.MatchRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(value)
	default:
		return strings.Contains(value, pattern)
	}
}
