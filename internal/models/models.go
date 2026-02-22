package models

import "time"

// CloudAccount represents a configured cloud provider account
type CloudAccount struct {
	ID               string         `json:"id" db:"id"`
	Name             string         `json:"name" db:"name"`
	Provider         string         `json:"provider" db:"provider"`
	Credentials      map[string]any `json:"-" db:"credentials"` // Not exposed in API
	Regions          []string       `json:"regions" db:"regions"`
	ConnectionStatus string         `json:"connection_status,omitempty" db:"connection_status"`
	LastSyncAt       *time.Time     `json:"last_sync_at,omitempty" db:"last_sync_at"`
	LastError        *string        `json:"last_error,omitempty" db:"last_error"`
	DeletedAt        *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

// Instance represents a discovered database instance
type Instance struct {
	ID              string            `json:"id" db:"id"`
	CloudAccountID  string            `json:"cloud_account_id" db:"cloud_account_id"`
	AccountID       string            `json:"account_id" db:"account_id"`       // AWS account ID from provider (not stored in DB, used for mapping)
	Provider        string            `json:"provider" db:"provider"`           // Cloud provider type: "aws" or "gcp"
	ProviderName    string            `json:"provider_name" db:"provider_name"` // Full provider identifier: "aws_{accountID}_{region}"
	ProviderID      string            `json:"provider_id" db:"provider_id"`
	Name            string            `json:"name" db:"name"`
	Region          string            `json:"region" db:"region"`
	InstanceType    string            `json:"instance_type" db:"instance_type"`
	Engine          string            `json:"engine" db:"engine"`
	Status          string            `json:"status" db:"status"`
	Managed         bool              `json:"managed" db:"managed"`
	Tags            map[string]string `json:"tags" db:"tags"`
	HourlyCostCents int               `json:"hourly_cost_cents" db:"hourly_cost_cents"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// Schedule represents a sleep/wake schedule
type Schedule struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Selectors   []Selector `json:"selectors" db:"selectors"`
	Timezone    string     `json:"timezone" db:"timezone"`
	SleepCron   string     `json:"sleep_cron" db:"sleep_cron"`
	WakeCron    string     `json:"wake_cron" db:"wake_cron"`
	Enabled     bool       `json:"enabled" db:"enabled"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Selector defines matching criteria for dynamic schedule assignment
type Selector struct {
	Name     *Matcher            `json:"name,omitempty" db:"name"`
	Provider *string             `json:"provider,omitempty" db:"provider"`
	Region   *Matcher            `json:"region,omitempty" db:"region"`
	Engine   *Matcher            `json:"engine,omitempty" db:"engine"`
	Tags     map[string]*Matcher `json:"tags,omitempty" db:"tags"`
}

// Matcher defines how to match a string value
type Matcher struct {
	Pattern string    `json:"pattern" db:"pattern"`
	Type    MatchType `json:"type" db:"type"`
}

// MatchType defines the matching strategy
type MatchType string

const (
	MatchExact    MatchType = "exact"
	MatchContains MatchType = "contains"
	MatchPrefix   MatchType = "prefix"
	MatchSuffix   MatchType = "suffix"
	MatchRegex    MatchType = "regex"
)

// Recommendation represents a suggested schedule based on activity patterns
type Recommendation struct {
	ID                string     `json:"id" db:"id"`
	InstanceID        string     `json:"instance_id" db:"instance_id"`
	DetectedPattern   []byte     `json:"detected_pattern" db:"detected_pattern"`     // JSONB
	SuggestedSchedule []byte     `json:"suggested_schedule" db:"suggested_schedule"` // JSONB
	ConfidenceScore   float64    `json:"confidence_score" db:"confidence_score"`
	Status            string     `json:"status" db:"status"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt        *time.Time `json:"resolved_at" db:"resolved_at"`
}

// Override represents an active manual override
type Override struct {
	ID         string     `json:"id" db:"id"`
	InstanceID string     `json:"instance_id" db:"instance_id"`
	Type       string     `json:"type" db:"type"`
	SkipAction *string    `json:"skip_action,omitempty" db:"skip_action"`
	UntilTime  *time.Time `json:"until_time,omitempty" db:"until_time"`
	Reason     *string    `json:"reason,omitempty" db:"reason"`
	CreatedBy  string     `json:"created_by" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	Expired    bool       `json:"expired" db:"expired"`
}

// Event represents a start/stop event
type Event struct {
	ID             string    `json:"id" db:"id"`
	InstanceID     string    `json:"instance_id" db:"instance_id"`
	EventType      string    `json:"event_type" db:"event_type"`
	TriggeredBy    string    `json:"triggered_by" db:"triggered_by"`
	PreviousStatus string    `json:"previous_status" db:"previous_status"`
	NewStatus      string    `json:"new_status" db:"new_status"`
	Metadata       []byte    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Saving represents aggregated savings data
type Saving struct {
	ID                    string `json:"id" db:"id"`
	InstanceID            string `json:"instance_id" db:"instance_id"`
	Date                  string `json:"date" db:"date"`
	StoppedMinutes        int    `json:"stopped_minutes" db:"stopped_minutes"`
	EstimatedSavingsCents int    `json:"estimated_savings_cents" db:"estimated_savings_cents"`
}

// APIToken represents an API key
type APIToken struct {
	ID         string     `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	KeyHash    string     `json:"-" db:"key_hash"`
	LastUsedAt *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at" db:"revoked_at"`
}

// DashboardStats represents the summary statistics for the dashboard
type DashboardStats struct {
	SavingsCents   int `json:"savings_cents"`
	RunningCount   int `json:"running_count"`
	SleepingCount  int `json:"sleeping_count"`
	PendingActions int `json:"pending_actions"`
}

// Settings represents application settings
type Settings struct {
	Key       string    `json:"key" db:"key"`
	Value     []byte    `json:"value" db:"value"` // JSONB
	Scope     string    `json:"scope" db:"scope"`
	ScopeID   string    `json:"scope_id" db:"scope_id"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
