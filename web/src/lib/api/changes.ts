import { ChangeRequest } from '@/domain/change_request';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchChanges(
  teamId: string,
  status?: string
): Promise<ChangeRequest[]> {
  let url = `${getApiUrlCached()}/api/v1/changes/?team_id=${teamId}`;
  if (status) {
    url += `&status=${status}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch changes: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchChange(id: string): Promise<ChangeRequest> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/changes/${id}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch change: ${res.statusText}`);
  }
  return res.json();
}

export async function approveChange(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/changes/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to approve change: ${res.statusText}`);
  }
}

export async function rejectChange(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/changes/${id}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to reject change: ${res.statusText}`);
  }
}
