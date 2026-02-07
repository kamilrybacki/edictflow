'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { Clock, FileCheck, Shield, Wifi, ArrowRight } from 'lucide-react';
import { useRequireAuth } from '@/contexts/AuthContext';
import { Rule, TargetLayer } from '@/domain/rule';
import { TeamData, TeamMember } from '@/domain/team';
import {
  DashboardLayout,
  StatCard,
  SystemHealth,
  RuleHierarchy,
  RuleCard,
  Activity,
  CreateTeamDialog,
  RuleDetailsPanel,
  ActivitySidebar,
} from '@/components/dashboard';
import { fetchWorkerHealth, fetchConnectedAgents } from '@/lib/api/agents';
import { createTeam } from '@/lib/api';
import { layerConfig } from '@/lib/layerConfig';
import { cn } from '@/lib/utils';

// Lazy load heavy modal components
const RuleEditor = dynamic(() => import('@/components/RuleEditor').then(mod => ({ default: mod.RuleEditor })), {
  loading: () => <div className="flex items-center justify-center p-8"><div className="animate-spin h-8 w-8 border-2 border-primary rounded-full border-t-transparent" /></div>,
  ssr: false,
});

const AgentListModal = dynamic(() => import('@/components/dashboard/AgentListModal').then(mod => ({ default: mod.AgentListModal })), {
  loading: () => null,
  ssr: false,
});

const RuleHistoryPanel = dynamic(() => import('@/components/dashboard/RuleHistoryPanel').then(mod => ({ default: mod.RuleHistoryPanel })), {
  loading: () => <div className="animate-pulse bg-muted h-32 rounded-lg" />,
  ssr: false,
});

