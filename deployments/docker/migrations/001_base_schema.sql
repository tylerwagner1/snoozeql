-- Database: SnoozeQL
-- Run this script to create the database schema

-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Cloud provider accounts
CREATE TABLE cloud_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('aws', 'gcp')),
    credentials JSONB NOT NULL,
    regions TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Discovered/managed database instances
CREATE TABLE instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cloud_account_id UUID REFERENCES cloud_accounts(id) ON DELETE SET NULL,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('aws', 'gcp')),
    provider_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100) NOT NULL,
    instance_type VARCHAR(100),
    engine VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    managed BOOLEAN DEFAULT FALSE,
    tags JSONB DEFAULT '{}',
    hourly_cost_cents INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider, provider_id)
);

-- Sleep/wake schedules with dynamic selectors
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    selectors JSONB NOT NULL,
    timezone VARCHAR(100) NOT NULL,
    sleep_cron VARCHAR(100) NOT NULL,
    wake_cron VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Configuration settings
CREATE TABLE settings (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    scope VARCHAR(50) DEFAULT 'global',
    scope_id VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Activity-based schedule recommendations
CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    detected_pattern JSONB NOT NULL,
    suggested_schedule JSONB NOT NULL,
    confidence_score FLOAT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'dismissed')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- Active overrides (keep-alive, skip-next)
CREATE TABLE overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('keep_alive', 'skip_next')),
    skip_action VARCHAR(50) CHECK (skip_action IN ('start', 'stop')),
    until_time TIMESTAMPTZ,
    reason TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expired BOOLEAN DEFAULT FALSE,
    UNIQUE(instance_id, type) WHERE NOT expired
);

-- Event log for tracking and cost calculation
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    triggered_by VARCHAR(100),
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Aggregated savings data
CREATE TABLE savings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    stopped_minutes INTEGER NOT NULL,
    estimated_savings_cents INTEGER NOT NULL,
    UNIQUE(instance_id, date)
);

-- API keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_instances_cloud_account_id ON instances(cloud_account_id);
CREATE INDEX idx_instances_provider ON instances(provider);
CREATE INDEX idx_instances_status ON instances(status);
CREATE INDEX idx_instances_tags ON instances USING GIN (tags);
CREATE INDEX idx_instances_enabled ON schedules(enabled);
CREATE INDEX idx_recommendations_instance_id ON recommendations(instance_id);
CREATE INDEX idx_recommendations_status ON recommendations(status);
CREATE INDEX idx_overrides_instance_id ON overrides(instance_id);
CREATE INDEX idx_events_instance_id ON events(instance_id);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_savings_instance_id ON savings(instance_id);
CREATE INDEX idx_savings_date ON savings(date);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_instances_updated_at BEFORE UPDATE ON instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE instances IS 'Discovered database instances from cloud providers';
COMMENT ON TABLE schedules IS 'Sleep/wake schedules with dynamic instance selection';
COMMENT ON TABLE settings IS 'Application configuration and thresholds';
COMMENT ON TABLE recommendations IS 'AI-generated schedule suggestions';
COMMENT ON TABLE overrides IS 'Temporary manual overrides for instances';
COMMENT ON TABLE events IS 'Audit log of all database state changes';
COMMENT ON TABLE savings IS 'Daily aggregated cost savings';
