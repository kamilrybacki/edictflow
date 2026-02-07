import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Users } from 'lucide-react';

export interface TeamNodeData {
  name: string;
  memberCount: number;
}

function TeamNodeComponent({ data, selected }: NodeProps<TeamNodeData>) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg border-2 min-w-[140px]
        bg-blue-500 text-white border-blue-600
        ${selected ? 'ring-2 ring-blue-300 ring-offset-2' : ''}
        transition-all duration-200
      `}
    >
      <Handle type="target" position={Position.Top} className="!bg-blue-700" />
      <div className="flex items-center gap-2">
        <Users className="w-4 h-4" />
        <span className="font-medium text-sm">{data.name}</span>
      </div>
      <div className="text-xs text-blue-100 mt-1">
        {data.memberCount} member{data.memberCount !== 1 ? 's' : ''}
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-blue-700" />
    </div>
  );
}

export const TeamNode = memo(TeamNodeComponent);
