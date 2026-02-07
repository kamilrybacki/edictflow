import { TargetLayer, TriggerType, EnforcementMode } from '@/domain/rule';

export const targetLayers: { value: TargetLayer; label: string }[] = [
  { value: 'organization', label: 'Organization - Applies to all teams and projects' },
  { value: 'team', label: 'Team - Applies to all team projects' },
  { value: 'project', label: 'Project - Applies to a single repository' },
];

export const triggerTypes: TriggerType[] = ['path', 'context', 'tag'];

export const enforcementModes: { value: EnforcementMode; label: string; description: string }[] = [
  { value: 'block', label: 'Block', description: 'Changes are immediately reverted and require admin approval to apply.' },
  { value: 'temporary', label: 'Temporary', description: 'Changes apply temporarily but auto-revert if not approved within the timeout.' },
  { value: 'warning', label: 'Warning', description: 'Changes apply permanently but are flagged for admin review.' },
];

// Timeout configuration constants
export const DEFAULT_TIMEOUT_HOURS = 24;
export const MIN_TIMEOUT_HOURS = 1;
export const MAX_TIMEOUT_HOURS = 168; // 1 week
