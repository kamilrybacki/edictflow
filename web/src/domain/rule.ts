export type TargetLayer = 'organization' | 'team' | 'project';
// Deprecated layer types for backwards compatibility
export type LegacyTargetLayer = 'enterprise' | 'user' | 'global' | 'local';
export type AllTargetLayers = TargetLayer | LegacyTargetLayer;

export type TriggerType = 'path' | 'context' | 'tag';
export type RuleStatus = 'draft' | 'pending' | 'approved' | 'rejected';
export type EnforcementMode = 'block' | 'temporary' | 'warning';

export interface Trigger {
  type: TriggerType;
  pattern?: string;
  contextTypes?: string[];
  tags?: string[];
}

export interface Category {
  id: string;
  name: string;
  isSystem: boolean;
  orgId?: string;
  displayOrder: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface Rule {
  id: string;
  name: string;
  content: string;
  description?: string;
  targetLayer: TargetLayer;
  categoryId?: string;
  priorityWeight: number;
  overridable: boolean;
  effectiveStart?: string;
  effectiveEnd?: string;
  targetTeams?: string[];
  targetUsers?: string[];
  tags?: string[];
  triggers: Trigger[];
  teamId?: string;           // Changed: now optional for global rules
  force: boolean;            // NEW field
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  createdBy?: string;
  createdByName?: string;
  submittedAt?: string;
  approvedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export function getTargetLayerPath(layer: TargetLayer | AllTargetLayers): string {
  switch (layer) {
    case 'organization':
    case 'enterprise': // deprecated
      return '/etc/claude-code/CLAUDE.md';
    case 'team':
    case 'user': // deprecated
    case 'global': // deprecated
      return '~/.claude/CLAUDE.md';
    case 'project':
    case 'local': // deprecated
      return './CLAUDE.md';
  }
}

export function getStatusColor(status: RuleStatus): string {
  switch (status) {
    case 'draft':
      return 'bg-zinc-100 dark:bg-zinc-700 text-zinc-800 dark:text-zinc-300';
    case 'pending':
      return 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-400';
    case 'approved':
      return 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-400';
    case 'rejected':
      return 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-400';
    default:
      return 'bg-zinc-100 dark:bg-zinc-700 text-zinc-800 dark:text-zinc-300';
  }
}

