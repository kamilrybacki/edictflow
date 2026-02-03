'use client';

import { useState, useEffect } from 'react';
import { Rule, TargetLayer, RuleStatus, getTargetLayerPath, getStatusColor } from '@/domain/rule';
import { fetchRules, deleteRule, submitRuleForApproval, getApprovalStatus, ApprovalStatus } from '@/lib/api';

interface RuleListProps {
  teamId: string;
  teamName: string;
  onCreateRule: () => void;
  onEditRule: (rule: Rule) => void;
  refreshKey?: number;
}

const targetLayerColors: Record<TargetLayer, string> = {
  enterprise: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
  global: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  project: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  local: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
};

export function RuleList({ teamId, teamName, onCreateRule, onEditRule, refreshKey }: RuleListProps) {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedRuleId, setExpandedRuleId] = useState<string | null>(null);
  const [approvalStatuses, setApprovalStatuses] = useState<Record<string, ApprovalStatus>>({});

  const loadRules = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchRules(teamId);
      setRules(data);

      // Load approval status for pending rules
      const pendingRules = data.filter(r => r.status === 'pending');
      const statuses: Record<string, ApprovalStatus> = {};
      await Promise.all(
        pendingRules.map(async (rule) => {
          try {
            statuses[rule.id] = await getApprovalStatus(rule.id);
          } catch {
            // Ignore errors
          }
        })
      );
      setApprovalStatuses(statuses);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load rules');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRules();
  }, [teamId, refreshKey]);

  const handleDeleteRule = async (rule: Rule) => {
    if (rule.status !== 'draft') {
      alert('Only draft rules can be deleted');
      return;
    }
    if (!confirm(`Delete rule "${rule.name}"?`)) {
      return;
    }

    try {
      await deleteRule(rule.id);
      setRules(rules.filter((r) => r.id !== rule.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete rule');
    }
  };

  const handleSubmitForApproval = async (rule: Rule) => {
    if (!confirm(`Submit "${rule.name}" for approval?`)) {
      return;
    }

    try {
      await submitRuleForApproval(rule.id);
      await loadRules();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to submit for approval');
    }
  };

  const canEdit = (rule: Rule) => rule.status === 'draft' || rule.status === 'rejected';
  const canDelete = (rule: Rule) => rule.status === 'draft';
  const canSubmit = (rule: Rule) => rule.status === 'draft' || rule.status === 'rejected';

  return (
    <div className="flex flex-col h-full">
      <div className="p-4 border-b border-zinc-200 dark:border-zinc-800 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Rules</h2>
          <p className="text-sm text-zinc-500">Team: {teamName}</p>
        </div>
        <button
          onClick={onCreateRule}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Rule
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="text-center text-zinc-500">Loading rules...</div>
        ) : error ? (
          <div>
            <div className="text-red-500 text-sm mb-2">{error}</div>
            <button onClick={loadRules} className="text-sm text-blue-600 hover:underline">
              Retry
            </button>
          </div>
        ) : rules.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-zinc-400 mb-4">
              <svg className="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </div>
            <p className="text-zinc-500 mb-4">No rules yet for this team.</p>
            <button
              onClick={onCreateRule}
              className="px-4 py-2 text-sm font-medium text-blue-600 border border-blue-600 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20"
            >
              Create your first rule
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {rules.map((rule) => (
              <div
                key={rule.id}
                className="border border-zinc-200 dark:border-zinc-700 rounded-lg overflow-hidden"
              >
                <div
                  className="p-4 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
                  onClick={() => setExpandedRuleId(expandedRuleId === rule.id ? null : rule.id)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1 flex-wrap">
                        <h3 className="font-medium">{rule.name}</h3>
                        <span
                          className={`px-2 py-0.5 text-xs font-medium rounded ${
                            targetLayerColors[rule.targetLayer as TargetLayer]
                          }`}
                        >
                          {rule.targetLayer}
                        </span>
                        <span
                          className={`px-2 py-0.5 text-xs font-medium rounded ${getStatusColor(rule.status as RuleStatus)}`}
                        >
                          {rule.status}
                        </span>
                        {rule.priorityWeight > 0 && (
                          <span className="px-2 py-0.5 text-xs font-medium rounded bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
                            Priority: {rule.priorityWeight}
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-zinc-500 truncate">
                        {getTargetLayerPath(rule.targetLayer as TargetLayer)}
                      </p>

                      {/* Approval Progress for Pending Rules */}
                      {rule.status === 'pending' && approvalStatuses[rule.id] && (
                        <div className="mt-2 flex items-center gap-2">
                          <div className="flex-1 h-2 bg-zinc-200 dark:bg-zinc-700 rounded-full overflow-hidden">
                            <div
                              className="h-full bg-blue-500 transition-all"
                              style={{
                                width: `${(approvalStatuses[rule.id].current_count / approvalStatuses[rule.id].required_count) * 100}%`,
                              }}
                            />
                          </div>
                          <span className="text-xs text-zinc-500">
                            {approvalStatuses[rule.id].current_count}/{approvalStatuses[rule.id].required_count} approvals
                          </span>
                        </div>
                      )}

                      {rule.triggers.length > 0 && (
                        <div className="flex gap-1 mt-2 flex-wrap">
                          {rule.triggers.map((trigger, idx) => (
                            <span
                              key={idx}
                              className="px-2 py-0.5 text-xs rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400"
                            >
                              {trigger.type}: {trigger.pattern || trigger.contextTypes?.join(', ') || trigger.tags?.join(', ')}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                    <div className="flex items-center gap-2 ml-4">
                      {canSubmit(rule) && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleSubmitForApproval(rule);
                          }}
                          className="px-2 py-1 text-xs font-medium text-blue-600 border border-blue-600 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20"
                          title="Submit for approval"
                        >
                          Submit
                        </button>
                      )}
                      {canEdit(rule) && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            onEditRule(rule);
                          }}
                          className="p-1 text-zinc-400 hover:text-blue-500"
                          title="Edit rule"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                            />
                          </svg>
                        </button>
                      )}
                      {canDelete(rule) && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteRule(rule);
                          }}
                          className="p-1 text-zinc-400 hover:text-red-500"
                          title="Delete rule"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                          </svg>
                        </button>
                      )}
                      <svg
                        className={`w-4 h-4 text-zinc-400 transition-transform ${
                          expandedRuleId === rule.id ? 'rotate-180' : ''
                        }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                      </svg>
                    </div>
                  </div>
                </div>

                {expandedRuleId === rule.id && (
                  <div className="border-t border-zinc-200 dark:border-zinc-700 p-4 bg-zinc-50 dark:bg-zinc-800/30">
                    <h4 className="text-sm font-medium mb-2">Content:</h4>
                    <pre className="text-sm bg-white dark:bg-zinc-900 p-3 rounded border border-zinc-200 dark:border-zinc-700 overflow-x-auto whitespace-pre-wrap">
                      {rule.content}
                    </pre>
                    <div className="mt-3 text-xs text-zinc-500 space-y-1">
                      <div>Created: {new Date(rule.createdAt).toLocaleString()}</div>
                      <div>Updated: {new Date(rule.updatedAt).toLocaleString()}</div>
                      {rule.submittedAt && (
                        <div>Submitted: {new Date(rule.submittedAt).toLocaleString()}</div>
                      )}
                      {rule.approvedAt && (
                        <div>Approved: {new Date(rule.approvedAt).toLocaleString()}</div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
