-- Add email verification support and API rate limiting

-- Add verification column to users table
ALTER TABLE users 
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN email_verified_at TIMESTAMP;

-- Create email_verifications table (similar to password_resets)
CREATE TABLE email_verifications (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hashed_token TEXT NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

-- Create API call logs table for rate limiting
-- This tracks individual API calls to enforce daily/monthly limits
CREATE TABLE api_call_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_email_verifications_hashed_token ON email_verifications(hashed_token);
CREATE INDEX idx_api_call_logs_user_date ON api_call_logs(user_id, created_at DESC);

-- Comments
COMMENT ON COLUMN users.email_verified IS 'Whether the user has verified their email address';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified';
COMMENT ON TABLE email_verifications IS 'Email verification tokens (valid for 24 hours)';
COMMENT ON TABLE api_call_logs IS 'Tracks API calls for rate limiting (unverified: 10/day, verified: 100/month)';

