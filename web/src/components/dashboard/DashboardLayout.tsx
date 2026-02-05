'use client';

import { ReactNode, useState } from 'react';
import {
  Search,
  Bell,
  Settings,
  Command,
  Plus,
  LayoutGrid,
  List,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button, Input } from '@/components/ui';
import { TeamCard } from './TeamCard';
import { TargetLayer } from '@/domain/rule';

interface TeamMember {
  id: string;
  name: string;
  avatar?: string;
}

interface TeamData {
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

interface UserData {
  name: string;
  role: string;
  initials: string;
}

interface DashboardLayoutProps {
  children: ReactNode;
  teams: TeamData[];
  selectedTeam?: TeamData | null;
  onSelectTeam: (team: TeamData) => void;
  viewMode: 'grid' | 'list';
  onViewModeChange: (mode: 'grid' | 'list') => void;
  currentUser?: UserData;
  onCreateRule?: () => void;
  notificationCount?: number;
}

export function DashboardLayout({
  children,
  teams,
  selectedTeam,
  onSelectTeam,
  viewMode,
  onViewModeChange,
  currentUser,
  onCreateRule,
  notificationCount = 0,
}: DashboardLayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

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
                <button className="p-1 rounded hover:bg-sidebar-accent transition-colors">
                  <Plus className="w-4 h-4 text-muted-foreground" />
                </button>
              </div>
              <div className="space-y-1">
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
          </div>
        )}
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0">
        {/* Top Bar */}
        <header className="bg-card border-b h-14 flex items-center justify-between px-4 gap-4">
          {/* Search */}
          <div className="flex-1 max-w-md relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              type="text"
              placeholder="Search rules, teams... (⌘K)"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 bg-muted/50 border-0 focus-visible:ring-1"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2">
            {/* View Toggle */}
            <div className="flex items-center border rounded-lg p-0.5 bg-muted/30">
              <button
                onClick={() => onViewModeChange('grid')}
                className={cn(
                  'p-1.5 rounded transition-colors',
                  viewMode === 'grid' ? 'bg-card shadow-subtle' : 'hover:bg-muted'
                )}
              >
                <LayoutGrid className="w-4 h-4" />
              </button>
              <button
                onClick={() => onViewModeChange('list')}
                className={cn(
                  'p-1.5 rounded transition-colors',
                  viewMode === 'list' ? 'bg-card shadow-subtle' : 'hover:bg-muted'
                )}
              >
                <List className="w-4 h-4" />
              </button>
            </div>

            <button className="p-2 rounded-lg hover:bg-muted transition-colors relative">
              <Bell className="w-5 h-5 text-muted-foreground" />
              {notificationCount > 0 && (
                <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-status-pending" />
              )}
            </button>

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
    </div>
  );
}
