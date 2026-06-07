CREATE POLICY tenant_isolation_deal_line_items ON deal_line_items
    USING (
        EXISTS (
            SELECT 1 FROM deals d
            WHERE d.id = deal_line_items.deal_id
            AND d.tenant_id::TEXT = current_setting('app.current_tenant', true)
        )
    );
