-- Drop triggers first
DROP TRIGGER IF EXISTS notifications_updated_at ON notifications;
DROP TRIGGER IF EXISTS automation_actions_updated_at ON automation_actions;
DROP TRIGGER IF EXISTS automations_updated_at ON automations;
DROP TRIGGER IF EXISTS custom_field_definitions_updated_at ON custom_field_definitions;
DROP TRIGGER IF EXISTS activities_updated_at ON activities;
DROP TRIGGER IF EXISTS deal_line_items_updated_at ON deal_line_items;
DROP TRIGGER IF EXISTS products_updated_at ON products;
DROP TRIGGER IF EXISTS deals_updated_at ON deals;
DROP TRIGGER IF EXISTS pipeline_stages_updated_at ON pipeline_stages;
DROP TRIGGER IF EXISTS pipelines_updated_at ON pipelines;
DROP TRIGGER IF EXISTS persons_updated_at ON persons;
DROP TRIGGER IF EXISTS companies_updated_at ON companies;
DROP TRIGGER IF EXISTS api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS teams_updated_at ON teams;
DROP TRIGGER IF EXISTS roles_updated_at ON roles;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TRIGGER IF EXISTS tenant_branding_updated_at ON tenant_branding;
DROP TRIGGER IF EXISTS tenants_updated_at ON tenants;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation_notifications ON notifications;
DROP POLICY IF EXISTS tenant_isolation_automations ON automations;
DROP POLICY IF EXISTS tenant_isolation_custom_fields ON custom_field_definitions;
DROP POLICY IF EXISTS tenant_isolation_activities ON activities;
DROP POLICY IF EXISTS tenant_isolation_products ON products;
DROP POLICY IF EXISTS tenant_isolation_deals ON deals;
DROP POLICY IF EXISTS tenant_isolation_pipeline_stages ON pipeline_stages;
DROP POLICY IF EXISTS tenant_isolation_pipelines ON pipelines;
DROP POLICY IF EXISTS tenant_isolation_persons ON persons;
DROP POLICY IF EXISTS tenant_isolation_companies ON companies;
DROP POLICY IF EXISTS tenant_isolation_audit_logs ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_api_keys ON api_keys;
DROP POLICY IF EXISTS tenant_isolation_teams ON teams;
DROP POLICY IF EXISTS tenant_isolation_roles ON roles;
DROP POLICY IF EXISTS tenant_isolation_users ON users;

-- Disable RLS
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE automation_run_actions DISABLE ROW LEVEL SECURITY;
ALTER TABLE automation_runs DISABLE ROW LEVEL SECURITY;
ALTER TABLE automation_actions DISABLE ROW LEVEL SECURITY;
ALTER TABLE automations DISABLE ROW LEVEL SECURITY;
ALTER TABLE custom_field_definitions DISABLE ROW LEVEL SECURITY;
ALTER TABLE activity_associations DISABLE ROW LEVEL SECURITY;
ALTER TABLE activities DISABLE ROW LEVEL SECURITY;
ALTER TABLE deal_line_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE products DISABLE ROW LEVEL SECURITY;
ALTER TABLE deals DISABLE ROW LEVEL SECURITY;
ALTER TABLE pipeline_stages DISABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines DISABLE ROW LEVEL SECURITY;
ALTER TABLE persons DISABLE ROW LEVEL SECURITY;
ALTER TABLE companies DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys DISABLE ROW LEVEL SECURITY;
ALTER TABLE team_members DISABLE ROW LEVEL SECURITY;
ALTER TABLE teams DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_branding DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS automation_run_actions;
DROP TABLE IF EXISTS automation_runs;
DROP TABLE IF EXISTS automation_actions;
DROP TABLE IF EXISTS automations;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS custom_field_definitions;
DROP TABLE IF EXISTS activity_associations;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS deal_line_items;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS pipeline_stages;
DROP TABLE IF EXISTS pipelines;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenant_branding;
DROP TABLE IF EXISTS tenants;

-- Drop functions
DROP FUNCTION IF EXISTS set_tenant_context(UUID);
DROP FUNCTION IF EXISTS get_current_tenant();
DROP FUNCTION IF EXISTS set_updated_at();

-- Drop enums
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS automation_run_status;
DROP TYPE IF EXISTS automation_action_type;
DROP TYPE IF EXISTS automation_trigger_type;
DROP TYPE IF EXISTS custom_field_type;
DROP TYPE IF EXISTS entity_type;
DROP TYPE IF EXISTS pipeline_stage_type;
DROP TYPE IF EXISTS activity_direction;
DROP TYPE IF EXISTS activity_status;
DROP TYPE IF EXISTS activity_type;
DROP TYPE IF EXISTS company_size;
DROP TYPE IF EXISTS person_source;
DROP TYPE IF EXISTS person_status;
DROP TYPE IF EXISTS person_type;
DROP TYPE IF EXISTS subscription_tier;
DROP TYPE IF EXISTS tenant_status;

-- Drop schema
DROP SCHEMA IF EXISTS app;
