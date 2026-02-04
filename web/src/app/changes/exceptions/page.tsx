'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { ExceptionRequest, ExceptionRequestStatus } from '@/domain/change_request';
import { useAuth, useRequireAuth } from '@/contexts/AuthContext';
import { fetchExceptions, approveException, denyException } from '@/lib/api';
import { NotificationBell } from '@/components/NotificationBell';
import { UserMenu } from '@/components/UserMenu';

const statusColors: Record<ExceptionRequestStatus, string> = {
  pending: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  approved: 'bg-green-500/10 text-green-400 border-green-500/20',
  denied: 'bg-red-500/10 text-red-400 border-red-500/20',
};

const statusLabels: Record<ExceptionRequestStatus, string> = {
  pending: 'Pending',
  approved: 'Approved',
  denied: 'Denied',
};

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function ExceptionsPage() {
  const auth = useRequireAuth();
  const [exceptions, setExceptions] = useState<ExceptionRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string>('pending');
  const [error, setError] = useState<string | null>(null);

  const teamId = auth.user?.team_id;

  useEffect(() => {
    if (!teamId) return;

    async function loadExceptions() {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchExceptions(teamId, statusFilter || undefined);
        setExceptions(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load exceptions');
      } finally {
        setLoading(false);
      }
    }

    loadExceptions();
  }, [teamId, statusFilter]);

  const handleApprove = async (id: string) => {
    try {
      await approveException(id);
      setExceptions(exceptions.filter((e) => e.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve');
    }
  };

  const handleDeny = async (id: string) => {
    try {
      await denyException(id);
      setExceptions(exceptions.filter((e) => e.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to deny');
    }
  };

  if (auth.isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
      </div>
    );
  }

  const statusOptions: { value: string; label: string }[] = [
    { value: 'pending', label: 'Pending' },
    { value: '', label: 'All' },
    { value: 'approved', label: 'Approved' },
    { value: 'denied', label: 'Denied' },
  ];

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
              <span className="text-gray-300">Exceptions</span>
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
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-white">Exception Requests</h1>
          <div className="flex bg-gray-800 rounded-lg p-1">
            {statusOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => setStatusFilter(option.value)}
                className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                  statusFilter === option.value
                    ? 'bg-blue-600 text-white'
                    : 'text-gray-400 hover:text-white'
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-900/20 border border-red-500/50 rounded-lg text-red-400">
            {error}
          </div>
        )}

        <div className="bg-gray-800 rounded-lg overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
            </div>
          ) : exceptions.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              No exception requests found
            </div>
          ) : (
            <div className="divide-y divide-gray-700">
              {exceptions.map((exception) => (
                <div key={exception.id} className="p-4 hover:bg-gray-700/50">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <span
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${
                            statusColors[exception.status]
                          }`}
                        >
                          {statusLabels[exception.status]}
                        </span>
                        <span className="text-sm text-gray-400">
                          {exception.exception_type === 'time_limited'
                            ? 'Time Limited'
                            : 'Permanent'}
                        </span>
                        <span className="text-sm text-gray-500">
                          {formatDate(exception.created_at)}
                        </span>
                      </div>
                      <p className="text-gray-300 mb-2">{exception.justification}</p>
                      <Link
                        href={`/changes/${exception.change_request_id}`}
                        className="text-sm text-blue-400 hover:text-blue-300"
                      >
                        View original change request
                      </Link>
                    </div>
                    {exception.status === 'pending' &&
                      auth.hasPermission('exceptions.approve') && (
                        <div className="flex items-center gap-2 ml-4">
                          <button
                            onClick={() => handleApprove(exception.id)}
                            className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-sm font-medium"
                          >
                            Approve
                          </button>
                          <button
                            onClick={() => handleDeny(exception.id)}
                            className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-sm font-medium"
                          >
                            Deny
                          </button>
                        </div>
                      )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
