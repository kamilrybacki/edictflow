export type TargetLayer = 'enterprise' | 'global' | 'project' | 'local';
export type TriggerType = 'path' | 'context' | 'tag';

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
