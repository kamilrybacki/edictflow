'use client';

import { useState, useEffect } from 'react';
import { Rule, TargetLayer, isGlobalRule, getEnforcementLabel } from '@/domain/rule';
import { fetchRules, fetchGlobalRules } from '@/lib/api';

interface RuleHierarchyProps {
  teamId?: string;
  onRuleClick?: (rule: Rule) => void;
}

interface LayerData {
  layer: TargetLayer;
  label: string;
  description: string;
  rules: Rule[];
  color: string;
  bgColor: string;
  width: string;
}

export function RuleHierarchy({ teamId, onRuleClick }: RuleHierarchyProps) {
  const [globalRules, setGlobalRules] = useState<Rule[]>([]);
  const [teamRules, setTeamRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedLayer, setExpandedLayer] = useState<TargetLayer | null>(null);
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null);

  useEffect(() => {
    const loadRules = async () => {
      setLoading(true);
      try {
        const [global, team] = await Promise.all([
          fetchGlobalRules().catch(() => []),
          teamId ? fetchRules(teamId).catch(() => []) : Promise.resolve([]),
        ]);
        setGlobalRules(global);
        setTeamRules(team);
      } catch (err) {
        console.error('Failed to load rules:', err);
      } finally {
        setLoading(false);
      }
    };

    loadRules();
  }, [teamId]);

  // Organize rules by layer
  const layers: LayerData[] = [
    {
      layer: 'enterprise',
      label: 'Enterprise',
      description: 'Organization-wide policies that apply to all teams',
      rules: [
        ...globalRules.filter(r => r.status === 'approved'),
        ...teamRules.filter(r => r.targetLayer === 'enterprise' && r.status === 'approved'),
      ],
      color: 'text-purple-700 dark:text-purple-300',
      bgColor: 'bg-purple-100 dark:bg-purple-900/30 hover:bg-purple-200 dark:hover:bg-purple-900/50',
      width: 'w-48',
    },
    {
      layer: 'user',
      label: 'User',
      description: 'Personal standards pushed by admins or created by users',
      rules: teamRules.filter(r => r.targetLayer === 'user' && r.status === 'approved'),
      color: 'text-blue-700 dark:text-blue-300',
      bgColor: 'bg-blue-100 dark:bg-blue-900/30 hover:bg-blue-200 dark:hover:bg-blue-900/50',
      width: 'w-64',
    },
    {
      layer: 'project',
      label: 'Project',
      description: 'Team-specific rules for the current project',
      rules: teamRules.filter(r => r.targetLayer === 'project' && r.status === 'approved'),
      color: 'text-green-700 dark:text-green-300',
      bgColor: 'bg-green-100 dark:bg-green-900/30 hover:bg-green-200 dark:hover:bg-green-900/50',
      width: 'w-80',
    },
  ];

  const handleLayerClick = (layer: TargetLayer) => {
    setExpandedLayer(expandedLayer === layer ? null : layer);
    setSelectedRule(null);
  };

  const handleRuleClick = (rule: Rule) => {
    setSelectedRule(selectedRule?.id === rule.id ? null : rule);
    onRuleClick?.(rule);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
      </div>
    );
  }

  const totalRules = layers.reduce((sum, l) => sum + l.rules.length, 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center">
        <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">
          Rule Hierarchy
        </h3>
        <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-1">
          {totalRules} active rule{totalRules !== 1 ? 's' : ''} across {layers.filter(l => l.rules.length > 0).length} layers
        </p>
      </div>

      {/* Pyramid Visualization */}
      <div className="flex flex-col items-center space-y-2">
        {layers.map((layerData, index) => (
          <div key={layerData.layer} className="flex flex-col items-center w-full">
            {/* Layer Block */}
            <button
              onClick={() => handleLayerClick(layerData.layer)}
              className={`
                ${layerData.width} ${layerData.bgColor}
                rounded-lg py-3 px-4 transition-all duration-200
                border-2 ${expandedLayer === layerData.layer ? 'border-zinc-400 dark:border-zinc-500' : 'border-transparent'}
                shadow-sm hover:shadow-md
              `}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <LayerIcon layer={layerData.layer} />
                  <span className={`font-semibold ${layerData.color}`}>
                    {layerData.label}
                  </span>
                </div>
                <span className={`text-sm font-medium ${layerData.color}`}>
                  {layerData.rules.length}
                </span>
              </div>
              <p className="text-xs text-zinc-600 dark:text-zinc-400 mt-1 text-left">
                {layerData.description}
              </p>
            </button>

            {/* Expanded Rules List */}
            {expandedLayer === layerData.layer && (
              <div className={`${layerData.width} mt-2 space-y-1 animate-in fade-in slide-in-from-top-2 duration-200`}>
                {layerData.rules.length === 0 ? (
                  <p className="text-sm text-zinc-500 text-center py-2">
                    No rules at this level
                  </p>
                ) : (
                  layerData.rules.map((rule) => (
                    <RuleCard
                      key={rule.id}
                      rule={rule}
                      isSelected={selectedRule?.id === rule.id}
                      onClick={() => handleRuleClick(rule)}
                    />
                  ))
                )}
              </div>
            )}

            {/* Connector Line */}
            {index < layers.length - 1 && (
              <div className="h-4 w-0.5 bg-zinc-300 dark:bg-zinc-600" />
            )}
          </div>
        ))}
      </div>

      {/* Selected Rule Details */}
      {selectedRule && (
        <RuleDetails rule={selectedRule} onClose={() => setSelectedRule(null)} />
      )}

      {/* Legend */}
      <div className="border-t border-zinc-200 dark:border-zinc-700 pt-4">
        <div className="flex flex-wrap justify-center gap-4 text-xs">
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-purple-200 dark:bg-purple-800" />
            <span className="text-zinc-600 dark:text-zinc-400">Enterprise (Highest Priority)</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-blue-200 dark:bg-blue-800" />
            <span className="text-zinc-600 dark:text-zinc-400">User</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-green-200 dark:bg-green-800" />
            <span className="text-zinc-600 dark:text-zinc-400">Project (Lowest Priority)</span>
          </div>
        </div>
      </div>
    </div>
  );
}

