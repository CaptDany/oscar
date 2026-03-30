-- Rollback: Remove composite index for cursor-based pagination
DROP INDEX IF EXISTS idx_companies_tenant_cursor;
