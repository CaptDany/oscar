DROP TABLE IF EXISTS oauth_users;
ALTER TABLE tenants DROP COLUMN IF EXISTS require_email_verification;
