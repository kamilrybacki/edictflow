import { TargetLayer } from './rule';

export interface TeamSettings {
  driftThresholdMinutes: number;
  inheritGlobalRules: boolean;
}

export interface Team {
  id: string;
  name: string;
  settings: TeamSettings;
  createdAt: string;
}

// Shared UI types for team-related components
export interface TeamMember {
  id: string;
  name: string;
  avatar?: string;
}

export interface TeamData {
  id: string;
  name: string;
  members: TeamMember[];
  rulesCount: Record<TargetLayer, number>;
  inheritGlobalRules: boolean;
  notifications?: {
    slack?: boolean;
    email?: boolean;
  };
}
