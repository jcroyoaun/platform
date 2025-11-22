{! if eq .Database "mysql" !}DROP INDEX idx_sessions_expiry ON sessions;{! else !}DROP INDEX idx_sessions_expiry;{! end !}

DROP TABLE sessions;