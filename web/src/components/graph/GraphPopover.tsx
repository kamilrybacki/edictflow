import { X, ExternalLink } from 'lucide-react';
import Link from 'next/link';
import { GraphTeam, GraphUser, GraphRule } from '@/lib/api/graph';

interface BasePopoverProps {
  position: { x: number; y: number };
  onClose: () => void;
}

interface TeamPopoverProps extends BasePopoverProps {
  type: 'team';
  data: GraphTeam;
}

interface UserPopoverProps extends BasePopoverProps {
  type: 'user';
  data: GraphUser;
}

interface RulePopoverProps extends BasePopoverProps {
  type: 'rule';
  data: GraphRule;
}

type GraphPopoverProps = TeamPopoverProps | UserPopoverProps | RulePopoverProps;

export function GraphPopover(props: GraphPopoverProps) {
  const { position, onClose, type, data } = props;

  const getDetailsLink = () => {
    switch (type) {
      case 'team':
        return `/admin/teams/${data.id}/settings`;
      case 'user':
        return `/admin/users/${data.id}`;
      case 'rule':
        return `/rules/${data.id}`;
    }
  };

  const renderContent = () => {
    switch (type) {
      case 'team':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="text-sm text-muted-foreground mt-1">
              {data.memberCount} member{data.memberCount !== 1 ? 's' : ''}
            </div>
          </>
        );
      case 'user':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="text-sm text-muted-foreground mt-1">{data.email}</div>
            {data.teamId && (
              <div className="text-xs text-muted-foreground mt-1">
                Team: {data.teamId}
              </div>
            )}
          </>
        );
      case 'rule':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="flex gap-2 mt-2">
              <span className="text-xs px-2 py-0.5 rounded bg-muted">
                {data.status}
              </span>
              <span className="text-xs px-2 py-0.5 rounded bg-muted">
                {data.enforcementMode}
              </span>
            </div>
          </>
        );
    }
  };

  return (
    <div
      className="absolute z-50 bg-card border rounded-lg shadow-lg p-3 min-w-[180px]"
      style={{
        left: position.x,
        top: position.y,
        transform: 'translate(-50%, -100%) translateY(-10px)',
      }}
    >
      <button
        onClick={onClose}
        className="absolute top-2 right-2 text-muted-foreground hover:text-foreground"
      >
        <X className="w-4 h-4" />
      </button>

      <div className="pr-6">{renderContent()}</div>

      <Link
        href={getDetailsLink()}
        className="flex items-center gap-1 text-xs text-primary hover:underline mt-3"
      >
        View details <ExternalLink className="w-3 h-3" />
      </Link>
    </div>
  );
}
