'use client';

import { ChevronRight, MessageSquare, Mail, ToggleLeft, ToggleRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { layerConfig } from '@/lib/layerConfig';
import { TargetLayer } from '@/domain/rule';

interface TeamMember {
  id: string;
  name: string;
  avatar?: string;
}

interface TeamCardData {
  id: string;
  name: string;
  members: TeamMember[];
  rulesCount: Record<TargetLayer, number>;
  inheritGlobalRules: boolean;
  notifications?: {
    slack?: boolean;
    email?: boolean;
  };
}

interface TeamCardProps {
  team: TeamCardData;
  isSelected?: boolean;
  onClick?: () => void;
}

export function TeamCard({ team, isSelected, onClick }: TeamCardProps) {
  const layers: TargetLayer[] = ['enterprise', 'user', 'project'];

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full text-left p-3 rounded-lg transition-all duration-200 group',
        isSelected
          ? 'bg-primary/10 border border-primary/30'
          : 'hover:bg-muted/50 border border-transparent'
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center text-primary-foreground font-semibold text-sm">
            {team.name.slice(0, 2).toUpperCase()}
          </div>
          <div>
            <h4 className="font-medium text-sm text-foreground">{team.name}</h4>
            <p className="text-caption">{team.members.length} members</p>
          </div>
        </div>
        <ChevronRight className={cn(
          'w-4 h-4 text-muted-foreground transition-transform',
          isSelected ? 'rotate-90' : 'group-hover:translate-x-0.5'
        )} />
      </div>

      {/* Rules count by layer */}
      <div className="flex items-center gap-3 mt-3">
        {layers.map((layer) => (
          <div key={layer} className="flex items-center gap-1.5">
            <div className={cn('w-2 h-2 rounded-full', layerConfig[layer].className)} />
            <span className="text-xs text-muted-foreground">{team.rulesCount[layer] || 0}</span>
          </div>
        ))}
      </div>

      {/* Notifications */}
      {team.notifications && (
        <div className="mt-3 pt-3 border-t flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            {team.notifications.slack && (
              <div className="p-1 rounded bg-muted" title="Slack notifications enabled">
                <MessageSquare className="w-3 h-3 text-muted-foreground" />
              </div>
            )}
            {team.notifications.email && (
              <div className="p-1 rounded bg-muted" title="Email notifications enabled">
                <Mail className="w-3 h-3 text-muted-foreground" />
              </div>
            )}
          </div>
        </div>
      )}

      {/* Global inheritance indicator */}
      <div className={cn(
        'mt-2 flex items-center gap-1.5 text-xs',
        team.inheritGlobalRules ? 'text-status-approved' : 'text-muted-foreground'
      )}>
        {team.inheritGlobalRules ? (
          <ToggleRight className="w-4 h-4" />
        ) : (
          <ToggleLeft className="w-4 h-4" />
        )}
        <span>Global rules {team.inheritGlobalRules ? 'inherited' : 'disabled'}</span>
      </div>
    </button>
  );
}
