-- Rollback API Key support

DROP INDEX IF EXISTS idx_users_api_key;

ALTER TABLE users 
DROP COLUMN IF EXISTS api_key,
DROP COLUMN IF EXISTS api_calls_count,
DROP COLUMN IF EXISTS api_key_created_at;

