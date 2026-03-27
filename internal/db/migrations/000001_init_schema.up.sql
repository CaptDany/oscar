-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create custom error codes schema
CREATE SCHEMA IF NOT EXISTS app;

-- =============================================================================
-- ENUMS
-- =============================================================================

CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'deleted');
CREATE TYPE subscription_tier AS ENUM ('free', 'starter', 'professional', 'enterprise');
CREATE TYPE person_type AS ENUM ('lead', 'contact', 'customer');
CREATE TYPE person_status AS ENUM ('new', 'contacted', 'qualified', 'unqualified', 'active', 'inactive');
CREATE TYPE person_source AS ENUM ('website', 'referral', 'social', 'email', 'phone', 'event', 'other');
CREATE TYPE company_size AS ENUM ('startup', 'small', 'medium', 'large', 'enterprise');
CREATE TYPE activity_type AS ENUM ('note', 'call', 'email', 'meeting', 'task', 'whatsapp', 'sms');
CREATE TYPE activity_status AS ENUM ('planned', 'completed', 'cancelled');
CREATE TYPE activity_direction AS ENUM ('inbound', 'outbound');
CREATE TYPE pipeline_stage_type AS ENUM ('open', 'won', 'lost');
CREATE TYPE entity_type AS ENUM ('person', 'company', 'deal');
CREATE TYPE custom_field_type AS ENUM ('text', 'number', 'date', 'select', 'multi_select', 'boolean', 'url', 'currency');
CREATE TYPE automation_trigger_type AS ENUM (
    'person.created', 'person.updated', 'person.converted', 'person.score_changed', 'person.assigned',
    'deal.created', 'deal.updated', 'deal.stage_changed', 'deal.won', 'deal.lost', 'deal.close_date_passed',
    'activity.created', 'activity.completed',
    'company.created', 'company.updated'
);
CREATE TYPE automation_action_type AS ENUM (
    'create_task', 'send_email', 'update_field', 'add_tag', 'remove_tag',
    'assign_owner', 'move_stage', 'convert_person', 'send_notification', 'webhook', 'send_sms'
);
CREATE TYPE automation_run_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE permission_scope AS ENUM ('none', 'own', 'team', 'all');

-- =============================================================================
-- TENANT & BRANDING
-- =============================================================================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(63) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    status tenant_status NOT NULL DEFAULT 'active',
    subscription_tier subscription_tier NOT NULL DEFAULT 'free',
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);

CREATE TABLE tenant_branding (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    logo_light_url TEXT,
    logo_dark_url TEXT,
    favicon_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#6366f1',
    secondary_color VARCHAR(7) DEFAULT '#8b5cf6',
    accent_color VARCHAR(7) DEFAULT '#06b6d4',
    font_family VARCHAR(255) DEFAULT 'Inter, system-ui, sans-serif',
    app_name VARCHAR(255) DEFAULT 'OpenCRM',
    custom_css TEXT,
    email_header_html TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- USERS & AUTHENTICATION
-- =============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    timezone VARCHAR(50) DEFAULT 'UTC',
    locale VARCHAR(10) DEFAULT 'en',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_active ON users(tenant_id, is_active) WHERE deleted_at IS NULL;

-- =============================================================================
-- ROLES & PERMISSIONS
-- =============================================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    permissions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- =============================================================================
-- TEAMS
-- =============================================================================

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_teams_tenant_id ON teams(tenant_id);

CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_lead BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- =============================================================================
-- API KEYS
-- =============================================================================

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- =============================================================================
-- AUDIT LOGS
-- =============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    diff JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- =============================================================================
-- COMPANIES
-- =============================================================================

CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    industry VARCHAR(100),
    size company_size,
    annual_revenue NUMERIC(15, 2),
    website TEXT,
    address JSONB DEFAULT '{}',
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    parent_company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    tags TEXT[] DEFAULT '{}',
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_companies_tenant_id ON companies(tenant_id);
CREATE INDEX idx_companies_owner_id ON companies(owner_id);
CREATE INDEX idx_companies_deleted_at ON companies(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_companies_search ON companies USING gin(to_tsvector('english', name || ' ' || COALESCE(domain, '')));

-- =============================================================================
-- PERSONS (LEADS, CONTACTS, CUSTOMERS)
-- =============================================================================

CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type person_type NOT NULL DEFAULT 'lead',
    status person_status NOT NULL DEFAULT 'new',
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email TEXT[] DEFAULT '{}',
    phone TEXT[] DEFAULT '{}',
    avatar_url TEXT,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    source person_source,
    score INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    custom_fields JSONB DEFAULT '{}',
    converted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_persons_tenant_id ON persons(tenant_id);
CREATE INDEX idx_persons_type_status ON persons(tenant_id, type, status);
CREATE INDEX idx_persons_owner_id ON persons(owner_id);
CREATE INDEX idx_persons_company_id ON persons(company_id);
CREATE INDEX idx_persons_deleted_at ON persons(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_search ON persons USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(array_to_string(email, ' '), '')));

-- =============================================================================
-- PIPELINES & STAGES
-- =============================================================================

CREATE TABLE pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_pipelines_tenant_id ON pipelines(tenant_id);
CREATE INDEX idx_pipelines_is_default ON pipelines(is_default) WHERE is_default = true;

CREATE TABLE pipeline_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    probability INTEGER NOT NULL DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),
    stage_type pipeline_stage_type NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(pipeline_id, name),
    UNIQUE(pipeline_id, position)
);

