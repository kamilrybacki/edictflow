import { Team } from '@/domain/team';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchTeams(): Promise<Team[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch teams: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

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

export async function deleteTeam(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete team: ${res.statusText}`);
  }
}

export async function updateTeamSettings(
  teamId: string,
  settings: { inherit_global_rules?: boolean; drift_threshold_minutes?: number }
): Promise<Team> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${teamId}/settings`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    body: JSON.stringify(settings),
  });
  if (!res.ok) throw new Error('Failed to update team settings');
  return res.json();
}
