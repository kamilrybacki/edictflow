// Agent API functions

import { getAuthHeaders } from './http';

export interface ConnectedAgent {
  agent_id: string;
  user_id: string;
  team_id: string;
  hostname?: string;
  version?: string;
  os?: string;
  connected_at?: string;
  remote_addr?: string;
}

// Get the worker URL (WebSocket server, different from API server)
function getWorkerUrl(): string {
  if (typeof window !== 'undefined') {
    // Check for runtime config
    const runtimeUrl = (window as unknown as { __WORKER_URL__?: string }).__WORKER_URL__;
    if (runtimeUrl) return runtimeUrl;

    // Use environment variable if available
    if (process.env.NEXT_PUBLIC_WORKER_URL) {
      return process.env.NEXT_PUBLIC_WORKER_URL;
    }

    // Default fallback - assume worker is on port 8081
    return `http://${window.location.hostname}:8081`;
  }
  return process.env.NEXT_PUBLIC_WORKER_URL || 'http://localhost:8081';
}

// Fetch list of connected agents from the worker
export async function fetchConnectedAgents(teamId?: string): Promise<ConnectedAgent[]> {
  const workerUrl = getWorkerUrl();
  const url = teamId ? `${workerUrl}/agents?team_id=${teamId}` : `${workerUrl}/agents`;
  const res = await fetch(url, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error('Failed to fetch connected agents');
  }
  const data = await res.json();
  return data || [];
}

// Fetch worker health/stats
export async function fetchWorkerHealth(): Promise<{
  status: string;
  agents: number;
  teams: number;
  subscriptions: number;
}> {
  const workerUrl = getWorkerUrl();
  const res = await fetch(`${workerUrl}/health`);
  if (!res.ok) {
    throw new Error('Failed to fetch worker health');
  }
  return res.json();
}