CREATE INDEX idx_pipeline_stages_pipeline_id ON pipeline_stages(pipeline_id);
CREATE INDEX idx_pipeline_stages_position ON pipeline_stages(pipeline_id, position);

-- =============================================================================
-- DEALS
-- =============================================================================

CREATE TABLE deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    value NUMERIC(15, 2) NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    stage_id UUID REFERENCES pipeline_stages(id) ON DELETE SET NULL,
    pipeline_id UUID REFERENCES pipelines(id) ON DELETE SET NULL,
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    expected_close_date DATE,
    closed_at TIMESTAMPTZ,
    won_reason TEXT,
    lost_reason TEXT,
    probability INTEGER DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),
    tags TEXT[] DEFAULT '{}',
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_deals_tenant_id ON deals(tenant_id);
CREATE INDEX idx_deals_stage_id ON deals(stage_id);
CREATE INDEX idx_deals_pipeline_id ON deals(pipeline_id);
CREATE INDEX idx_deals_owner_id ON deals(owner_id);
CREATE INDEX idx_deals_person_id ON deals(person_id);
CREATE INDEX idx_deals_company_id ON deals(company_id);
CREATE INDEX idx_deals_deleted_at ON deals(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_expected_close ON deals(expected_close_date) WHERE deleted_at IS NULL AND closed_at IS NULL;

-- =============================================================================
-- PRODUCTS
-- =============================================================================

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(100),
    price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    unit VARCHAR(50) DEFAULT 'unit',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_tenant_id ON products(tenant_id);
CREATE INDEX idx_products_sku ON products(sku) WHERE sku IS NOT NULL;

CREATE TABLE deal_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    quantity NUMERIC(10, 2) NOT NULL DEFAULT 1,
    unit_price NUMERIC(15, 2) NOT NULL,
    discount_pct NUMERIC(5, 2) NOT NULL DEFAULT 0,
    total NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deal_line_items_deal_id ON deal_line_items(deal_id);

-- =============================================================================
-- ACTIVITIES
-- =============================================================================

CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type activity_type NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body TEXT,
    outcome TEXT,
    direction activity_direction,
    status activity_status NOT NULL DEFAULT 'planned',
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_activities_tenant_id ON activities(tenant_id);
CREATE INDEX idx_activities_owner_id ON activities(owner_id);
CREATE INDEX idx_activities_status ON activities(status) WHERE status = 'planned';
CREATE INDEX idx_activities_due_at ON activities(due_at) WHERE status = 'planned' AND due_at IS NOT NULL;
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE activity_associations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    entity_type entity_type NOT NULL,
    entity_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(activity_id, entity_type, entity_id)
);

CREATE INDEX idx_activity_associations_entity ON activity_associations(entity_type, entity_id);

-- =============================================================================
-- CUSTOM FIELD DEFINITIONS
-- =============================================================================

