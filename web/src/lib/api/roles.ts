import { getApiUrlCached, getAuthHeaders } from './http';

export interface Role {
  id: string;
  name: string;
  description: string;
  hierarchy_level: number;
  permissions: string[];
  created_at: string;
}

export async function fetchRoles(): Promise<Role[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch roles: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function createRole(data: { name: string; description: string; hierarchy_level: number }): Promise<Role> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    throw new Error(`Failed to create role: ${res.statusText}`);
  }
  return res.json();
}

export async function addRolePermission(roleId: string, permissionId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/${roleId}/permissions`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ permission_id: permissionId }),
  });
  if (!res.ok) {
    throw new Error(`Failed to add permission: ${res.statusText}`);
  }
}

export async function removeRolePermission(roleId: string, permissionId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/${roleId}/permissions/${permissionId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to remove permission: ${res.statusText}`);
  }
}
