-- Add password reset fields to users table (only if they don't exist)
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS force_password_reset BOOLEAN DEFAULT 0;
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS password_reset_required_at TIMESTAMP;

-- Create password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON password_reset_tokens(token);
