CREATE TABLE IF NOT EXISTS pastes (
    id TEXT PRIMARY KEY,
    content TEXT,
    language TEXT NOT NULL DEFAULT 'plaintext',
    password_hash TEXT,
    salt BYTEA,
    encrypted_iv BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);
