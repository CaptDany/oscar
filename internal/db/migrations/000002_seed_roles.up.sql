-- Seed built-in roles for each tenant

-- Function to create default roles for a tenant
CREATE OR REPLACE FUNCTION seed_tenant_roles(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    -- Owner role - all permissions
    INSERT INTO roles (tenant_id, name, description, is_system, permissions)
    VALUES (
        tenant_uuid,
        'Owner',
        'Full access to all features and settings',
        true,
        '{
            "persons": {"view": "all", "create": "all", "edit": "all", "delete": "all", "export": "all"},
            "companies": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "deals": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "activities": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "products": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "pipelines": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "reports": {"view": "all"},
            "settings": {"view": "all", "edit": "all"},
            "users": {"view": "all", "invite": "all", "edit": "all", "delete": "all"},
            "roles": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "teams": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "automations": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "audit_logs": {"view": "all"},
            "api_keys": {"view": "all", "create": "all", "delete": "all"}
        }'::jsonb
    );

    -- Admin role - all except delete users/roles and billing
    INSERT INTO roles (tenant_id, name, description, is_system, permissions)
    VALUES (
        tenant_uuid,
        'Admin',
        'Administrative access without billing or owner deletion',
        true,
        '{
            "persons": {"view": "all", "create": "all", "edit": "all", "delete": "all", "export": "all"},
            "companies": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "deals": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "activities": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "products": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "pipelines": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "reports": {"view": "all"},
            "settings": {"view": "all", "edit": "all"},
            "users": {"view": "all", "invite": "all", "edit": "all", "delete": "none"},
            "roles": {"view": "all", "create": "all", "edit": "all", "delete": "none"},
            "teams": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "automations": {"view": "all", "create": "all", "edit": "all", "delete": "all"},
            "audit_logs": {"view": "all"},
            "api_keys": {"view": "all", "create": "all", "delete": "all"}
        }'::jsonb
    );

    -- Manager role - team-wide visibility
    INSERT INTO roles (tenant_id, name, description, is_system, permissions)
    VALUES (
        tenant_uuid,
        'Manager',
        'Team management with team-wide visibility',
        true,
        '{
            "persons": {"view": "team", "create": "team", "edit": "team", "delete": "own", "export": "team"},
            "companies": {"view": "team", "create": "team", "edit": "team", "delete": "own"},
            "deals": {"view": "team", "create": "team", "edit": "team", "delete": "own"},
            "activities": {"view": "team", "create": "team", "edit": "team", "delete": "own"},
            "products": {"view": "team", "create": "team", "edit": "team", "delete": "none"},
            "pipelines": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "reports": {"view": "team"},
            "settings": {"view": "none", "edit": "none"},
            "users": {"view": "none", "invite": "none", "edit": "none", "delete": "none"},
            "roles": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "teams": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "automations": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "audit_logs": {"view": "none"},
            "api_keys": {"view": "none", "create": "none", "delete": "none"}
        }'::jsonb
    );

    -- Sales Rep role - own records only
    INSERT INTO roles (tenant_id, name, description, is_system, permissions)
    VALUES (
        tenant_uuid,
        'Sales Rep',
        'Standard sales representative access',
        true,
        '{
            "persons": {"view": "team", "create": "own", "edit": "own", "delete": "none", "export": "own"},
            "companies": {"view": "team", "create": "own", "edit": "own", "delete": "none"},
            "deals": {"view": "team", "create": "own", "edit": "own", "delete": "none"},
            "activities": {"view": "team", "create": "own", "edit": "own", "delete": "own"},
            "products": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "pipelines": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "reports": {"view": "own"},
            "settings": {"view": "none", "edit": "none"},
            "users": {"view": "none", "invite": "none", "edit": "none", "delete": "none"},
            "roles": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "teams": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "automations": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "audit_logs": {"view": "none"},
            "api_keys": {"view": "none", "create": "none", "delete": "none"}
        }'::jsonb
    );

    -- Read Only role
    INSERT INTO roles (tenant_id, name, description, is_system, permissions)
    VALUES (
        tenant_uuid,
        'Read Only',
        'View access to all records',
        true,
        '{
            "persons": {"view": "team", "create": "none", "edit": "none", "delete": "none", "export": "none"},
            "companies": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "deals": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "activities": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "products": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "pipelines": {"view": "team", "create": "none", "edit": "none", "delete": "none"},
            "reports": {"view": "none"},
            "settings": {"view": "none", "edit": "none"},
            "users": {"view": "none", "invite": "none", "edit": "none", "delete": "none"},
            "roles": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "teams": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "automations": {"view": "none", "create": "none", "edit": "none", "delete": "none"},
            "audit_logs": {"view": "none"},
            "api_keys": {"view": "none", "create": "none", "delete": "none"}
        }'::jsonb
    );
END;
$$ LANGUAGE plpgsql;
