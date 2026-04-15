-- OAuth users table
CREATE TABLE oauth_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id),
    UNIQUE(provider, user_id)
);

CREATE INDEX idx_oauth_users_user_id ON oauth_users(user_id);
CREATE INDEX idx_oauth_users_provider ON oauth_users(provider);

-- Add email verification required flag to tenants
ALTER TABLE tenants ADD COLUMN require_email_verification BOOLEAN NOT NULL DEFAULT true;
