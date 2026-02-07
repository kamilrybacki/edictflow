import { Rule } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchLibraryRules(): Promise<Rule[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch library rules');
  return res.json() || [];
}

export async function fetchLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch library rule');
  return res.json();
}

export interface CreateLibraryRuleRequest {
  name: string;
  content: string;
  description?: string;
  target_layer: string;
  category_id?: string;
  priority_weight?: number;
  overridable?: boolean;
  tags?: string[];
  triggers?: Array<{
    type: string;
    pattern?: string;
    context_types?: string[];
    tags?: string[];
  }>;
}

export async function createLibraryRule(request: CreateLibraryRuleRequest): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to create library rule');
  return res.json();
}

export async function updateLibraryRule(id: string, request: CreateLibraryRuleRequest): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to update library rule');
}

export async function deleteLibraryRule(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete library rule');
}

export async function submitLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/submit`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to submit library rule');
  return res.json();
}

export async function approveLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to approve library rule');
  return res.json();
}

export async function rejectLibraryRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/library/rules/${id}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to reject library rule');
  return res.json();
}
