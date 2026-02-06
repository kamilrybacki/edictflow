import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';

export interface UserNodeData {
  name: string;
  email: string;
  initials: string;
}

function UserNodeComponent({ data, selected }: NodeProps<UserNodeData>) {
  return (
    <div
      className={`
        flex flex-col items-center
        ${selected ? 'scale-110' : ''}
        transition-all duration-200
      `}
    >
      <Handle type="target" position={Position.Top} className="!bg-green-700" />
      <div
        className={`
          w-10 h-10 rounded-full flex items-center justify-center
          bg-green-500 text-white font-medium text-sm
          border-2 border-green-600
          ${selected ? 'ring-2 ring-green-300 ring-offset-2' : ''}
        `}
      >
        {data.initials}
      </div>
      <span className="text-xs mt-1 text-foreground max-w-[80px] truncate">
        {data.name}
      </span>
      <Handle type="source" position={Position.Bottom} className="!bg-green-700" />
    </div>
  );
}

export const UserNode = memo(UserNodeComponent);
