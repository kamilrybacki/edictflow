'use client';

import { memo, useMemo } from 'react';
import Image from 'next/image';
import { ChevronRight, MessageSquare, Mail, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { layerConfig } from '@/lib/layerConfig';
import { TargetLayer } from '@/domain/rule';
import { TeamData } from '@/domain/team';

interface TeamCardProps {
  team: TeamData;
  isSelected?: boolean;
  onClick?: () => void;
}

const layers: TargetLayer[] = ['organization', 'team', 'project'];

export const TeamCard = memo(function TeamCard({ team, isSelected, onClick }: TeamCardProps) {
  const hasNotifications = team.notifications?.slack || team.notifications?.email;
  const totalRules = useMemo(
    () => layers.reduce((sum, layer) => sum + (team.rulesCount[layer] || 0), 0),
    [team.rulesCount]
  );

  return (
    <div
      onClick={onClick}
      className={cn(
        'w-full text-left p-3 rounded-lg transition-all duration-200 group cursor-pointer',
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
      <div className="flex items-center gap-3 mt-2">
        {layers.map((layer) => (
          <div key={layer} className="flex items-center gap-1.5" title={layerConfig[layer].label}>
            <div className={cn('w-2 h-2 rounded-full', layerConfig[layer].className)} />
            <span className="text-xs text-muted-foreground">{team.rulesCount[layer] || 0}</span>
          </div>
        ))}
      </div>

      {/* Expanded content when selected */}
      {isSelected && (
        <div className="mt-3 pt-3 border-t border-border/50 space-y-3 animate-fade-in">
          {/* Team members */}
          <div>
            <p className="text-xs text-muted-foreground mb-2">Members</p>
            <div className="space-y-1.5">
              {team.members.length > 0 ? (
                team.members.slice(0, 4).map((member) => (
                  <div key={member.id} className="flex items-center gap-2">
                    {member.avatar ? (
                      <Image
                        src={member.avatar}
                        alt={member.name}
                        width={20}
                        height={20}
                        className="rounded-full"
                      />
                    ) : (
                      <div className="w-5 h-5 rounded-full bg-muted flex items-center justify-center">
                        <User className="w-3 h-3 text-muted-foreground" />
                      </div>
                    )}
                    <span className="text-xs text-foreground truncate">{member.name}</span>
                  </div>
                ))
              ) : (
                <p className="text-xs text-muted-foreground italic">No members yet</p>
              )}
              {team.members.length > 4 && (
                <p className="text-xs text-muted-foreground">
                  +{team.members.length - 4} more
                </p>
              )}
            </div>
          </div>

          {/* Quick stats */}
          <div className="flex items-center justify-between text-xs">
            <span className="text-muted-foreground">{totalRules} total rules</span>
            {hasNotifications && (
              <div className="flex items-center gap-1">
                {team.notifications?.slack && (
                  <span title="Slack enabled">
                    <MessageSquare className="w-3 h-3 text-muted-foreground" />
                  </span>
                )}
                {team.notifications?.email && (
                  <span title="Email enabled">
                    <Mail className="w-3 h-3 text-muted-foreground" />
                  </span>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
});
