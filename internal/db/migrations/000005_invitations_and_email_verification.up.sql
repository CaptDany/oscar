-- Invitations table for invite-only registration
CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role_name VARCHAR(100) DEFAULT 'Member',
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_tenant_id ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(email);

-- Add email verification fields to users
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN email_verification_token VARCHAR(64);
ALTER TABLE users ADD COLUMN email_verification_sent_at TIMESTAMPTZ;

-- Add invitation-only flag to tenants
ALTER TABLE tenants ADD COLUMN invite_only BOOLEAN NOT NULL DEFAULT false;

-- Add pending_invitation field for users created via invitation
ALTER TABLE users ADD COLUMN pending_invitation_id UUID REFERENCES invitations(id) ON DELETE SET NULL;
