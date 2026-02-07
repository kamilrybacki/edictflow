import { Rule } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export interface CreateRuleRequest {
  name: string;
  content: string;
  description?: string;
  target_layer: string;
  category_id?: string;
  priority_weight?: number;
  overridable?: boolean;
  effective_start?: string;
  effective_end?: string;
  target_teams?: string[];
  target_users?: string[];
  tags?: string[];
  team_id: string;
  triggers: Array<{
    type: string;
    pattern?: string;
    context_types?: string[];
    tags?: string[];
  }>;
  enforcement_mode?: string;
  temporary_timeout_hours?: number;
}

export async function createRule(request: CreateRuleRequest): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error(`Failed to create rule: ${res.statusText}`);
  }
  return res.json();
}

export async function updateRule(id: string, request: CreateRuleRequest): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error(`Failed to update rule: ${res.statusText}`);
  }
}

export async function fetchGlobalRules(): Promise<Rule[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/global`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch global rules');
  const data = await res.json();
  return data || [];
}

export async function createGlobalRule(data: {
  name: string;
  content: string;
  description?: string;
  force: boolean;
}): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/global`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create global rule');
  return res.json();
}
