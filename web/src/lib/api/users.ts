import { User } from '@/domain/user';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchUsers(): Promise<User[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch users: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function deactivateUser(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/${id}/deactivate`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to deactivate user: ${res.statusText}`);
  }
}

export async function assignUserRole(userId: string, roleId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/${userId}/roles`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ role_id: roleId }),
  });
  if (!res.ok) {
    throw new Error(`Failed to assign role: ${res.statusText}`);
  }
}
