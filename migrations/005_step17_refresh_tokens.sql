CREATE TABLE IF NOT EXISTS refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id int NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    user_agent text,
    ip_address text,
    expires_at timestamp NOT NULL,
    revoked_at timestamp,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rotated_from uuid
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
