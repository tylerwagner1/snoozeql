-- Cost tracking indexes and materialized view for savings queries

-- Index for time-range queries on events (for finding last stop event per instance)
CREATE INDEX IF NOT EXISTS idx_events_instance_time ON events(instance_id, created_at DESC);
-- Index for filtering by event type
CREATE INDEX IF NOT EXISTS idx_events_type_time ON events(event_type, created_at);

-- Add hourly_rate_cents column to savings table
-- Stores rate at time of calculation (AUD-02 requires storing rate for accurate reporting)
ALTER TABLE savings ADD COLUMN IF NOT EXISTS hourly_rate_cents INTEGER;

-- Create materialized view for dashboard performance
-- Aggregates savings by day/week/month for quick querying
CREATE MATERIALIZED VIEW IF NOT EXISTS savings_summary AS
SELECT 
    instance_id,
    DATE_TRUNC('day', date) as day,
    DATE_TRUNC('week', date) as week,
    DATE_TRUNC('month', date) as month,
    SUM(stopped_minutes) as total_stopped_minutes,
    SUM(estimated_savings_cents) as total_savings_cents
FROM savings
GROUP BY instance_id, DATE_TRUNC('day', date), DATE_TRUNC('week', date), DATE_TRUNC('month', date);

-- Unique index on materialized view for efficient daily lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_savings_summary_instance_day ON savings_summary(instance_id, day);

-- Comments for documentation
COMMENT ON INDEX idx_events_instance_time IS 'For finding last stop event per instance in time-range queries';
COMMENT ON INDEX idx_events_type_time IS 'For filtering events by type';
COMMENT ON COLUMN savings.hourly_rate_cents IS 'Hourly rate cents at time of calculation for accurate savings reporting';
COMMENT ON MATERIALIZED VIEW savings_summary IS 'Aggregated savings data for dashboard performance';
COMMENT ON INDEX idx_savings_summary_instance_day IS 'Unique index for efficient daily savings lookups by instance';
