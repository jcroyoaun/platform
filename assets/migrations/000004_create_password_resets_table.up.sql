CREATE TABLE password_resets (
    hashed_token TEXT NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry TIMESTAMP NOT NULL
);

CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
