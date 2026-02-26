-- 5-minute granularity metrics from CloudWatch for detailed time-range views
CREATE TABLE metrics_5min (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    metric_name VARCHAR(50) NOT NULL,
    minute TIMESTAMPTZ NOT NULL,
    avg_value FLOAT NOT NULL,
    max_value FLOAT NOT NULL,
    min_value FLOAT NOT NULL,
    sample_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(instance_id, metric_name, minute)
);

CREATE INDEX idx_metrics_5min_instance_time ON metrics_5min(instance_id, minute DESC);
CREATE INDEX idx_metrics_5min_minute ON metrics_5min(minute DESC);

CREATE TRIGGER update_metrics_5min_updated_at BEFORE UPDATE ON metrics_5min
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE metrics_5min IS '5-minute granularity CloudWatch metrics for 1-hour view';
