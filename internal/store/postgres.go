package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
func (s *InstanceStore) UpsertInstance(ctx context.Context, instance *models.Instance) error {
	tagsJSON, err := json.Marshal(instance.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO instances (
			cloud_account_id, provider, provider_id, name, region,
			instance_type, engine, status, managed, tags, hourly_cost_cents
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (provider, provider_id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			tags = EXCLUDED.tags,
			hourly_cost_cents = EXCLUDED.hourly_cost_cents,
			updated_at = NOW()
		RETURNING id`

	return s.db.QueryRowContext(ctx, query,
		instance.CloudAccountID, instance.Provider, instance.ProviderID,
		instance.Name, instance.Region, instance.InstanceType, instance.Engine,
		instance.Status, instance.Managed, tagsJSON, instance.HourlyCostCents,
	).Scan(&instance.ID)
}

// ListInstances returns all instances from the database
func (s *InstanceStore) ListInstances(ctx context.Context) ([]models.Instance, error) {
	query := `
		SELECT id, cloud_account_id, provider, provider_id, name, region,
			instance_type, engine, status, managed, tags, hourly_cost_cents,
			created_at, updated_at
		FROM instances ORDER BY created_at DESC`

	rows, err := s.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []models.Instance
	for rows.Next() {
		var instance models.Instance
		var tagsJSON []byte

		err := rows.Scan(
			&instance.ID, &instance.CloudAccountID, &instance.Provider,
			&instance.ProviderID, &instance.Name, &instance.Region,
			&instance.InstanceType, &instance.Engine, &instance.Status,
			&instance.Managed, &tagsJSON, &instance.HourlyCostCents,
			&instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
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
func (s *InstanceStore) GetInstanceByProviderID(ctx context.Context, provider, providerID string) (*models.Instance, error) {
	var instance models.Instance
	var tagsJSON []byte

	err := s.db.db.QueryRowContext(ctx, `
		SELECT id, cloud_account_id, provider, provider_id, name, region,
			instance_type, engine, status, managed, tags, hourly_cost_cents,
			created_at, updated_at
		FROM instances WHERE provider = $1 AND provider_id = $2`, provider, providerID).Scan(
		&instance.ID, &instance.CloudAccountID, &instance.Provider,
		&instance.ProviderID, &instance.Name, &instance.Region,
		&instance.InstanceType, &instance.Engine, &instance.Status,
		&instance.Managed, &tagsJSON, &instance.HourlyCostCents,
		&instance.CreatedAt, &instance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse tags JSONB
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &instance.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	} else {
		instance.Tags = make(map[string]string)
	}

	return &instance, nil
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

	var events []models.Event
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

	var events []models.Event
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

// ListRecommendationsByStatus returns recommendations by status (for InstanceStore usage)
func (s *InstanceStore) ListRecommendationsByStatus(ctx context.Context, status string) ([]map[string]interface{}, error) {
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
		// Using a struct to receive the scan, then convert to map
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

// CreateRecommendation creates a new recommendation
func (s *RecommendationStore) CreateRecommendation(recommendation *models.Recommendation) error {
	return s.db.db.QueryRowContext(context.Background(), `
		INSERT INTO recommendations (instance_id, detected_pattern, suggested_schedule, confidence_score, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, recommendation.InstanceID, recommendation.DetectedPattern,
		recommendation.SuggestedSchedule, recommendation.ConfidenceScore, recommendation.Status).Scan(&recommendation.ID)
}

// UpdateRecommendation updates an existing recommendation
func (s *RecommendationStore) UpdateRecommendation(recommendation *models.Recommendation) error {
	_, err := s.db.db.ExecContext(context.Background(), `
		UPDATE recommendations SET
			status = $1, resolved_at = NOW()
		WHERE id = $2`,
		recommendation.Status, recommendation.ID)
	return err
}

// DeleteRecommendation deletes a recommendation
func (s *RecommendationStore) DeleteRecommendation(id string) error {
	_, err := s.db.db.ExecContext(context.Background(), "DELETE FROM recommendations WHERE id = $1", id)
	return err
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
	var connectionStatus, lastError sql.NullString
	var lastSyncAt sql.NullTime

	err := s.db.db.QueryRowContext(context.Background(), `
		SELECT id, name, provider, regions, connection_status, last_sync_at, last_error, created_at
		FROM cloud_accounts WHERE id = $1`, id).Scan(
		&account.ID, &account.Name, &account.Provider, &account.Regions,
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
		SELECT id, name, provider, regions, credentials, connection_status, last_sync_at, last_error, created_at
		FROM cloud_accounts ORDER BY created_at DESC`

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

		err := rows.Scan(
			&account.ID, &account.Name, &account.Provider, &regionsStr,
			&credentialsJSON, &connectionStatus, &lastSyncAt, &lastError, &account.CreatedAt,
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
	_, err := s.db.db.ExecContext(context.Background(), `
		UPDATE cloud_accounts SET
			name = $1, regions = $2
		WHERE id = $3`,
		account.Name, account.Regions, account.ID)
	return err
}

// DeleteCloudAccount deletes a cloud account
func (s *CloudAccountStore) DeleteCloudAccount(id string) error {
	_, err := s.db.db.ExecContext(context.Background(), "DELETE FROM cloud_accounts WHERE id = $1", id)
	return err
}
