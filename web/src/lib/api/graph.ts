import { RuleStatus, EnforcementMode } from '@/domain/rule';

export interface GraphTeam {
  id: string;
  name: string;
  memberCount: number;
}

export interface GraphUser {
  id: string;
  name: string;
  email: string;
  teamId: string | null;
}

export interface GraphRule {
  id: string;
  name: string;
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  teamId: string | null;
  targetTeams: string[];
  targetUsers: string[];
}

export interface GraphData {
  teams: GraphTeam[];
  users: GraphUser[];
  rules: GraphRule[];
}

export async function fetchGraphData(): Promise<GraphData> {
  const response = await fetch('/api/graph');
  if (!response.ok) {
    throw new Error('Failed to fetch graph data');
  }
  return response.json();
}
