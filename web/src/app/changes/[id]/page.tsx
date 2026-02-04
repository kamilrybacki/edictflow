'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ChangeRequest } from '@/domain/change_request';
import { DiffViewer } from '@/components/DiffViewer';
import { useAuth, useRequireAuth } from '@/contexts/AuthContext';
import { fetchChange, approveChange, rejectChange } from '@/lib/api';
import { NotificationBell } from '@/components/NotificationBell';
import { UserMenu } from '@/components/UserMenu';

const statusColors = {
  pending: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  approved: 'bg-green-500/10 text-green-400 border-green-500/20',
  rejected: 'bg-red-500/10 text-red-400 border-red-500/20',
  auto_reverted: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  exception_granted: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

const statusLabels = {
  pending: 'Pending',
  approved: 'Approved',
  rejected: 'Rejected',
  auto_reverted: 'Auto-reverted',
  exception_granted: 'Exception Granted',
};

export default function ChangeDetailPage() {
  const params = useParams();
  const auth = useRequireAuth();
  const [change, setChange] = useState<ChangeRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const id = params.id as string;

  useEffect(() => {
    async function loadChange() {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchChange(id);
        setChange(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load change');
      } finally {
        setLoading(false);
      }
    }

    loadChange();
  }, [id]);

  const handleApprove = async () => {
    if (!change) return;
    setActionLoading(true);
    try {
      await approveChange(change.id);
      setChange({ ...change, status: 'approved' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve');
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async () => {
    if (!change) return;
    setActionLoading(true);
    try {
      await rejectChange(change.id);
      setChange({ ...change, status: 'rejected' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reject');
    } finally {
      setActionLoading(false);
    }
  };

  if (auth.isLoading || loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <div className="text-red-400 text-center">
          <p className="text-xl mb-4">Error loading change</p>
          <p className="text-sm">{error}</p>
          <Link href="/changes" className="text-blue-400 hover:underline mt-4 inline-block">
            Back to changes
          </Link>
        </div>
      </div>
    );
  }

  if (!change) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <div className="text-gray-400 text-center">
          <p className="text-xl mb-4">Change not found</p>
          <Link href="/changes" className="text-blue-400 hover:underline">
            Back to changes
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link href="/" className="text-xl font-bold">
                <span className="text-blue-400">Claude</span>ception
              </Link>
              <span className="text-gray-500">/</span>
              <Link href="/changes" className="text-gray-300 hover:text-white">
                Changes
              </Link>
              <span className="text-gray-500">/</span>
              <span className="text-gray-400 truncate max-w-xs">
                {change.file_path.split('/').pop()}
              </span>
            </div>
            <div className="flex items-center gap-4">
              <NotificationBell />
              <UserMenu />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Change Info */}
        <div className="bg-gray-800 rounded-lg p-6 mb-6">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-white mb-2">
                {change.file_path}
              </h1>
              <div className="flex items-center gap-4 text-sm text-gray-400">
                <span>Created: {new Date(change.created_at).toLocaleString()}</span>
                <span>Mode: {change.enforcement_mode}</span>
                {change.timeout_at && (
                  <span>Timeout: {new Date(change.timeout_at).toLocaleString()}</span>
                )}
              </div>
            </div>
            <div className="flex items-center gap-4">
              <span
                className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${
                  statusColors[change.status]
                }`}
              >
                {statusLabels[change.status]}
              </span>
              {change.status === 'pending' && auth.hasPermission('changes.approve') && (
                <div className="flex items-center gap-2">
                  <button
                    onClick={handleApprove}
                    disabled={actionLoading}
                    className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded font-medium disabled:opacity-50"
                  >
                    {actionLoading ? 'Processing...' : 'Approve'}
                  </button>
                  <button
                    onClick={handleReject}
                    disabled={actionLoading}
                    className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded font-medium disabled:opacity-50"
                  >
                    {actionLoading ? 'Processing...' : 'Reject'}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Diff Viewer */}
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Changes</h2>
          <DiffViewer diff={change.diff_content} fileName={change.file_path} />
        </div>

        {/* Hashes */}
        <div className="mt-6 bg-gray-800 rounded-lg p-6">
          <h2 className="text-lg font-semibold text-white mb-4">File Hashes</h2>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-gray-400">Original Hash:</span>
              <code className="ml-2 text-gray-300 font-mono">{change.original_hash}</code>
            </div>
            <div>
              <span className="text-gray-400">Modified Hash:</span>
              <code className="ml-2 text-gray-300 font-mono">{change.modified_hash}</code>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
