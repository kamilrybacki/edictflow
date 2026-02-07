import { getApiUrlCached, getAuthHeaders } from './http';

export interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
}

export async function fetchPermissions(): Promise<Permission[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/permissions/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch permissions: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}
