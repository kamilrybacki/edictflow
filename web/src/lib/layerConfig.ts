import { TargetLayer, RuleStatus, EnforcementMode } from '@/domain/rule';
import { Building2, Users, FolderCode, Ban, AlertTriangle, Clock, LucideIcon } from 'lucide-react';

/**
 * Normalize legacy layer names to current naming convention.
 * Maps: enterprise -> organization, user -> team
 */
export function normalizeTargetLayer(layer: string): TargetLayer {
  const legacyMapping: Record<string, TargetLayer> = {
    enterprise: 'organization',
    user: 'team',
    global: 'organization', // legacy alias
    local: 'project', // legacy alias
  };
  return legacyMapping[layer] || (layer as TargetLayer);
}

/**
 * Safely get layer config, normalizing legacy layer names.
 */
export function getLayerConfig(layer: string) {
  const normalized = normalizeTargetLayer(layer);
  return layerConfig[normalized];
}

export const layerConfig: Record<TargetLayer, {
  label: string;
  description: string;
  icon: LucideIcon;
  className: string;
  borderClassName: string;
  glowClassName: string;
  bgClassName: string;
}> = {
  organization: {
    label: 'Organization',
    description: 'Applies to all teams and projects across the entire organization',
    icon: Building2,
    className: 'layer-organization text-white',
    borderClassName: 'border-layer-organization',
    glowClassName: 'glow-organization',
    bgClassName: 'bg-layer-organization/10',
  },
  team: {
    label: 'Team',
    description: 'Applies to all projects owned by a specific team',
    icon: Users,
    className: 'layer-team text-white',
    borderClassName: 'border-layer-team',
    glowClassName: 'glow-team',
    bgClassName: 'bg-layer-team/10',
  },
  project: {
    label: 'Project',
    description: 'Applies only to a single repository or project',
    icon: FolderCode,
    className: 'layer-project text-white',
    borderClassName: 'border-layer-project',
    glowClassName: 'glow-project',
    bgClassName: 'bg-layer-project/10',
  },
};

export const statusConfig: Record<RuleStatus, {
  label: string;
  className: string;
}> = {
  draft: { label: 'Draft', className: 'status-draft' },
  pending: { label: 'Pending', className: 'status-pending' },
  approved: { label: 'Approved', className: 'status-approved' },
  rejected: { label: 'Rejected', className: 'status-rejected' },
};

export const enforcementConfig: Record<EnforcementMode, {
  label: string;
  icon: LucideIcon;
  className: string;
}> = {
  block: {
    label: 'Block',
    icon: Ban,
    className: 'text-enforce-block',
  },
  warning: {
    label: 'Warning',
    icon: AlertTriangle,
    className: 'text-enforce-warning',
  },
  temporary: {
    label: 'Temporary',
    icon: Clock,
    className: 'text-enforce-temporary',
  },
};
