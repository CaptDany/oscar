-- Function to remove roles for a tenant (for cleanup)
CREATE OR REPLACE FUNCTION unseed_tenant_roles(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    DELETE FROM user_roles WHERE role_id IN (
        SELECT id FROM roles WHERE tenant_id = tenant_uuid AND is_system = true
    );
    DELETE FROM roles WHERE tenant_id = tenant_uuid AND is_system = true;
END;
$$ LANGUAGE plpgsql;
