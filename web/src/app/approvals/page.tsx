'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { Rule, getStatusColor, getTargetLayerPath, TargetLayer } from '@/domain/rule';
import { fetchPendingApprovals, approveRule, rejectRule, getApprovalStatus, ApprovalStatus } from '@/lib/api';
import Link from 'next/link';

export default function ApprovalsPage() {
  const { isAuthenticated, isLoading: authLoading, hasAnyPermission } = useAuth();
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null);
  const [approvalStatus, setApprovalStatus] = useState<ApprovalStatus | null>(null);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const canApprove = hasAnyPermission('approve_local', 'approve_project', 'approve_global');

  useEffect(() => {
    if (!authLoading && isAuthenticated) {
      loadPendingRules();
    }
  }, [authLoading, isAuthenticated]);

  useEffect(() => {
    if (selectedRule) {
      loadApprovalStatus(selectedRule.id);
    }
  }, [selectedRule]);

  const loadPendingRules = async () => {
    try {
      setLoading(true);
      const data = await fetchPendingApprovals();
      setRules(data);
      if (data.length > 0 && !selectedRule) {
        setSelectedRule(data[0]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pending rules');
    } finally {
      setLoading(false);
    }
  };

  const loadApprovalStatus = async (ruleId: string) => {
    try {
      const status = await getApprovalStatus(ruleId);
      setApprovalStatus(status);
    } catch {
      setApprovalStatus(null);
    }
  };

  const handleApprove = async () => {
    if (!selectedRule) return;
    setSubmitting(true);
    try {
      await approveRule(selectedRule.id, comment);
      setComment('');
      await loadPendingRules();
      if (rules.length > 1) {
        setSelectedRule(rules.find(r => r.id !== selectedRule.id) || null);
      } else {
        setSelectedRule(null);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to approve');
    } finally {
      setSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!selectedRule) return;
    if (!comment.trim()) {
      alert('Please provide a reason for rejection');
      return;
    }
    setSubmitting(true);
    try {
      await rejectRule(selectedRule.id, comment);
      setComment('');
      await loadPendingRules();
      if (rules.length > 1) {
        setSelectedRule(rules.find(r => r.id !== selectedRule.id) || null);
      } else {
        setSelectedRule(null);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to reject');
    } finally {
      setSubmitting(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-900">
        <div className="text-zinc-600 dark:text-zinc-400">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-900">
        <div className="text-center">
          <p className="text-zinc-600 dark:text-zinc-400 mb-4">Please sign in to view approvals</p>
          <Link href="/login" className="text-blue-600 hover:underline">
            Sign in
          </Link>
        </div>
      </div>
    );
  }

  if (!canApprove) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-900">
        <div className="text-center">
          <p className="text-zinc-600 dark:text-zinc-400 mb-4">
            You don&apos;t have permission to approve rules
          </p>
          <Link href="/" className="text-blue-600 hover:underline">
            Back to Dashboard
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-900">
      {/* Header */}
      <header className="bg-white dark:bg-zinc-800 border-b border-zinc-200 dark:border-zinc-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center gap-4">
              <Link href="/" className="text-xl font-bold text-zinc-900 dark:text-white">
                Claudeception
              </Link>
              <span className="px-2 py-1 text-xs font-medium bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 rounded">
                Approvals
              </span>
            </div>
            <Link href="/" className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400">
              Back to Dashboard
            </Link>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-6">
          Pending Approvals ({rules.length})
        </h1>

        {loading ? (
          <div className="text-zinc-600 dark:text-zinc-400">Loading...</div>
        ) : error ? (
          <div className="text-red-600">{error}</div>
        ) : rules.length === 0 ? (
          <div className="bg-white dark:bg-zinc-800 rounded-lg border border-zinc-200 dark:border-zinc-700 p-12 text-center">
            <div className="text-zinc-400 mb-4">
              <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p className="text-zinc-500 text-lg">No pending approvals</p>
            <p className="text-zinc-400 text-sm mt-2">All caught up!</p>
          </div>
        ) : (
          <div className="flex gap-6">
            {/* Rules List */}
            <div className="w-80 flex-shrink-0 space-y-2">
              {rules.map((rule) => (
                <button
                  key={rule.id}
                  onClick={() => setSelectedRule(rule)}
                  className={`w-full text-left p-4 rounded-lg border transition-colors ${
                    selectedRule?.id === rule.id
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-700'
                  }`}
                >
                  <div className="font-medium text-zinc-900 dark:text-white">{rule.name}</div>
                  <div className="flex items-center gap-2 mt-1">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${getStatusColor('pending')}`}>
                      pending
                    </span>
                    <span className="text-xs text-zinc-500">{rule.targetLayer}</span>
                  </div>
                </button>
              ))}
            </div>

            {/* Rule Details */}
            {selectedRule && (
              <div className="flex-1 bg-white dark:bg-zinc-800 rounded-lg border border-zinc-200 dark:border-zinc-700 p-6">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h2 className="text-xl font-semibold text-zinc-900 dark:text-white">
                      {selectedRule.name}
                    </h2>
                    <p className="text-sm text-zinc-500 mt-1">
                      {getTargetLayerPath(selectedRule.targetLayer as TargetLayer)}
                    </p>
                  </div>
                  <span className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor('pending')}`}>
                    {selectedRule.targetLayer}
                  </span>
                </div>

                {/* Approval Progress */}
                {approvalStatus && (
                  <div className="mb-6 p-4 bg-zinc-50 dark:bg-zinc-900 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
                        Approval Progress
                      </span>
                      <span className="text-sm text-zinc-500">
                        {approvalStatus.current_count}/{approvalStatus.required_count} required
                      </span>
                    </div>
                    <div className="h-2 bg-zinc-200 dark:bg-zinc-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-blue-500"
                        style={{
                          width: `${(approvalStatus.current_count / approvalStatus.required_count) * 100}%`,
                        }}
                      />
                    </div>
                    {approvalStatus.approvals.length > 0 && (
                      <div className="mt-3 space-y-2">
                        {approvalStatus.approvals.map((approval) => (
                          <div key={approval.id} className="flex items-center gap-2 text-sm">
                            <span
                              className={`w-2 h-2 rounded-full ${
                                approval.decision === 'approved' ? 'bg-green-500' : 'bg-red-500'
                              }`}
                            />
                            <span className="text-zinc-600 dark:text-zinc-400">
                              {approval.user_name || approval.user_id.slice(0, 8)} - {approval.decision}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Rule Content */}
                <div className="mb-6">
                  <h3 className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">Content</h3>
                  <pre className="text-sm bg-zinc-50 dark:bg-zinc-900 p-4 rounded-lg border border-zinc-200 dark:border-zinc-700 overflow-x-auto whitespace-pre-wrap">
                    {selectedRule.content}
                  </pre>
                </div>

                {/* Comment */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">
                    Comment (required for rejection)
                  </label>
                  <textarea
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    rows={3}
                    className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md dark:bg-zinc-700 dark:text-white"
                    placeholder="Optional comment for approval, required for rejection..."
                  />
                </div>

                {/* Actions */}
                <div className="flex gap-3">
                  <button
                    onClick={handleApprove}
                    disabled={submitting}
                    className="flex-1 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-medium rounded-md"
                  >
                    {submitting ? 'Processing...' : 'Approve'}
                  </button>
                  <button
                    onClick={handleReject}
                    disabled={submitting}
                    className="flex-1 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white font-medium rounded-md"
                  >
                    {submitting ? 'Processing...' : 'Reject'}
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