function LayerIcon({ layer }: { layer: TargetLayer }) {
  switch (layer) {
    case 'enterprise':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
      );
    case 'user':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
      );
    case 'project':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      );
  }
}

function RuleCard({ rule, isSelected, onClick }: { rule: Rule; isSelected: boolean; onClick: () => void }) {
  const isGlobal = isGlobalRule(rule);

  return (
    <button
      onClick={onClick}
      className={`
        w-full text-left px-3 py-2 rounded-md transition-all
        ${isSelected
          ? 'bg-white dark:bg-zinc-700 shadow-md ring-2 ring-blue-500'
          : 'bg-white/50 dark:bg-zinc-800/50 hover:bg-white dark:hover:bg-zinc-700'
        }
      `}
    >
      <div className="flex items-center justify-between">
        <span className="font-medium text-sm text-zinc-800 dark:text-zinc-200 truncate">
          {rule.name}
        </span>
        {isGlobal && (
          <span className={`
            text-xs px-1.5 py-0.5 rounded
            ${rule.force
              ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
              : 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300'
            }
          `}>
            {getEnforcementLabel(rule)}
          </span>
        )}
      </div>
      {!rule.overridable && (
        <span className="text-xs text-amber-600 dark:text-amber-400">
          Non-overridable
        </span>
      )}
    </button>
  );
}

function RuleDetails({ rule, onClose }: { rule: Rule; onClose: () => void }) {
  return (
    <div className="bg-white dark:bg-zinc-800 rounded-lg shadow-lg p-4 border border-zinc-200 dark:border-zinc-700 animate-in fade-in slide-in-from-bottom-2 duration-200">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h4 className="font-semibold text-zinc-900 dark:text-white">{rule.name}</h4>
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            {rule.targetLayer.charAt(0).toUpperCase() + rule.targetLayer.slice(1)} Level
          </p>
        </div>
        <button
          onClick={onClose}
          className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {rule.description && (
        <p className="text-sm text-zinc-600 dark:text-zinc-300 mb-3">{rule.description}</p>
      )}

      <div className="bg-zinc-50 dark:bg-zinc-900 rounded-md p-3 max-h-48 overflow-y-auto">
        <pre className="text-xs text-zinc-700 dark:text-zinc-300 whitespace-pre-wrap font-mono">
          {rule.content}
        </pre>
      </div>

      <div className="flex flex-wrap gap-2 mt-3">
        {isGlobalRule(rule) && (
          <span className={`
            text-xs px-2 py-1 rounded-full
            ${rule.force
              ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
              : 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300'
            }
          `}>
            {rule.force ? 'Forced on All Teams' : 'Inheritable'}
          </span>
        )}
        {!rule.overridable && (
          <span className="text-xs px-2 py-1 rounded-full bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
            Non-overridable
          </span>
        )}
        <span className="text-xs px-2 py-1 rounded-full bg-zinc-100 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300">
          {rule.enforcementMode}
        </span>
      </div>
    </div>
  );
}

export default RuleHierarchy;
