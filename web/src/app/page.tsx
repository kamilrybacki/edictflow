'use client';

import { useState } from 'react';
import { Team } from '@/domain/team';
import { Rule } from '@/domain/rule';
import { StatusBadge } from '@/components/StatusBadge';
import { TeamList } from '@/components/TeamList';
import { RuleList } from '@/components/RuleList';
import { RuleEditor } from '@/components/RuleEditor';

export default function Dashboard() {
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [showRuleEditor, setShowRuleEditor] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | undefined>(undefined);
  const [refreshKey, setRefreshKey] = useState(0);

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

  return (
    <div className="flex flex-col h-screen bg-zinc-50 dark:bg-zinc-950">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold tracking-tight">
            <span className="text-blue-600">Claude</span>ception
          </h1>
          <span className="text-sm text-zinc-500">Rule Management</span>
        </div>
        <StatusBadge />
      </header>

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - Teams */}
        <aside className="w-72 border-r border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
          <TeamList selectedTeamId={selectedTeam?.id || null} onSelectTeam={setSelectedTeam} />
        </aside>

        {/* Main Area - Rules */}
        <main className="flex-1 overflow-hidden bg-white dark:bg-zinc-900">
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
              <p className="text-zinc-500 max-w-sm">
                Choose a team from the sidebar to view and manage its rules, or create a new team to
                get started.
              </p>
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
