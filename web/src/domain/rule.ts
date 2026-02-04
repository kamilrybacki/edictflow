export type TargetLayer = 'enterprise' | 'global' | 'project' | 'local';
export type TriggerType = 'path' | 'context' | 'tag';
export type RuleStatus = 'draft' | 'pending' | 'approved' | 'rejected';
export type EnforcementMode = 'block' | 'temporary' | 'warning';

export interface Trigger {
  type: TriggerType;
  pattern?: string;
  contextTypes?: string[];
  tags?: string[];
}

export interface Rule {
  id: string;
  name: string;
  content: string;
  targetLayer: TargetLayer;
  priorityWeight: number;
  triggers: Trigger[];
  teamId: string;
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  createdBy?: string;
  submittedAt?: string;
  approvedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export function getSpecificity(trigger: Trigger): number {
  switch (trigger.type) {
    case 'path':
      return 100;
    case 'context':
      return 50;
    case 'tag':
      return 10;
    default:
      return 0;
  }
}

export function getTargetLayerPath(layer: TargetLayer): string {
  switch (layer) {
    case 'enterprise':
      return '/etc/claude-code/CLAUDE.md';
    case 'global':
      return '~/.claude/CLAUDE.md';
    case 'project':
      return './CLAUDE.md';
    case 'local':
      return './CLAUDE.local.md';
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
