-- Migration: Add missing tables (events, overrides, savings, api_keys)
-- These tables were defined in 001_base_schema.sql but may not have been created

-- Active overrides (keep-alive, skip-next)
CREATE TABLE IF NOT EXISTS overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('keep_alive', 'skip_next')),
    skip_action VARCHAR(50) CHECK (skip_action IN ('start', 'stop')),
    until_time TIMESTAMPTZ,
    reason TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expired BOOLEAN DEFAULT FALSE
);

-- Event log for tracking and cost calculation
CREATE TABLE IF NOT EXISTS events (
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
CREATE TABLE IF NOT EXISTS savings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    stopped_minutes INTEGER NOT NULL,
    estimated_savings_cents INTEGER NOT NULL,
    UNIQUE(instance_id, date)
);

-- API keys
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Indexes for performance (IF NOT EXISTS not supported for indexes, so we use DO block)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_overrides_instance_id') THEN
        CREATE INDEX idx_overrides_instance_id ON overrides(instance_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_events_instance_id') THEN
        CREATE INDEX idx_events_instance_id ON events(instance_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_events_created_at') THEN
        CREATE INDEX idx_events_created_at ON events(created_at);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_savings_instance_id') THEN
        CREATE INDEX idx_savings_instance_id ON savings(instance_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_savings_date') THEN
        CREATE INDEX idx_savings_date ON savings(date);
    END IF;
END $$;