export default function Dashboard() {
  const auth = useRequireAuth();
  const router = useRouter();

  const [teams, setTeams] = useState<TeamData[]>([]);
  const [rules, setRules] = useState<Rule[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<TeamData | null>(null);
  const [selectedRule, setSelectedRule] = useState<Rule | undefined>();
  const [selectedLayer, setSelectedLayer] = useState<TargetLayer | undefined>();
  const [showRuleEditor, setShowRuleEditor] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(true);
  const [showAgentModal, setShowAgentModal] = useState(false);
  const [showCreateTeamDialog, setShowCreateTeamDialog] = useState(false);
  const [showRuleHistory, setShowRuleHistory] = useState(false);
  const [agentCount, setAgentCount] = useState(0);
  const [rulesRefreshKey, setRulesRefreshKey] = useState(0);
  const [highlightedRuleId, setHighlightedRuleId] = useState<string | undefined>();

  // Helper to highlight a rule temporarily
  const highlightRule = useCallback((ruleId: string) => {
    setHighlightedRuleId(ruleId);
    setTimeout(() => setHighlightedRuleId(undefined), 1500);
  }, []);

  // Fetch teams from API
  useEffect(() => {
    async function fetchTeams() {
      try {
        const response = await fetch('/api/teams');
        if (response.ok) {
          const teamsData = await response.json();
          // Fetch members and rules for each team in parallel
          const transformedTeams: TeamData[] = await Promise.all(
            teamsData.map(async (team: { id: string; name: string; settings?: { inherit_global_rules?: boolean } }) => {
              // Fetch team members and rules in parallel
              const [membersResult, rulesResult] = await Promise.allSettled([
                fetch(`/api/users?team_id=${team.id}&active_only=true`),
                fetch(`/api/rules?team_id=${team.id}`),
              ]);

              let members: TeamMember[] = [];
              if (membersResult.status === 'fulfilled' && membersResult.value.ok) {
                const usersData = await membersResult.value.json();
                members = (usersData || []).map((user: { id: string; name: string; avatar_url?: string }) => ({
                  id: user.id,
                  name: user.name,
                  avatar: user.avatar_url,
                  status: 'online' as const,
                }));
              }

              // Count rules by target layer
              const rulesCount: Record<TargetLayer, number> = { organization: 0, team: 0, project: 0 };
              if (rulesResult.status === 'fulfilled' && rulesResult.value.ok) {
                const rulesData = await rulesResult.value.json();
                (rulesData || []).forEach((rule: { target_layer?: string; targetLayer?: string }) => {
                  const layer = (rule.target_layer || rule.targetLayer) as TargetLayer;
                  if (layer && rulesCount[layer] !== undefined) {
                    rulesCount[layer]++;
                  }
                });
              }

              return {
                id: team.id,
                name: team.name,
                members,
                rulesCount,
                inheritGlobalRules: team.settings?.inherit_global_rules ?? true,
                notifications: { slack: false, email: false },
              };
            })
          );
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
      setIsLoading(true);
      try {
        if (selectedTeam) {
          // Fetch team-specific rules
          const response = await fetch(`/api/rules?team_id=${selectedTeam.id}`);
          if (response.ok) {
            const data = await response.json();
            setRules(data || []);
          } else {
            setRules([]);
          }
        } else {
          // Aggregate rules from all teams
          const allRules: Rule[] = [];
          await Promise.all(
            teams.map(async (team) => {
              try {
                const response = await fetch(`/api/rules?team_id=${team.id}`);
                if (response.ok) {
                  const data = await response.json();
                  if (Array.isArray(data)) {
                    allRules.push(...data);
                  }
                }
              } catch {
                // Ignore errors for individual teams
              }
            })
          );
          setRules(allRules);
        }
      } catch (error) {
        console.error('Failed to fetch rules:', error);
        setRules([]);
      } finally {
        setIsLoading(false);
      }
    }

    if (auth.isAuthenticated && (selectedTeam || teams.length > 0)) {
      fetchRules();
    }
  }, [auth.isAuthenticated, selectedTeam, teams, rulesRefreshKey]);

  // Fetch agent count from worker (filtered by selected team)
  useEffect(() => {
    async function fetchAgentCountForTeam() {
      try {
        if (selectedTeam) {
          // Fetch agents filtered by team
          const agents = await fetchConnectedAgents(selectedTeam.id);
          setAgentCount(agents.length);
        } else {
          // Fetch all agents
          const health = await fetchWorkerHealth();
          setAgentCount(health.agents);
        }
      } catch (error) {
        console.error('Failed to fetch agent count:', error);
        setAgentCount(0);
      }
    }

    if (auth.isAuthenticated) {
      fetchAgentCountForTeam();
      // Refresh every 30 seconds
      const interval = setInterval(fetchAgentCountForTeam, 30000);
      return () => clearInterval(interval);
    }
  }, [auth.isAuthenticated, selectedTeam]);

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

  const handleCreateTeam = async (name: string) => {
    const team = await createTeam(name);
    const newTeamData: TeamData = {
      id: team.id,
      name: team.name,
      members: [],
      rulesCount: { organization: 0, team: 0, project: 0 },
      inheritGlobalRules: team.settings?.inheritGlobalRules ?? true,
      notifications: { slack: false, email: false },
    };
    setTeams([...teams, newTeamData]);
    setSelectedTeam(newTeamData);
  };

  const handleEditRule = (rule: Rule) => {
    setEditingRule(rule);
    setShowRuleEditor(true);
  };

  const handleRuleSaved = () => {
    setShowRuleEditor(false);
    setEditingRule(undefined);
    // Trigger rules refresh
    setRulesRefreshKey(prev => prev + 1);
  };

  const handleViewRuleHistory = async (ruleId: string) => {
    // Find the rule in the current rules list
    const rule = rules.find(r => r.id === ruleId);
    if (rule) {
      setSelectedRule(rule);
      setShowRuleHistory(true);
      highlightRule(rule.id);
    } else {
      // If not found locally, try to fetch it
      try {
        const response = await fetch(`/api/rules?id=${ruleId}`);
        if (response.ok) {
          const fetchedRule = await response.json();
          if (fetchedRule) {
            setSelectedRule(fetchedRule);
            setShowRuleHistory(true);
            highlightRule(fetchedRule.id);
          }
        }
      } catch (error) {
        console.error('Failed to fetch rule:', error);
      }
    }
  };

  // Memoize filtered rules and stats to prevent recalculation on every render
  const filteredRules = useMemo(
    () => selectedLayer ? rules.filter(r => r.targetLayer === selectedLayer) : rules,
    [rules, selectedLayer]
  );

  // Calculate stats with useMemo
  const { pendingApprovals, activeRules, blockedRules } = useMemo(() => ({
    pendingApprovals: rules.filter(r => r.status === 'pending').length,
    activeRules: rules.filter(r => r.status === 'approved').length,
    blockedRules: rules.filter(r => r.enforcementMode === 'block').length,
  }), [rules]);

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
      currentUser={currentUser}
      onCreateRule={handleCreateRule}
      onCreateTeam={() => setShowCreateTeamDialog(true)}
      onViewRuleHistory={handleViewRuleHistory}
      rules={rules}
      onSelectRule={(rule) => {
        setSelectedRule(rule);
        setShowRuleHistory(false);
        highlightRule(rule.id);
      }}
      onViewApprovals={() => router.push('/approvals')}
      onViewAgents={() => setShowAgentModal(true)}
    >
      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard
          title="Pending Approvals"
          value={pendingApprovals}
          icon={<Clock className="w-5 h-5 text-status-pending" />}
          variant="warning"
          onClick={() => router.push('/approvals')}
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
        <div className="lg:col-span-1 relative">
          <div className="bg-card rounded-xl border p-4">
            <h2 className="text-heading mb-4">Rule Hierarchy</h2>
            <RuleHierarchy
              rules={rules}
              selectedRule={selectedRule}
              onSelectRule={setSelectedRule}
              onSelectLayer={setSelectedLayer}
            />
          </div>
          {/* Connection Arrow - visible on lg screens when layer selected */}
          {selectedLayer && (
            <div className="hidden lg:flex absolute top-5 -right-3 z-10">
              <div className={cn(
                'flex items-center justify-center w-6 h-6 rounded-full shadow-md',
                layerConfig[selectedLayer].className
              )}>
                <ArrowRight className="w-3.5 h-3.5" />
              </div>
            </div>
          )}
        </div>

        {/* Rules Grid - Middle Column */}
        <div className="lg:col-span-1">
          {/* Header with layer indicator */}
          <div className={cn(
            'rounded-xl border p-4 transition-all duration-300',
            selectedLayer
              ? `${layerConfig[selectedLayer].bgClassName} ${layerConfig[selectedLayer].borderClassName}`
              : 'bg-card'
          )}>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                {selectedLayer && (
                  <div className={cn(
                    'w-8 h-8 rounded-lg flex items-center justify-center',
                    layerConfig[selectedLayer].className
                  )}>
                    {(() => {
                      const Icon = layerConfig[selectedLayer].icon;
                      return <Icon className="w-4 h-4" />;
                    })()}
                  </div>
                )}
                <h2 className="text-heading">
                  {selectedLayer
                    ? `${layerConfig[selectedLayer].label} Rules`
                    : 'All Rules'
                  }
                </h2>
              </div>
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
                <div className="text-center py-8 text-muted-foreground">
                  {selectedLayer
                    ? `No ${layerConfig[selectedLayer].label.toLowerCase()} rules found`
                    : 'No rules found'
                  }
                </div>
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
                      isHighlighted={highlightedRuleId === rule.id}
                      onClick={() => {
                        setSelectedRule(rule);
                        setShowRuleHistory(false);
                      }}
                      onEdit={handleEditRule}
                      onViewHistory={(r) => {
                        setSelectedRule(r);
                        setShowRuleHistory(true);
                      }}
                      onViewDetails={(r) => {
                        setSelectedRule(r);
                        setShowRuleHistory(false);
                      }}
                    />
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Right Column - Rule History or Rule Details */}
        <div className="lg:col-span-1">
          {selectedRule && showRuleHistory ? (
            <RuleHistoryPanel
              rule={selectedRule}
              onClose={() => {
                setSelectedRule(undefined);
                setShowRuleHistory(false);
              }}
              onBack={() => setShowRuleHistory(false)}
            />
          ) : selectedRule ? (
            <RuleDetailsPanel
              rule={selectedRule}
              onClose={() => setSelectedRule(undefined)}
              onEdit={handleEditRule}
              onViewHistory={() => setShowRuleHistory(true)}
            />
          ) : (
            <div className="bg-card/50 rounded-xl border border-dashed p-8 text-center">
              <div className="text-muted-foreground">
                <Shield className="w-12 h-12 mx-auto mb-3 opacity-30" />
                <p className="text-sm">Select a rule to view details</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Activity Sidebar */}
      <ActivitySidebar activities={activities} />

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
        teamId={selectedTeam?.id}
        teamName={selectedTeam?.name}
      />

      {/* Create Team Dialog */}
      <CreateTeamDialog
        isOpen={showCreateTeamDialog}
        onClose={() => setShowCreateTeamDialog(false)}
        onCreateTeam={handleCreateTeam}
      />
    </DashboardLayout>
  );
}
