-- Add connection status tracking to cloud_accounts table
ALTER TABLE cloud_accounts ADD COLUMN connection_status VARCHAR(50) DEFAULT 'unknown';
ALTER TABLE cloud_accounts ADD COLUMN last_sync_at TIMESTAMPTZ;
ALTER TABLE cloud_accounts ADD COLUMN last_error TEXT;
