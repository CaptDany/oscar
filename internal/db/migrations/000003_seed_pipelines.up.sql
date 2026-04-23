-- Seed default pipeline for new tenants

-- Function to create default pipeline
CREATE OR REPLACE FUNCTION seed_tenant_pipeline(tenant_uuid UUID)
RETURNS VOID AS $$
DECLARE
    pipeline_id UUID;
BEGIN
    -- Create default pipeline
    INSERT INTO pipelines (tenant_id, name, is_default, currency)
    VALUES (tenant_uuid, 'Sales Pipeline', true, 'USD')
    RETURNING id INTO pipeline_id;

    -- Create default stages
    INSERT INTO pipeline_stages (pipeline_id, name, position, probability, stage_type)
    VALUES
        (pipeline_id, 'Prospecting', 1, 20, 'open'),
        (pipeline_id, 'Qualified', 2, 40, 'open'),
        (pipeline_id, 'Negotiation', 3, 60, 'open'),
        (pipeline_id, 'Closed Won', 4, 100, 'won'),
        (pipeline_id, 'Closed Lost', 5, 0, 'lost');
END;
$$ LANGUAGE plpgsql;
