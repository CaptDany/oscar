export interface AuditLogEntry {
  id: string;
  tenant_id: string;
  user_id?: string;
  user_email?: string;
  first_name?: string;
  last_name?: string;
  action: string;
  entity_type: string;
  entity_id: string;
  diff?: any;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface ListAuditLogsParams {
  entity_type?: string;
  entity_id?: string;
  user_id?: string;
  cursor?: string;
  limit?: number;
}

export interface ListAuditLogsResponse {
  data: AuditLogEntry[];
  meta: {
    total: number;
    next_cursor: string;
  };
}
