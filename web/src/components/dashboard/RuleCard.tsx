'use client';

import { useState, memo } from 'react';
import { Edit2, History, Eye, Lock, Unlock } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui';
import { getLayerConfig, statusConfig, enforcementConfig } from '@/lib/layerConfig';
import { Rule } from '@/domain/rule';

interface RuleCardProps {
  rule: Rule;
  isSelected?: boolean;
  isHighlighted?: boolean;
  onClick?: () => void;
  onEdit?: (rule: Rule) => void;
  onViewHistory?: (rule: Rule) => void;
  onViewDetails?: (rule: Rule) => void;
}

export const RuleCard = memo(function RuleCard({ rule, isSelected, isHighlighted, onClick, onEdit, onViewHistory, onViewDetails }: RuleCardProps) {
  const [isHovered, setIsHovered] = useState(false);

  const layer = getLayerConfig(rule.targetLayer);
  const status = statusConfig[rule.status];
  const enforcement = enforcementConfig[rule.enforcementMode];
  const EnforcementIcon = enforcement.icon;
  const LayerIcon = layer.icon;

  const isGlobal = !rule.teamId;
  const inheritanceType = isGlobal ? (rule.force ? 'forced' : 'inheritable') : 'none';

  return (
    <div
      className={cn(
        'card-interactive group p-4 transition-all duration-200 relative',
        isSelected && `ring-2 ring-layer-${rule.targetLayer} ${layer.glowClassName}`,
        isHighlighted && 'animate-highlight-pulse'
      )}
      onClick={onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          {/* Layer indicator */}
          <div className={cn('w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 shadow-medium', layer.className)}>
            <LayerIcon className="w-4 h-4" />
          </div>

          <div className="min-w-0">
            <h3 className="font-semibold text-foreground truncate">{rule.name}</h3>
            <p className="text-caption truncate">{rule.description || 'No description'}</p>
          </div>
        </div>

        {/* Status & Enforcement */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <div className={enforcement.className} title={enforcement.label}>
            <EnforcementIcon className="w-4 h-4" />
          </div>
          <Badge variant="outline" className={cn(status.className, 'text-xs border')}>
            {status.label}
          </Badge>
        </div>
      </div>

      {/* Quick Actions (on hover) - below header */}
      <div className={cn(
        'flex items-center gap-1 mt-2 pt-2 border-t border-transparent transition-all',
        isHovered ? 'opacity-100 border-border' : 'opacity-0 h-0 mt-0 pt-0 overflow-hidden'
      )}>
        <button
          className="p-1.5 rounded-md hover:bg-muted transition-colors flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          title="Edit"
          onClick={(e) => {
            e.stopPropagation();
            onEdit?.(rule);
          }}
        >
          <Edit2 className="w-3.5 h-3.5" />
          <span>Edit</span>
        </button>
        <button
          className="p-1.5 rounded-md hover:bg-muted transition-colors flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          title="View History"
          onClick={(e) => {
            e.stopPropagation();
            onViewHistory?.(rule);
          }}
        >
          <History className="w-3.5 h-3.5" />
          <span>History</span>
        </button>
        <button
          className="p-1.5 rounded-md hover:bg-muted transition-colors flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          title="View Details"
          onClick={(e) => {
            e.stopPropagation();
            onViewDetails?.(rule);
          }}
        >
          <Eye className="w-3.5 h-3.5" />
          <span>Details</span>
        </button>
      </div>

      {/* Tags */}
      <div className="flex items-center gap-2 mt-3 flex-wrap">
        {inheritanceType !== 'none' && (
          <Badge
            variant="outline"
            className={cn(
              'text-[10px] gap-1',
              inheritanceType === 'forced'
                ? 'bg-enforce-block/10 text-enforce-block border-enforce-block/30'
                : 'bg-layer-enterprise/10 text-layer-enterprise border-layer-enterprise/30'
            )}
          >
            {inheritanceType === 'forced' ? <Lock className="w-2.5 h-2.5" /> : <Unlock className="w-2.5 h-2.5" />}
            Global: {inheritanceType === 'forced' ? 'Forced' : 'Inheritable'}
          </Badge>
        )}
        {rule.tags?.map((tag) => (
          <Badge key={tag} variant="secondary" className="text-[10px]">
            {tag}
          </Badge>
        ))}
      </div>
    </div>
  );
});
