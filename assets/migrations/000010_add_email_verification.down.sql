-- Rollback email verification and API rate limiting

DROP TABLE IF EXISTS api_call_logs;
DROP TABLE IF EXISTS email_verifications;

ALTER TABLE users 
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS email_verified_at;

