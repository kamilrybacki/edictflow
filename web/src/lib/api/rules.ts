import { Rule, TargetLayer } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchRules(teamId: string, status?: string): Promise<Rule[]> {
  let url = `${getApiUrlCached()}/api/v1/rules/?team_id=${teamId}`;
  if (status) {
    url += `&status=${status}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch rules: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch rule: ${res.statusText}`);
  }
  return res.json();
}

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

export async function deleteRule(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete rule: ${res.statusText}`);
  }
}

export async function getMergedContent(level: TargetLayer): Promise<string> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/merged?level=${level}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to get merged content: ${res.statusText}`);
  }
  return res.text();
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
