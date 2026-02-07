import { RuleAttachment, EnforcementMode } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchTeamAttachments(teamId: string): Promise<RuleAttachment[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${teamId}/attachments`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch team attachments');
  return res.json() || [];
}

export interface CreateAttachmentRequest {
  rule_id: string;
  enforcement_mode: EnforcementMode;
  temporary_timeout_hours?: number;
}

export async function createAttachment(teamId: string, request: CreateAttachmentRequest): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${teamId}/attachments`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) throw new Error('Failed to create attachment');
  return res.json();
}

export async function fetchAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to fetch attachment');
  return res.json();
}

export async function updateAttachment(id: string, enforcementMode: EnforcementMode, timeoutHours?: number): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      enforcement_mode: enforcementMode,
      temporary_timeout_hours: timeoutHours,
    }),
  });
  if (!res.ok) throw new Error('Failed to update attachment');
  return res.json();
}

export async function deleteAttachment(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to delete attachment');
}

export async function approveAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to approve attachment');
  return res.json();
}

export async function rejectAttachment(id: string): Promise<RuleAttachment> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/attachments/${id}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) throw new Error('Failed to reject attachment');
  return res.json();
}
