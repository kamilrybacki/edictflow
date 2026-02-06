'use client';

import { useState, useEffect } from 'react';
import { Clock, FileCheck, Shield, Wifi } from 'lucide-react';
import { useRequireAuth } from '@/contexts/AuthContext';
import { Rule, TargetLayer } from '@/domain/rule';
import { RuleEditor } from '@/components/RuleEditor';
import {
  DashboardLayout,
  StatCard,
  SystemHealth,
  RuleHierarchy,
  RuleCard,
  ActivityFeed,
  Activity,
  AgentListModal,
} from '@/components/dashboard';
import { fetchWorkerHealth } from '@/lib/api/agents';

// Temporary team data structure until API integration
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

export default function Dashboard() {
  const auth = useRequireAuth();

  const [teams, setTeams] = useState<TeamData[]>([]);
  const [rules, setRules] = useState<Rule[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<TeamData | null>(null);
  const [selectedRule, setSelectedRule] = useState<Rule | undefined>();
  const [selectedLayer, setSelectedLayer] = useState<TargetLayer | undefined>();
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [showRuleEditor, setShowRuleEditor] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(true);
  const [showAgentModal, setShowAgentModal] = useState(false);
  const [agentCount, setAgentCount] = useState(0);

  // Fetch teams from API
  useEffect(() => {
    async function fetchTeams() {
      try {
        const response = await fetch('/api/teams');
        if (response.ok) {
          const data = await response.json();
          // Transform API response to TeamData format
          const transformedTeams: TeamData[] = data.map((team: { id: string; name: string; settings?: { inherit_global_rules?: boolean } }) => ({
            id: team.id,
            name: team.name,
            members: [], // Would need separate API call
            rulesCount: { enterprise: 0, user: 0, project: 0 }, // Would need separate calculation
            inheritGlobalRules: team.settings?.inherit_global_rules ?? true,
            notifications: { slack: false, email: false },
          }));
          setTeams(transformedTeams);
        }
      } catch (error) {
        console.error('Failed to fetch teams:', error);
      }
    }

    if (auth.isAuthenticated) {
      fetchTeams();
    }
  }, [auth.isAuthenticated]);

  // Fetch rules from API
  useEffect(() => {
    async function fetchRules() {
      if (!selectedTeam) {
        setRules([]);
        setIsLoading(false);
        return;
      }

      try {
        const url = `/api/rules?team_id=${selectedTeam.id}`;
        const response = await fetch(url);
        if (response.ok) {
          const data = await response.json();
          setRules(data || []);
        } else {
          setRules([]);
        }
      } catch (error) {
        console.error('Failed to fetch rules:', error);
        setRules([]);
      } finally {
        setIsLoading(false);
      }
    }

    if (auth.isAuthenticated) {
      fetchRules();
    }
  }, [auth.isAuthenticated, selectedTeam]);

  // Fetch agent count from worker
  useEffect(() => {
    async function fetchAgentCount() {
      try {
        const health = await fetchWorkerHealth();
        setAgentCount(health.agents);
      } catch (error) {
        console.error('Failed to fetch agent count:', error);
        setAgentCount(0);
      }
    }

    if (auth.isAuthenticated) {
      fetchAgentCount();
      // Refresh every 30 seconds
      const interval = setInterval(fetchAgentCount, 30000);
      return () => clearInterval(interval);
    }
  }, [auth.isAuthenticated]);

  // Mock activities - would come from API
  useEffect(() => {
    setActivities([
      {
        id: '1',
        type: 'rule_approved',
        message: 'Security Standards rule was approved',
        timestamp: new Date(Date.now() - 1000 * 60 * 30),
        userName: 'Alex Rivera',
      },
      {
        id: '2',
        type: 'change_detected',
        message: 'Unauthorized change detected in CLAUDE.md',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2),
        userName: 'Jordan Kim',
      },
      {
        id: '3',
        type: 'rule_created',
        message: 'New API Documentation rule created',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 5),
        userName: 'Alex Rivera',
      },
      {
        id: '4',
        type: 'exception_granted',
        message: 'Exception granted for temporary deployment',
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
        userName: 'Sarah Chen',
      },
    ]);
  }, []);

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
    // Refresh rules
    setIsLoading(true);
  };

  // Filter rules by selected layer
  const filteredRules = selectedLayer
    ? rules.filter(r => r.targetLayer === selectedLayer)
    : rules;

  // Calculate stats
  const pendingApprovals = rules.filter(r => r.status === 'pending').length;
  const activeRules = rules.filter(r => r.status === 'approved').length;
  const blockedRules = rules.filter(r => r.enforcementMode === 'block').length;

  // Show loading state
  if (auth.isLoading || !auth.isAuthenticated) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  const currentUser = auth.user ? {
    name: auth.user.email?.split('@')[0] || 'User',
    role: 'Admin',
    initials: (auth.user.email?.slice(0, 2) || 'US').toUpperCase(),
  } : undefined;

  return (
    <DashboardLayout
      teams={teams}
      selectedTeam={selectedTeam}
      onSelectTeam={setSelectedTeam}
      viewMode={viewMode}
      onViewModeChange={setViewMode}
      currentUser={currentUser}
      onCreateRule={handleCreateRule}
      notificationCount={pendingApprovals}
    >
      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard
          title="Pending Approvals"
          value={pendingApprovals}
          icon={<Clock className="w-5 h-5 text-status-pending" />}
          variant="warning"
        />
        <StatCard
          title="Active Rules"
          value={activeRules}
          icon={<FileCheck className="w-5 h-5 text-status-approved" />}
          trend={{ value: 12, isPositive: true }}
        />
        <StatCard
          title="Blocking Rules"
          value={blockedRules}
          icon={<Shield className="w-5 h-5 text-enforce-block" />}
        />
        <StatCard
          title="Connected Agents"
          value={agentCount}
          icon={<Wifi className="w-5 h-5 text-layer-user" />}
          onClick={() => setShowAgentModal(true)}
        />
      </div>

      {/* System Health */}
      <div className="mb-6">
        <SystemHealth
          syncStatus="synced"
          agentsOnline={agentCount}
          pendingExceptions={2}
          errorsLast24h={0}
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Rule Hierarchy - Left Column */}
        <div className="lg:col-span-1">
          <div className="bg-card rounded-xl border p-4">
            <h2 className="text-heading mb-4">Rule Hierarchy</h2>
            <RuleHierarchy
              rules={rules}
              selectedRule={selectedRule}
              onSelectRule={setSelectedRule}
              onSelectLayer={setSelectedLayer}
            />
          </div>
        </div>

        {/* Rules Grid - Middle Column */}
        <div className="lg:col-span-1">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-heading">
              {selectedLayer
                ? `${selectedLayer.charAt(0).toUpperCase() + selectedLayer.slice(1)} Rules`
                : 'Enterprise Rules'
              }
            </h2>
            {selectedLayer && (
              <button
                onClick={() => setSelectedLayer(undefined)}
                className="text-xs text-primary hover:underline"
              >
                Show all
              </button>
            )}
          </div>
          <div className="space-y-3">
            {isLoading ? (
              <div className="text-center py-8 text-muted-foreground">Loading rules...</div>
            ) : filteredRules.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">No rules found</div>
            ) : (
              filteredRules.map((rule, index) => (
                <div
                  key={rule.id}
                  className="animate-fade-in"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <RuleCard
                    rule={rule}
                    isSelected={selectedRule?.id === rule.id}
                    onClick={() => setSelectedRule(rule)}
                    onEdit={handleEditRule}
                  />
                </div>
              ))
            )}
          </div>
        </div>

        {/* Activity Feed - Right Column */}
        <div className="lg:col-span-1">
          <div className="bg-card rounded-xl border p-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-heading">Recent Activity</h2>
              <button className="text-xs text-primary hover:underline">View all</button>
            </div>
            <ActivityFeed activities={activities} />
          </div>
        </div>
      </div>

      {/* Rule Editor Modal */}
      {showRuleEditor && (
        <RuleEditor
          teamId={selectedTeam?.id || ''}
          rule={editingRule}
          onSave={handleRuleSaved}
          onCancel={() => {
            setShowRuleEditor(false);
            setEditingRule(undefined);
          }}
        />
      )}

      {/* Agent List Modal */}
      <AgentListModal
        isOpen={showAgentModal}
        onClose={() => setShowAgentModal(false)}
      />
    </DashboardLayout>
  );
}
