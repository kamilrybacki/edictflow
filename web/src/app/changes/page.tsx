'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { ChangeRequest, ChangeRequestStatus } from '@/domain/change_request';
import { ChangeRequestTable } from '@/components/ChangeRequestTable';
import { useAuth, useRequireAuth } from '@/contexts/AuthContext';
import { fetchChanges, approveChange, rejectChange } from '@/lib/api';
import { NotificationBell } from '@/components/NotificationBell';
import { UserMenu } from '@/components/UserMenu';

export default function ChangesPage() {
  const auth = useRequireAuth();
  const [changes, setChanges] = useState<ChangeRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string>('pending');
  const [error, setError] = useState<string | null>(null);

  const teamId = auth.user?.teamId;

  useEffect(() => {
    if (!teamId) return;

    async function loadChanges(tid: string) {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchChanges(tid, statusFilter || undefined);
        setChanges(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load changes');
      } finally {
        setLoading(false);
      }
    }

    loadChanges(teamId);
  }, [teamId, statusFilter]);

  const handleApprove = async (id: string) => {
    try {
      await approveChange(id);
      setChanges(changes.filter((c) => c.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve');
    }
  };

  const handleReject = async (id: string) => {
    try {
      await rejectChange(id);
      setChanges(changes.filter((c) => c.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reject');
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
    { value: 'rejected', label: 'Rejected' },
    { value: 'auto_reverted', label: 'Auto-reverted' },
  ];

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link href="/" className="text-xl font-bold">
                Edictflow
              </Link>
              <span className="text-gray-500">/</span>
              <span className="text-gray-300">Changes</span>
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
          <h1 className="text-2xl font-bold text-white">Change Requests</h1>
          <div className="flex items-center gap-4">
            {/* Status filter tabs */}
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
            <Link
              href="/changes/exceptions"
              className="px-4 py-2 text-sm font-medium text-gray-400 hover:text-white bg-gray-800 rounded-lg"
            >
              Exceptions
            </Link>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-900/20 border border-red-500/50 rounded-lg text-red-400">
            {error}
          </div>
        )}

        <div className="bg-gray-800 rounded-lg overflow-hidden">
          <ChangeRequestTable
            changes={changes}
            loading={loading}
            onApprove={auth.hasPermission('changes.approve') ? handleApprove : undefined}
            onReject={auth.hasPermission('changes.approve') ? handleReject : undefined}
            showActions={auth.hasPermission('changes.approve')}
          />
        </div>
      </main>
    </div>
  );
}
