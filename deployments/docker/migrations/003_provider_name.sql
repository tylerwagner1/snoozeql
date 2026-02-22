-- Add provider_name column to instances table
-- The provider_name stores the full provider identifier (e.g., "aws_{accountID}_{region}")
-- while provider stores just the cloud provider type ("aws" or "gcp")

ALTER TABLE instances
ADD COLUMN provider_name VARCHAR(255);

-- Update existing instances to use provider as provider_name for compatibility
UPDATE instances SET provider_name = provider WHERE provider_name IS NULL;

-- Allow NULL for provider_name to support instances without provider information
ALTER TABLE instances
ALTER COLUMN provider_name DROP NOT NULL;
