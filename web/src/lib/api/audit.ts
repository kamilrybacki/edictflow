import { getApiUrlCached, getAuthHeaders } from './http';

export interface AuditEntry {
  id: string;
  entity_type: string;
  entity_id: string;
  action: string;
  actor_id?: string;
  actor_name?: string;
  changes?: Record<string, { old: unknown; new: unknown }>;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditListResponse {
  entries: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}

export async function fetchAuditLogs(params: {
  entity_type?: string;
  action?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}): Promise<AuditListResponse> {
  const searchParams = new URLSearchParams();
  if (params.entity_type) searchParams.set('entity_type', params.entity_type);
  if (params.action) searchParams.set('action', params.action);
  if (params.from) searchParams.set('from', params.from);
  if (params.to) searchParams.set('to', params.to);
  if (params.limit) searchParams.set('limit', params.limit.toString());
  if (params.offset) searchParams.set('offset', params.offset.toString());

  const res = await fetch(`${getApiUrlCached()}/api/v1/audit/?${searchParams}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch audit logs: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchEntityHistory(entityType: string, entityId: string): Promise<AuditEntry[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/audit/entity/${entityType}/${entityId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch entity history: ${res.statusText}`);
  }
  return res.json();
}
