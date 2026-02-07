import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Shield, Check, X, Clock, FileEdit } from 'lucide-react';
import { RuleStatus } from '@/domain/rule';

export interface RuleNodeData {
  name: string;
  status: RuleStatus;
  enforcementMode: string;
}

const statusConfig: Record<RuleStatus, { border: string; badge: React.ReactNode }> = {
  draft: {
    border: 'border-dashed border-amber-500',
    badge: <FileEdit className="w-3 h-3" />,
  },
  pending: {
    border: 'border-amber-500 animate-pulse',
    badge: <Clock className="w-3 h-3" />,
  },
  approved: {
    border: 'border-solid border-amber-600',
    badge: <Check className="w-3 h-3 text-green-500" />,
  },
  rejected: {
    border: 'border-solid border-red-500',
    badge: <X className="w-3 h-3 text-red-500" />,
  },
};

function RuleNodeComponent({ data, selected }: NodeProps<RuleNodeData>) {
  const config = statusConfig[data.status] || statusConfig.draft;

  return (
    <div
      className={`
        px-3 py-2 min-w-[120px]
        bg-amber-500 text-white border-2 rounded-lg
        ${config.border}
        ${selected ? 'ring-2 ring-amber-300 ring-offset-2' : ''}
        transition-all duration-200
      `}
      style={{
        clipPath: 'polygon(10% 0%, 90% 0%, 100% 50%, 90% 100%, 10% 100%, 0% 50%)',
        padding: '12px 20px',
      }}
    >
      <Handle type="target" position={Position.Top} className="!bg-amber-700" />
      <div className="flex items-center gap-2">
        <Shield className="w-4 h-4" />
        <span className="font-medium text-xs">{data.name}</span>
        <span className="ml-auto">{config.badge}</span>
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-amber-700" />
    </div>
  );
}

export const RuleNode = memo(RuleNodeComponent);
