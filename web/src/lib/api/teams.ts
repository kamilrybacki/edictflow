import { Team } from '@/domain/team';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchTeam(id: string): Promise<Team> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${id}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch team: ${res.statusText}`);
  }
  return res.json();
}

export async function createTeam(name: string): Promise<Team> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    throw new Error(`Failed to create team: ${res.statusText}`);
  }
  return res.json();
}

