'use client';

import { useEffect, useState, useRef, useMemo, useCallback } from 'react';
import {
  Search,
  Plus,
  FileCheck,
  Wifi,
  LogOut,
  ArrowRight,
} from 'lucide-react';
import { Rule } from '@/domain/rule';
import { TeamData } from '@/domain/team';
import { getLayerConfig, statusConfig } from '@/lib/layerConfig';
import { cn } from '@/lib/utils';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  rules: Rule[];
  teams: TeamData[];
  onSelectRule: (rule: Rule) => void;
  onSelectTeam: (team: TeamData | null) => void;
  onCreateRule: () => void;
  onCreateTeam: () => void;
  onViewApprovals: () => void;
  onViewAgents: () => void;
  onLogout: () => void;
}

interface SearchResult {
  id: string;
  type: 'rule' | 'team' | 'action';
  icon: React.ReactNode;
  primary: string;
  secondary: string;
  shortcut?: string;
  onSelect: () => void;
  data?: Rule | TeamData;
}

const MIN_SEARCH_LENGTH = 3;
const MAX_RULES = 5;
const MAX_TEAMS = 3;

export function CommandPalette({
  isOpen,
  onClose,
  rules,
  teams,
  onSelectRule,
  onSelectTeam,
  onCreateRule,
  onCreateTeam,
  onViewApprovals,
  onViewAgents,
  onLogout,
}: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Quick actions - always available
  const quickActions = useMemo<SearchResult[]>(() => [
    {
      id: 'action-create-rule',
      type: 'action',
      icon: <Plus className="w-4 h-4" />,
      primary: 'Create Rule',
      secondary: 'Add a new rule',
      shortcut: '',
      onSelect: () => { onCreateRule(); onClose(); },
    },
    {
      id: 'action-create-team',
      type: 'action',
      icon: <Plus className="w-4 h-4" />,
      primary: 'Create Team',
      secondary: 'Add a new team',
      shortcut: '',
      onSelect: () => { onCreateTeam(); onClose(); },
    },
    {
      id: 'action-view-approvals',
      type: 'action',
      icon: <FileCheck className="w-4 h-4" />,
      primary: 'View Approvals',
      secondary: 'Review pending approvals',
      shortcut: '',
      onSelect: () => { onViewApprovals(); onClose(); },
    },
    {
      id: 'action-view-agents',
      type: 'action',
      icon: <Wifi className="w-4 h-4" />,
      primary: 'View Agents',
      secondary: 'See connected agents',
      shortcut: '',
      onSelect: () => { onViewAgents(); onClose(); },
    },
    {
      id: 'action-logout',
      type: 'action',
      icon: <LogOut className="w-4 h-4" />,
      primary: 'Logout',
      secondary: 'Sign out of your account',
      shortcut: '',
      onSelect: () => { onLogout(); onClose(); },
    },
  ], [onCreateRule, onCreateTeam, onViewApprovals, onViewAgents, onLogout, onClose]);

  // Filter and build results
  const results = useMemo<SearchResult[]>(() => {
    const lowerQuery = query.toLowerCase().trim();
    const shouldSearch = lowerQuery.length >= MIN_SEARCH_LENGTH;

    const filtered: SearchResult[] = [];

    // Filter rules if query is long enough
    if (shouldSearch) {
      const matchingRules = rules
        .filter(rule =>
          rule.name.toLowerCase().includes(lowerQuery) ||
          (rule.description?.toLowerCase().includes(lowerQuery))
        )
        .sort((a, b) => {
          // Prefix matches first
          const aPrefix = a.name.toLowerCase().startsWith(lowerQuery);
          const bPrefix = b.name.toLowerCase().startsWith(lowerQuery);
          if (aPrefix && !bPrefix) return -1;
          if (!aPrefix && bPrefix) return 1;
          return 0;
        })
        .slice(0, MAX_RULES);

      matchingRules.forEach(rule => {
        const config = getLayerConfig(rule.targetLayer);
        const Icon = config.icon;
        filtered.push({
          id: `rule-${rule.id}`,
          type: 'rule',
          icon: (
            <div className={cn('w-6 h-6 rounded flex items-center justify-center', config.className)}>
              <Icon className="w-3.5 h-3.5" />
            </div>
          ),
          primary: rule.name,
          secondary: `${statusConfig[rule.status].label} · ${config.label}`,
          onSelect: () => { onSelectRule(rule); onClose(); },
          data: rule,
        });
      });

      // Filter teams
      const matchingTeams = teams
        .filter(team => team.name.toLowerCase().includes(lowerQuery))
        .sort((a, b) => {
          const aPrefix = a.name.toLowerCase().startsWith(lowerQuery);
          const bPrefix = b.name.toLowerCase().startsWith(lowerQuery);
          if (aPrefix && !bPrefix) return -1;
          if (!aPrefix && bPrefix) return 1;
          return 0;
        })
        .slice(0, MAX_TEAMS);

      matchingTeams.forEach(team => {
        filtered.push({
          id: `team-${team.id}`,
          type: 'team',
          icon: (
            <div className="w-6 h-6 rounded bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center text-primary-foreground text-xs font-semibold">
              {team.name.slice(0, 2).toUpperCase()}
            </div>
          ),
          primary: team.name,
          secondary: `${team.members.length} member${team.members.length !== 1 ? 's' : ''}`,
          onSelect: () => { onSelectTeam(team); onClose(); },
          data: team,
        });
      });
    }

    // Filter actions
    const matchingActions = shouldSearch
      ? quickActions.filter(action =>
          action.primary.toLowerCase().includes(lowerQuery)
        )
      : quickActions;

    filtered.push(...matchingActions);

    return filtered;
  }, [query, rules, teams, quickActions, onSelectRule, onSelectTeam, onClose]);

  // Group results by type
  const groupedResults = useMemo(() => {
    const groups: { type: 'rule' | 'team' | 'action'; label: string; items: SearchResult[] }[] = [];

    const ruleItems = results.filter(r => r.type === 'rule');
    const teamItems = results.filter(r => r.type === 'team');
    const actionItems = results.filter(r => r.type === 'action');

    if (ruleItems.length > 0) {
      groups.push({ type: 'rule', label: 'Rules', items: ruleItems });
    }
    if (teamItems.length > 0) {
      groups.push({ type: 'team', label: 'Teams', items: teamItems });
    }
    if (actionItems.length > 0) {
      groups.push({ type: 'action', label: 'Quick Actions', items: actionItems });
    }

    return groups;
  }, [results]);

  // Reset selection when results change
  useEffect(() => {
    setSelectedIndex(0);
  }, [results]);

  // Focus input when opened
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [isOpen]);

  // Keyboard navigation
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(i => Math.min(i + 1, results.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(i => Math.max(i - 1, 0));
        break;
      case 'Enter':
        e.preventDefault();
        if (results[selectedIndex]) {
          results[selectedIndex].onSelect();
        }
        break;
      case 'Escape':
        e.preventDefault();
        onClose();
        break;
    }
  }, [results, selectedIndex, onClose]);

  // Scroll selected item into view
  useEffect(() => {
    const list = listRef.current;
    if (!list) return;

    const selectedEl = list.querySelector(`[data-index="${selectedIndex}"]`);
    if (selectedEl) {
      selectedEl.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  // Handle backdrop click
  const handleBackdropClick = useCallback((e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  }, [onClose]);

  if (!isOpen) return null;

  let flatIndex = 0;

  return (
    <div
      className="fixed inset-0 z-50 bg-black/50 flex items-start justify-center pt-[15vh]"
      onClick={handleBackdropClick}
    >
      <div className="w-full max-w-lg bg-card rounded-xl shadow-2xl border overflow-hidden">
        {/* Search Input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b">
          <Search className="w-5 h-5 text-muted-foreground flex-shrink-0" />
          <input
            ref={inputRef}
            type="text"
            placeholder="Search rules, teams, or actions..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            className="flex-1 bg-transparent outline-none text-foreground placeholder:text-muted-foreground"
          />
          <kbd className="hidden sm:inline-flex items-center gap-1 px-2 py-0.5 text-xs text-muted-foreground bg-muted rounded">
            esc
          </kbd>
        </div>

        {/* Results */}
        <div ref={listRef} className="max-h-80 overflow-y-auto p-2">
          {results.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              No results found
            </div>
          ) : (
            groupedResults.map((group) => (
              <div key={group.type} className="mb-2 last:mb-0">
                <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                  {group.label}
                </div>
                {group.items.map((result) => {
                  const currentIndex = flatIndex++;
                  const isSelected = currentIndex === selectedIndex;

                  return (
                    <button
                      key={result.id}
                      data-index={currentIndex}
                      onClick={result.onSelect}
                      onMouseEnter={() => setSelectedIndex(currentIndex)}
                      className={cn(
                        'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors',
                        isSelected ? 'bg-accent' : 'hover:bg-accent/50'
                      )}
                    >
                      <div className="flex-shrink-0">
                        {result.icon}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{result.primary}</div>
                        <div className="text-sm text-muted-foreground truncate">{result.secondary}</div>
                      </div>
                      {isSelected && (
                        <ArrowRight className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                      )}
                    </button>
                  );
                })}
              </div>
            ))
          )}
        </div>

        {/* Footer hint */}
        <div className="px-4 py-2 border-t bg-muted/30 text-xs text-muted-foreground flex items-center gap-4">
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">↑</kbd>
            <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">↓</kbd>
            to navigate
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">↵</kbd>
            to select
          </span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">esc</kbd>
            to close
          </span>
        </div>
      </div>
    </div>
  );
}
