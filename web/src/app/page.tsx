'use client';

import { useState } from 'react';
import { Team } from '@/domain/team';
import { Rule } from '@/domain/rule';
import { StatusBadge } from '@/components/StatusBadge';
import { TeamList } from '@/components/TeamList';
import { RuleList } from '@/components/RuleList';
import { RuleEditor } from '@/components/RuleEditor';
import { RuleHierarchy } from '@/components/RuleHierarchy';
import { UserMenu } from '@/components/UserMenu';
import { NotificationBell } from '@/components/NotificationBell';
import { useRequireAuth } from '@/contexts/AuthContext';

export default function Dashboard() {
  const auth = useRequireAuth();

  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [showRuleEditor, setShowRuleEditor] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | undefined>(undefined);
  const [refreshKey, setRefreshKey] = useState(0);
  const [showHierarchy, setShowHierarchy] = useState(false);

  const handleCreateRule = () => {
    setEditingRule(undefined);
    setShowRuleEditor(true);
  };

  const handleEditRule = (rule: Rule) => {
    setEditingRule(rule);
    setShowRuleEditor(true);
  };

  const handleRuleSaved = () => {
    setShowRuleEditor(false);
    setEditingRule(undefined);
    setRefreshKey((k) => k + 1);
  };

  // Show loading state while checking auth, or nothing if redirecting to login
  if (auth.isLoading || !auth.isAuthenticated) {
    return (
      <div className="flex items-center justify-center h-screen bg-zinc-50 dark:bg-zinc-950">
        <div className="text-zinc-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-50 dark:bg-zinc-950">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold tracking-tight">
            Edictflow
          </h1>
          <span className="text-sm text-zinc-500">Rule Management</span>
        </div>
        <div className="flex items-center gap-4">
          <button
            onClick={() => setShowHierarchy(!showHierarchy)}
            className={`
              flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-colors
              ${showHierarchy
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800'
              }
            `}
            title="Toggle rule hierarchy view"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
            Hierarchy
          </button>
          <StatusBadge />
          <NotificationBell />
          <UserMenu />
        </div>
      </header>

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - Teams */}
        <aside className="w-72 border-r border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
          <TeamList selectedTeamId={selectedTeam?.id || null} onSelectTeam={setSelectedTeam} />
        </aside>

        {/* Main Area - Rules */}
        <main className="flex-1 overflow-hidden bg-white dark:bg-zinc-900 flex">
          {/* Content Area */}
          <div className="flex-1 overflow-hidden">
            {selectedTeam ? (
              <RuleList
                teamId={selectedTeam.id}
                teamName={selectedTeam.name}
                onCreateRule={handleCreateRule}
                onEditRule={handleEditRule}
                refreshKey={refreshKey}
              />
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-center p-8">
                <div className="text-zinc-400 mb-4">
                  <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1}
                      d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                    />
                  </svg>
                </div>
                <h2 className="text-xl font-semibold text-zinc-700 dark:text-zinc-300 mb-2">
                  Select a Team
                </h2>
                <p className="text-zinc-500 max-w-sm mb-6">
                  Choose a team from the sidebar to view and manage its rules, or create a new team to
                  get started.
                </p>
                <button
                  onClick={() => setShowHierarchy(!showHierarchy)}
                  className="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 text-sm font-medium"
                >
                  {showHierarchy ? 'Hide' : 'View'} Rule Hierarchy
                </button>
              </div>
            )}
          </div>

          {/* Hierarchy Panel - Collapsible Right Sidebar */}
          {(showHierarchy || selectedTeam) && (
            <div className={`
              border-l border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-900/50
              overflow-y-auto transition-all duration-300
              ${showHierarchy || selectedTeam ? 'w-96' : 'w-0'}
            `}>
              <div className="p-4">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold text-zinc-800 dark:text-zinc-200">
                    Rule Hierarchy
                  </h3>
                  <button
                    onClick={() => setShowHierarchy(false)}
                    className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 p-1"
                    title="Close hierarchy view"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <RuleHierarchy
                  teamId={selectedTeam?.id}
                  onRuleClick={(rule) => {
                    // Could navigate to rule or show details
                    console.log('Rule clicked:', rule);
                  }}
                />
              </div>
            </div>
          )}
        </main>
      </div>

      {/* Rule Editor Modal */}
      {showRuleEditor && selectedTeam && (
        <RuleEditor
          teamId={selectedTeam.id}
          rule={editingRule}
          onSave={handleRuleSaved}
          onCancel={() => {
            setShowRuleEditor(false);
            setEditingRule(undefined);
          }}
        />
      )}
    </div>
  );
}
