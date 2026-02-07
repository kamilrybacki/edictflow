import { ExceptionRequest } from '@/domain/change_request';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchExceptions(
  teamId: string,
  status?: string
): Promise<ExceptionRequest[]> {
  let url = `${getApiUrlCached()}/api/v1/exceptions/?team_id=${teamId}`;
  if (status) {
    url += `&status=${status}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch exceptions: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function approveException(id: string, expiresAt?: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/exceptions/${id}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ expires_at: expiresAt }),
  });
  if (!res.ok) {
    throw new Error(`Failed to approve exception: ${res.statusText}`);
  }
}

export async function denyException(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/exceptions/${id}/deny`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to deny exception: ${res.statusText}`);
  }
}
