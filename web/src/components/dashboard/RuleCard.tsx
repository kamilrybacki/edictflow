'use client';

import { useState } from 'react';
import { ChevronDown, Edit2, History, Eye, Lock, Unlock } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui';
import { layerConfig, statusConfig, enforcementConfig } from '@/lib/layerConfig';
import { Rule } from '@/domain/rule';

interface RuleCardProps {
  rule: Rule;
  isSelected?: boolean;
  onClick?: () => void;
  onEdit?: (rule: Rule) => void;
}

export function RuleCard({ rule, isSelected, onClick, onEdit }: RuleCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  const layer = layerConfig[rule.targetLayer];
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
        isSelected && `ring-2 ring-layer-${rule.targetLayer} ${layer.glowClassName}`
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

      {/* Expandable Content Preview */}
      <button
        className="w-full mt-3 pt-3 border-t flex items-center justify-between text-xs text-muted-foreground hover:text-foreground transition-colors"
        onClick={(e) => {
          e.stopPropagation();
          setIsExpanded(!isExpanded);
        }}
      >
        <span>Preview content</span>
        <ChevronDown className={cn('w-4 h-4 transition-transform', isExpanded && 'rotate-180')} />
      </button>

      {isExpanded && (
        <div className="mt-2 p-3 bg-muted/50 rounded-md font-mono text-xs text-muted-foreground whitespace-pre-wrap animate-fade-in">
          {rule.content.split('\n').slice(0, 3).join('\n')}
          {rule.content.split('\n').length > 3 && '\n...'}
        </div>
      )}

      {/* Quick Actions (on hover) */}
      <div className={cn(
        'absolute top-3 right-3 flex items-center gap-1 transition-opacity',
        isHovered ? 'opacity-100' : 'opacity-0'
      )}>
        <button
          className="p-1.5 rounded-md hover:bg-muted transition-colors"
          title="Edit"
          onClick={(e) => {
            e.stopPropagation();
            onEdit?.(rule);
          }}
        >
          <Edit2 className="w-3.5 h-3.5 text-muted-foreground" />
        </button>
        <button className="p-1.5 rounded-md hover:bg-muted transition-colors" title="View History">
          <History className="w-3.5 h-3.5 text-muted-foreground" />
        </button>
        <button className="p-1.5 rounded-md hover:bg-muted transition-colors" title="View Details">
          <Eye className="w-3.5 h-3.5 text-muted-foreground" />
        </button>
      </div>
    </div>
  );
}
