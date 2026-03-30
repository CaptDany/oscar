-- Add composite index for cursor-based pagination on companies
CREATE INDEX idx_companies_tenant_cursor ON companies(tenant_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
