-- Hourly aggregated metrics from CloudWatch
CREATE TABLE metrics_hourly (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    metric_name VARCHAR(50) NOT NULL,  -- CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS
    hour TIMESTAMPTZ NOT NULL,          -- Truncated to hour in UTC
    avg_value FLOAT NOT NULL,
    max_value FLOAT NOT NULL,
    min_value FLOAT NOT NULL,
    sample_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(instance_id, metric_name, hour)
);

-- Index for efficient time-range queries per instance
CREATE INDEX idx_metrics_hourly_instance_time ON metrics_hourly(instance_id, hour DESC);
-- Index for retention cleanup queries
CREATE INDEX idx_metrics_hourly_hour ON metrics_hourly(hour DESC);

-- Update trigger for updated_at
CREATE TRIGGER update_metrics_hourly_updated_at BEFORE UPDATE ON metrics_hourly
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE metrics_hourly IS 'Hourly aggregated CloudWatch metrics for activity analysis';
