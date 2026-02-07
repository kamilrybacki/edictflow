'use client';

import { ReactNode, useState, useEffect } from 'react';
import {
  Search,
  Settings,
  Command,
  Plus,
  ChevronLeft,
  ChevronRight,
  LogOut,
  Users,
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui';
import { TeamCard } from './TeamCard';
import { NotificationBell } from '@/components/NotificationBell';
import { Rule } from '@/domain/rule';
import { TeamData } from '@/domain/team';
import { CommandPalette } from '@/components/CommandPalette';

interface UserData {
  name: string;
  role: string;
  initials: string;
}

interface DashboardLayoutProps {
  children: ReactNode;
  teams: TeamData[];
  selectedTeam?: TeamData | null;
  onSelectTeam: (team: TeamData | null) => void;
  currentUser?: UserData;
  onCreateRule?: () => void;
  onCreateTeam?: () => void;
  onViewRuleHistory?: (ruleId: string) => void;
  // Command palette props
  rules?: Rule[];
  onSelectRule?: (rule: Rule) => void;
  onViewApprovals?: () => void;
  onViewAgents?: () => void;
}

export function DashboardLayout({
  children,
  teams,
  selectedTeam,
  onSelectTeam,
  currentUser,
  onCreateRule,
  onCreateTeam,
  onViewRuleHistory,
  rules = [],
  onSelectRule,
  onViewApprovals,
  onViewAgents,
}: DashboardLayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { logout } = useAuth();

  // Global keyboard shortcut for command palette
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setCommandPaletteOpen(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleLogout = () => {
    logout();
    window.location.href = '/login';
  };

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside
        className={cn(
          'bg-sidebar border-r border-sidebar-border transition-all duration-300 flex flex-col',
          sidebarCollapsed ? 'w-16' : 'w-72'
        )}
      >
        {/* Logo */}
        <div className="p-4 border-b border-sidebar-border flex items-center justify-between">
          {!sidebarCollapsed && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg layer-enterprise flex items-center justify-center">
                <Command className="w-4 h-4 text-white" />
              </div>
              <span className="font-bold text-lg">Edictflow</span>
            </div>
          )}
          {sidebarCollapsed && (
            <div className="w-8 h-8 rounded-lg layer-enterprise flex items-center justify-center mx-auto">
              <Command className="w-4 h-4 text-white" />
            </div>
          )}
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1 rounded hover:bg-sidebar-accent transition-colors"
          >
            {sidebarCollapsed ? (
              <ChevronRight className="w-4 h-4 text-muted-foreground" />
            ) : (
              <ChevronLeft className="w-4 h-4 text-muted-foreground" />
            )}
          </button>
        </div>

        {/* Teams Section */}
        <div className="flex-1 overflow-y-auto p-3">
          {!sidebarCollapsed && (
            <>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-caption font-semibold uppercase tracking-wide">Teams</h3>
                <button
                  onClick={onCreateTeam}
                  className="p-1 rounded hover:bg-sidebar-accent transition-colors"
                  title="Create new team"
                >
                  <Plus className="w-4 h-4 text-muted-foreground" />
                </button>
              </div>
              <div className="space-y-1">
                {/* All Teams Button */}
                <button
                  onClick={() => onSelectTeam(null)}
                  className={cn(
                    'w-full p-3 rounded-lg transition-colors flex items-center gap-3',
                    !selectedTeam
                      ? 'bg-primary/10 ring-1 ring-primary/20'
                      : 'hover:bg-sidebar-accent'
                  )}
                >
                  <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-muted-foreground to-muted-foreground/70 flex items-center justify-center">
                    <Users className="w-4 h-4 text-white" />
                  </div>
                  <div className="flex-1 text-left">
                    <p className="text-sm font-medium">All Teams</p>
                    <p className="text-xs text-muted-foreground">{teams.length} teams</p>
                  </div>
                </button>
                {/* Team Cards */}
                {teams.map((team) => (
                  <TeamCard
                    key={team.id}
                    team={team}
                    isSelected={selectedTeam?.id === team.id}
                    onClick={() => onSelectTeam(team)}
                  />
                ))}
              </div>
            </>
          )}
          {sidebarCollapsed && (
            <div className="space-y-2">
              {/* All Teams Button - Collapsed */}
              <button
                onClick={() => onSelectTeam(null)}
                className={cn(
                  'w-full p-2 rounded-lg transition-colors',
                  !selectedTeam
                    ? 'bg-primary/10'
                    : 'hover:bg-sidebar-accent'
                )}
                title="All Teams"
              >
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-muted-foreground to-muted-foreground/70 flex items-center justify-center mx-auto">
                  <Users className="w-4 h-4 text-white" />
                </div>
              </button>
              {teams.map((team) => (
                <button
                  key={team.id}
                  onClick={() => onSelectTeam(team)}
                  className={cn(
                    'w-full p-2 rounded-lg transition-colors',
                    selectedTeam?.id === team.id
                      ? 'bg-primary/10'
                      : 'hover:bg-sidebar-accent'
                  )}
                  title={team.name}
                >
                  <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center text-primary-foreground font-semibold text-xs mx-auto">
                    {team.name.slice(0, 2).toUpperCase()}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* User Menu */}
        {!sidebarCollapsed && currentUser && (
          <div className="p-3 border-t border-sidebar-border">
            <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-sidebar-accent transition-colors cursor-pointer">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-layer-user to-layer-user-dark flex items-center justify-center text-white text-sm font-medium">
                {currentUser.initials}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{currentUser.name}</p>
                <p className="text-caption truncate">{currentUser.role}</p>
              </div>
              <Settings className="w-4 h-4 text-muted-foreground" />
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 w-full mt-2 p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors"
            >
              <LogOut className="w-4 h-4" />
              <span className="text-sm">Logout</span>
            </button>
          </div>
        )}
        {sidebarCollapsed && (
          <div className="p-3 border-t border-sidebar-border">
            <button
              onClick={handleLogout}
              className="w-full p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors flex items-center justify-center"
              title="Logout"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        )}
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0">
        {/* Top Bar */}
        <header className="bg-card border-b h-14 flex items-center justify-between px-4 gap-4">
          {/* Selected Team Name */}
          <div className="flex items-center gap-3 min-w-0">
            {selectedTeam ? (
              <>
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center text-primary-foreground font-semibold text-xs flex-shrink-0">
                  {selectedTeam.name.slice(0, 2).toUpperCase()}
                </div>
                <div className="min-w-0">
                  <h1 className="font-semibold truncate">{selectedTeam.name}</h1>
                </div>
              </>
            ) : (
              <>
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-muted-foreground to-muted-foreground/70 flex items-center justify-center flex-shrink-0">
                  <Users className="w-4 h-4 text-white" />
                </div>
                <h1 className="font-semibold">All Teams</h1>
              </>
            )}
          </div>

          {/* Search */}
          <button
            onClick={() => setCommandPaletteOpen(true)}
            className="flex-1 max-w-md relative flex items-center gap-2 px-3 py-2 bg-muted/50 rounded-md text-muted-foreground hover:bg-muted/70 transition-colors text-left"
          >
            <Search className="w-4 h-4 flex-shrink-0" />
            <span className="flex-1 text-sm">Search rules, teams...</span>
            <kbd className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 text-xs bg-muted rounded">
              <span className="text-[10px]">⌘</span>K
            </kbd>
          </button>

          {/* Actions */}
          <div className="flex items-center gap-2">
            <NotificationBell onViewRuleHistory={onViewRuleHistory} />

            <Button size="sm" className="gap-1.5" onClick={onCreateRule}>
              <Plus className="w-4 h-4" />
              New Rule
            </Button>
          </div>
        </header>

        {/* Page Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {children}
        </div>
      </main>

      {/* Command Palette */}
      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
        rules={rules}
        teams={teams}
        onSelectRule={(rule) => {
          onSelectRule?.(rule);
          setCommandPaletteOpen(false);
        }}
        onSelectTeam={(team) => {
          onSelectTeam(team);
          setCommandPaletteOpen(false);
        }}
        onCreateRule={() => {
          onCreateRule?.();
          setCommandPaletteOpen(false);
        }}
        onCreateTeam={() => {
          onCreateTeam?.();
          setCommandPaletteOpen(false);
        }}
        onViewApprovals={() => {
          onViewApprovals?.();
          setCommandPaletteOpen(false);
        }}
        onViewAgents={() => {
          onViewAgents?.();
          setCommandPaletteOpen(false);
        }}
        onLogout={() => {
          handleLogout();
        }}
      />
    </div>
  );
}
