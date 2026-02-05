import { TargetLayer, RuleStatus, EnforcementMode } from '@/domain/rule';
import { Building2, User, FolderCode, Ban, AlertTriangle, Clock, LucideIcon } from 'lucide-react';

export const layerConfig: Record<TargetLayer, {
  label: string;
  icon: LucideIcon;
  className: string;
  borderClassName: string;
  glowClassName: string;
  bgClassName: string;
}> = {
  enterprise: {
    label: 'Enterprise',
    icon: Building2,
    className: 'layer-enterprise text-white',
    borderClassName: 'border-layer-enterprise',
    glowClassName: 'glow-enterprise',
    bgClassName: 'bg-layer-enterprise/10',
  },
  user: {
    label: 'User',
    icon: User,
    className: 'layer-user text-white',
    borderClassName: 'border-layer-user',
    glowClassName: 'glow-user',
    bgClassName: 'bg-layer-user/10',
  },
  project: {
    label: 'Project',
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
