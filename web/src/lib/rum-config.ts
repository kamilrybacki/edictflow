export interface RUMConfig {
  enabled: boolean;
  appKey: string;
  adrumExtUrl: string;
  beaconUrl: string;
}

export function getRUMConfig(): RUMConfig {
  return {
    enabled: process.env.NEXT_PUBLIC_APPDYNAMICS_ENABLED === 'true',
    appKey: process.env.NEXT_PUBLIC_APPDYNAMICS_APP_KEY || '',
    adrumExtUrl: process.env.NEXT_PUBLIC_APPDYNAMICS_ADR_URL || '',
    beaconUrl: process.env.NEXT_PUBLIC_APPDYNAMICS_BEACON_URL || '',
  };
}