CREATE TABLE custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type entity_type NOT NULL,
    field_key VARCHAR(100) NOT NULL,
    label VARCHAR(255) NOT NULL,
    field_type custom_field_type NOT NULL DEFAULT 'text',
    options JSONB DEFAULT '[]',
    is_required BOOLEAN NOT NULL DEFAULT false,
    show_in_list BOOLEAN NOT NULL DEFAULT false,
    show_in_card BOOLEAN NOT NULL DEFAULT true,
    position INTEGER NOT NULL DEFAULT 0,
    role_visibility TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, entity_type, field_key)
);

CREATE INDEX idx_custom_field_definitions_entity ON custom_field_definitions(tenant_id, entity_type);

-- =============================================================================
-- AUTOMATIONS
-- =============================================================================

CREATE TABLE automations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    trigger_type automation_trigger_type NOT NULL,
    trigger_config JSONB NOT NULL DEFAULT '{}',
    conditions JSONB NOT NULL DEFAULT '{"operator": "AND", "rules": []}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automations_tenant_id ON automations(tenant_id);
CREATE INDEX idx_automations_trigger ON automations(trigger_type) WHERE is_active = true;

CREATE TABLE automation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    action_type automation_action_type NOT NULL,
    action_config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_actions_automation ON automation_actions(automation_id);

CREATE TABLE automation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    trigger_entity_type entity_type,
    trigger_entity_id UUID,
    status automation_run_status NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_runs_automation ON automation_runs(automation_id);
CREATE INDEX idx_automation_runs_status ON automation_runs(status);

CREATE TABLE automation_run_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES automation_runs(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES automation_actions(id) ON DELETE CASCADE,
    status automation_run_status NOT NULL DEFAULT 'pending',
    result JSONB,
    executed_at TIMESTAMPTZ,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_run_actions_run ON automation_run_actions(run_id);

-- =============================================================================
-- NOTIFICATIONS
-- =============================================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    entity_type entity_type,
    entity_id UUID,
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;

-- =============================================================================
-- HELPER FUNCTIONS & TRIGGERS
-- =============================================================================

-- Function to set the current tenant context
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant', tenant_uuid::TEXT, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get the current tenant from context
CREATE OR REPLACE FUNCTION get_current_tenant() 
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant', true), '')::UUID;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Generic updated_at trigger function
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers to all tables with updated_at column
CREATE TRIGGER tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER tenant_branding_updated_at BEFORE UPDATE ON tenant_branding
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER companies_updated_at BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER persons_updated_at BEFORE UPDATE ON persons
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER pipelines_updated_at BEFORE UPDATE ON pipelines
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER pipeline_stages_updated_at BEFORE UPDATE ON pipeline_stages
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER deals_updated_at BEFORE UPDATE ON deals
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER deal_line_items_updated_at BEFORE UPDATE ON deal_line_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER activities_updated_at BEFORE UPDATE ON activities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER custom_field_definitions_updated_at BEFORE UPDATE ON custom_field_definitions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER automations_updated_at BEFORE UPDATE ON automations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER automation_actions_updated_at BEFORE UPDATE ON automation_actions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =============================================================================
-- ROW LEVEL SECURITY
-- =============================================================================

-- Enable RLS on all tenant-scoped tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_branding ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE team_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipelines ENABLE ROW LEVEL SECURITY;
ALTER TABLE pipeline_stages ENABLE ROW LEVEL SECURITY;
ALTER TABLE deals ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE deal_line_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE activity_associations ENABLE ROW LEVEL SECURITY;
ALTER TABLE custom_field_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE automations ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_actions ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_run_actions ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;

-- Create RLS policies (tenant can only see their own data)
-- Note: Policies are permissive by default, we use restrictive (AND) to ensure tenant isolation

CREATE POLICY tenant_isolation_users ON users
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_roles ON roles
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_teams ON teams
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_api_keys ON api_keys
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_audit_logs ON audit_logs
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_companies ON companies
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_persons ON persons
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_pipelines ON pipelines
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_pipeline_stages ON pipeline_stages
    USING (
        EXISTS (
            SELECT 1 FROM pipelines p 
            WHERE p.id = pipeline_stages.pipeline_id 
            AND p.tenant_id::TEXT = current_setting('app.current_tenant', true)
        )
    );

CREATE POLICY tenant_isolation_deals ON deals
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_products ON products
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_activities ON activities
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_custom_fields ON custom_field_definitions
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_automations ON automations
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY tenant_isolation_notifications ON notifications
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));
