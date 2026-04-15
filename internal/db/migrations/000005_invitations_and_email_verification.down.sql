DROP TABLE IF EXISTS invitations;

ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verification_token;
ALTER TABLE users DROP COLUMN IF EXISTS email_verification_sent_at;
ALTER TABLE users DROP COLUMN IF EXISTS pending_invitation_id;

ALTER TABLE tenants DROP COLUMN IF EXISTS invite_only;
