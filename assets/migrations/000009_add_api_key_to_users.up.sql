-- Add API Key support to users table for Developer API access

ALTER TABLE users 
ADD COLUMN api_key TEXT UNIQUE,
ADD COLUMN api_calls_count INTEGER DEFAULT 0 NOT NULL,
ADD COLUMN api_key_created_at TIMESTAMP;

-- Create index for fast API key lookups
CREATE INDEX idx_users_api_key ON users(api_key) WHERE api_key IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.api_key IS 'Hashed API key for stateless API authentication';
COMMENT ON COLUMN users.api_calls_count IS 'Total number of API calls made by this user';
COMMENT ON COLUMN users.api_key_created_at IS 'Timestamp when the API key was generated';

