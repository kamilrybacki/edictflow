import { Rule } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function approveRule(ruleId: string, comment?: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ comment }),
  });
  if (!res.ok) {
    throw new Error(`Failed to approve rule: ${res.statusText}`);
  }
}

export async function rejectRule(ruleId: string, comment: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ comment }),
  });
  if (!res.ok) {
    throw new Error(`Failed to reject rule: ${res.statusText}`);
  }
}

export interface ApprovalStatus {
  rule_id: string;
  status: string;
  required_count: number;
  current_count: number;
  approvals: Array<{
    id: string;
    user_id: string;
    user_name?: string;
    decision: string;
    comment?: string;
    created_at: string;
  }>;
}

export async function getApprovalStatus(ruleId: string): Promise<ApprovalStatus> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to get approval status: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchPendingApprovals(scope?: string): Promise<Rule[]> {
  let url = `${getApiUrlCached()}/api/v1/approvals/pending`;
  if (scope) {
    url += `?scope=${scope}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch pending approvals: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}
