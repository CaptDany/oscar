-- Function to remove pipeline for a tenant
CREATE OR REPLACE FUNCTION unseed_tenant_pipeline(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    DELETE FROM pipeline_stages WHERE pipeline_id IN (
        SELECT id FROM pipelines WHERE tenant_id = tenant_uuid AND is_default = true
    );
    DELETE FROM pipelines WHERE tenant_id = tenant_uuid AND is_default = true;
END;
$$ LANGUAGE plpgsql;
