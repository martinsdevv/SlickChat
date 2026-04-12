CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    discriminator TEXT NOT NULL,
    password_hash TEXT,
    recovery_key_hash TEXT,
    paranoid_mode BOOL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (username, discriminator)
);

CREATE INDEX idx_users_handle ON users(username, discriminator);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_hash TEXT
);

CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
