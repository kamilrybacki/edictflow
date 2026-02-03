export interface TeamSettings {
  driftThresholdMinutes: number;
}

export interface Team {
  id: string;
  name: string;
  settings: TeamSettings;
  createdAt: string;
}

export function createDefaultTeamSettings(): TeamSettings {
  return {
    driftThresholdMinutes: 60,
  };
}
