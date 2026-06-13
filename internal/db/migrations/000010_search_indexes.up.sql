CREATE INDEX idx_deals_search ON deals USING gin(to_tsvector('english', COALESCE(title, '')));
