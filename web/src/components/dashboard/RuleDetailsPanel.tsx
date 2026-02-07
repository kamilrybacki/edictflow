'use client';

import { X, Edit2, History, Copy, Lock, Unlock, Calendar, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge, Button } from '@/components/ui';
import { getLayerConfig, statusConfig, enforcementConfig } from '@/lib/layerConfig';
import { Rule } from '@/domain/rule';

interface RuleDetailsPanelProps {
  rule: Rule;
  onClose: () => void;
  onEdit?: (rule: Rule) => void;
  onViewHistory?: (rule: Rule) => void;
}

export function RuleDetailsPanel({ rule, onClose, onEdit, onViewHistory }: RuleDetailsPanelProps) {
  const layer = getLayerConfig(rule.targetLayer);
  const status = statusConfig[rule.status];
  const enforcement = enforcementConfig[rule.enforcementMode];
  const EnforcementIcon = enforcement.icon;
  const LayerIcon = layer.icon;

  const isGlobal = !rule.teamId;
  const inheritanceType = isGlobal ? (rule.force ? 'forced' : 'inheritable') : 'team-specific';

  const handleCopyContent = () => {
    navigator.clipboard.writeText(rule.content);
  };

  return (
    <div className="bg-card rounded-xl border h-full flex flex-col">
      {/* Header */}
      <div className={cn('p-4 rounded-t-xl', layer.bgClassName)}>
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <div className={cn('w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 shadow-medium', layer.className)}>
              <LayerIcon className="w-5 h-5" />
            </div>
            <div className="min-w-0">
              <h2 className="font-semibold text-lg truncate">{rule.name}</h2>
              <p className="text-sm text-muted-foreground">{layer.label} Rule</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md hover:bg-muted transition-colors"
          >
            <X className="w-5 h-5 text-muted-foreground" />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Status Row */}
        <div className="flex items-center gap-3 flex-wrap">
          <Badge variant="outline" className={cn(status.className, 'text-xs border')}>
            {status.label}
          </Badge>
          <div className={cn('flex items-center gap-1.5 text-sm', enforcement.className)}>
            <EnforcementIcon className="w-4 h-4" />
            <span>{enforcement.label}</span>
          </div>
          <Badge
            variant="outline"
            className={cn(
              'text-xs gap-1',
              inheritanceType === 'forced'
                ? 'bg-enforce-block/10 text-enforce-block border-enforce-block/30'
                : inheritanceType === 'inheritable'
                  ? 'bg-layer-enterprise/10 text-layer-enterprise border-layer-enterprise/30'
                  : 'bg-muted text-muted-foreground'
            )}
          >
            {inheritanceType === 'forced' ? <Lock className="w-2.5 h-2.5" /> : <Unlock className="w-2.5 h-2.5" />}
            {inheritanceType === 'forced' ? 'Forced' : inheritanceType === 'inheritable' ? 'Inheritable' : 'Team Only'}
          </Badge>
        </div>

        {/* Description */}
        {rule.description && (
          <div>
            <h3 className="text-sm font-medium mb-1">Description</h3>
            <p className="text-sm text-muted-foreground">{rule.description}</p>
          </div>
        )}

        {/* Tags */}
        {rule.tags && rule.tags.length > 0 && (
          <div>
            <h3 className="text-sm font-medium mb-2">Tags</h3>
            <div className="flex flex-wrap gap-1.5">
              {rule.tags.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {/* Metadata */}
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Calendar className="w-4 h-4" />
            <span>Created: {new Date(rule.createdAt).toLocaleDateString()}</span>
          </div>
          <div className="flex items-center gap-2 text-muted-foreground">
            <User className="w-4 h-4" />
            <span>By: {rule.createdByName || rule.createdBy || 'Unknown'}</span>
          </div>
        </div>

        {/* Content */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-medium">Rule Content</h3>
            <button
              onClick={handleCopyContent}
              className="p-1.5 rounded-md hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
              title="Copy content"
            >
              <Copy className="w-4 h-4" />
            </button>
          </div>
          <div className="p-3 bg-muted/50 rounded-lg font-mono text-xs text-muted-foreground whitespace-pre-wrap max-h-64 overflow-y-auto">
            {rule.content}
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="p-4 border-t flex items-center gap-2">
        <Button
          size="sm"
          className="gap-1.5"
          onClick={() => onEdit?.(rule)}
        >
          <Edit2 className="w-4 h-4" />
          Edit Rule
        </Button>
        <Button
          size="sm"
          variant="outline"
          className="gap-1.5"
          onClick={() => onViewHistory?.(rule)}
        >
          <History className="w-4 h-4" />
          View History
        </Button>
      </div>
    </div>
  );
}
