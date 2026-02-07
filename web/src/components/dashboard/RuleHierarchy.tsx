'use client';

import { useState } from 'react';
import { ChevronRight, Zap } from 'lucide-react';
import { cn } from '@/lib/utils';
import { layerConfig } from '@/lib/layerConfig';
import { Rule, TargetLayer } from '@/domain/rule';

interface RuleHierarchyProps {
  rules: Rule[];
  selectedRule?: Rule;
  onSelectRule?: (rule: Rule) => void;
  onSelectLayer?: (layer: TargetLayer) => void;
}

export function RuleHierarchy({ rules, selectedRule, onSelectRule, onSelectLayer }: RuleHierarchyProps) {
  const [hoveredLayer, setHoveredLayer] = useState<TargetLayer | null>(null);
  const [expandedLayer, setExpandedLayer] = useState<TargetLayer | null>(null);

  const layers: TargetLayer[] = ['organization', 'team', 'project'];

  const getRulesByLayer = (layer: TargetLayer) => rules.filter(r => r.targetLayer === layer);

  return (
    <div className="relative">
      {/* Hierarchy Stack */}
      <div className="space-y-3">
        {layers.map((layer, index) => {
          const config = layerConfig[layer];
          const LayerIcon = config.icon;
          const layerRules = getRulesByLayer(layer);
          const isExpanded = expandedLayer === layer;
          const isHovered = hoveredLayer === layer;
          const forcedCount = layerRules.filter(r => r.force).length;

          return (
            <div
              key={layer}
              className="relative animate-fade-in"
              style={{ animationDelay: `${index * 100}ms` }}
              onMouseEnter={() => setHoveredLayer(layer)}
              onMouseLeave={() => setHoveredLayer(null)}
            >
              {/* Layer Card */}
              <button
                onClick={() => {
                  setExpandedLayer(isExpanded ? null : layer);
                  onSelectLayer?.(layer);
                }}
                className={cn(
                  'w-full text-left transition-all duration-300 relative group',
                  (isHovered || isExpanded) && config.glowClassName
                )}
              >
                {/* Custom Tooltip */}
                <div className={cn(
                  'absolute -top-12 left-1/2 -translate-x-1/2 z-50 px-3 py-2 rounded-lg bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 text-xs font-medium whitespace-nowrap shadow-lg transition-all duration-200 pointer-events-none',
                  isHovered ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-1'
                )}>
                  {config.description}
                  {/* Arrow */}
                  <div className="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-zinc-900 dark:border-t-zinc-100" />
                </div>
                <div className={cn(
                  'relative overflow-hidden rounded-xl p-4 shadow-medium transition-transform duration-200',
                  config.className,
                  isHovered && 'scale-[1.02]'
                )}>
                  {/* Background pattern */}
                  <div className="absolute inset-0 opacity-10">
                    <div
                      className="absolute inset-0"
                      style={{
                        backgroundImage: 'radial-gradient(circle at 20% 50%, white 1px, transparent 1px)',
                        backgroundSize: '20px 20px'
                      }}
                    />
                  </div>

                  <div className="relative flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center backdrop-blur-sm">
                        <LayerIcon className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-lg">{config.label}</h3>
                        <p className="text-sm opacity-80">
                          {layerRules.length} rule{layerRules.length !== 1 ? 's' : ''}
                          {forcedCount > 0 && (
                            <span className="ml-2 inline-flex items-center gap-1">
                              <Zap className="w-3 h-3" />
                              {forcedCount} forced
                            </span>
                          )}
                        </p>
                      </div>
                    </div>
                    <ChevronRight className={cn(
                      'w-5 h-5 transition-transform duration-200',
                      isExpanded && 'rotate-90'
                    )} />
                  </div>

                  {/* Priority indicator */}
                  <div className="absolute top-2 right-2 text-xs opacity-60 font-mono">
                    P{index + 1}
                  </div>
                </div>
              </button>

              {/* Expanded Rules */}
              {isExpanded && (
                <div
                  className="mt-2 ml-6 pl-4 border-l-2 space-y-2 animate-fade-in"
                  style={{ borderColor: `hsl(var(--layer-${layer}))` }}
                >
                  {layerRules.map((rule) => (
                    <button
                      key={rule.id}
                      onClick={() => onSelectRule?.(rule)}
                      className={cn(
                        'w-full text-left p-3 rounded-lg transition-all duration-200',
                        selectedRule?.id === rule.id
                          ? `${config.bgClassName} ring-1`
                          : 'bg-card hover:bg-muted/50'
                      )}
                      style={selectedRule?.id === rule.id ? {
                        '--tw-ring-color': `hsl(var(--layer-${layer}))`
                      } as React.CSSProperties : {}}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-sm">{rule.name}</span>
                        {rule.force && (
                          <Zap className="w-3.5 h-3.5" style={{ color: `hsl(var(--layer-${layer}))` }} />
                        )}
                      </div>
                      <p className="text-caption mt-0.5 truncate">{rule.description || 'No description'}</p>
                    </button>
                  ))}
                </div>
              )}

              {/* Override indicator line */}
              {index < layers.length - 1 && (
                <div className="flex justify-center py-1">
                  <div className="w-px h-3 bg-gradient-to-b from-current to-transparent opacity-30" />
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Legend */}
      <div className="mt-6 pt-4 border-t">
        <p className="text-caption mb-2">Override Priority</p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full layer-organization" />
            Highest
          </span>
          <span>&rarr;</span>
          <span className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full layer-project" />
            Lowest
          </span>
        </div>
      </div>
    </div>
  );
}
