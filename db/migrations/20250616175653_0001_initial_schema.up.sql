CREATE TABLE IF NOT EXISTS pastes (
    id VARCHAR(10) PRIMARY KEY,
    content TEXT NOT NULL,
    language VARCHAR(50) DEFAULT 'plaintext',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    password_hash VARCHAR(255) NULL,          -- stores bcrypt hash of password
    salt BYTEA NULL,                          -- Salt for key derivation
    encrypted_iv BYTEA NULL                   -- IV for encryption
);