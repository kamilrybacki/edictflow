'use client';

import React, { memo, useCallback } from 'react';
import Link from 'next/link';
import { ChangeRequest, ChangeRequestStatus, EnforcementMode } from '@/domain/change_request';

interface ChangeRequestTableProps {
  changes: ChangeRequest[];
  loading: boolean;
  onApprove?: (id: string) => void;
  onReject?: (id: string) => void;
  showActions?: boolean;
}

const statusColors: Record<ChangeRequestStatus, string> = {
  pending: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  approved: 'bg-green-500/10 text-green-400 border-green-500/20',
  rejected: 'bg-red-500/10 text-red-400 border-red-500/20',
  auto_reverted: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  exception_granted: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

const statusLabels: Record<ChangeRequestStatus, string> = {
  pending: 'Pending',
  approved: 'Approved',
  rejected: 'Rejected',
  auto_reverted: 'Auto-reverted',
  exception_granted: 'Exception Granted',
};

const enforcementModeLabels: Record<EnforcementMode, string> = {
  block: 'Block',
  temporary: 'Temporary',
  warning: 'Warning',
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

function getFileName(filePath: string): string {
  return filePath.split('/').pop() || filePath;
}

// Memoized table row to prevent re-renders
const ChangeRequestRow = memo(function ChangeRequestRow({
  change,
  onApprove,
  onReject,
  showActions,
}: {
  change: ChangeRequest;
  onApprove?: (id: string) => void;
  onReject?: (id: string) => void;
  showActions: boolean;
}) {
  const handleApprove = useCallback(() => onApprove?.(change.id), [onApprove, change.id]);
  const handleReject = useCallback(() => onReject?.(change.id), [onReject, change.id]);

  return (
    <tr className="hover:bg-gray-800/50 transition-colors">
      <td className="px-4 py-3">
        <Link
          href={`/changes/${change.id}`}
          className="text-blue-400 hover:text-blue-300"
        >
          <div className="font-medium">{getFileName(change.file_path)}</div>
          <div className="text-xs text-gray-500 truncate max-w-xs">
            {change.file_path}
          </div>
        </Link>
      </td>
      <td className="px-4 py-3">
        <span
          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${
            statusColors[change.status]
          }`}
        >
          {statusLabels[change.status]}
        </span>
      </td>
      <td className="px-4 py-3 text-gray-300">
        {enforcementModeLabels[change.enforcement_mode]}
      </td>
      <td className="px-4 py-3 text-gray-400">
        {formatDate(change.created_at)}
      </td>
      {showActions && (
        <td className="px-4 py-3 text-right">
          {change.status === 'pending' && (
            <div className="flex items-center justify-end gap-2">
              {onApprove && (
                <button
                  onClick={handleApprove}
                  className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-xs font-medium"
                >
                  Approve
                </button>
              )}
              {onReject && (
                <button
                  onClick={handleReject}
                  className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-xs font-medium"
                >
                  Reject
                </button>
              )}
            </div>
          )}
        </td>
      )}
    </tr>
  );
});

export const ChangeRequestTable = memo(function ChangeRequestTable({
  changes,
  loading,
  onApprove,
  onReject,
  showActions = true,
}: ChangeRequestTableProps) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
      </div>
    );
  }

  if (changes.length === 0) {
    return (
      <div className="text-center py-12 text-gray-400">
        No change requests found
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead className="bg-gray-800 text-gray-300">
          <tr>
            <th className="px-4 py-3 text-left font-medium">File</th>
            <th className="px-4 py-3 text-left font-medium">Status</th>
            <th className="px-4 py-3 text-left font-medium">Mode</th>
            <th className="px-4 py-3 text-left font-medium">Created</th>
            {showActions && (
              <th className="px-4 py-3 text-right font-medium">Actions</th>
            )}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-700">
          {changes.map((change) => (
            <ChangeRequestRow
              key={change.id}
              change={change}
              onApprove={onApprove}
              onReject={onReject}
              showActions={showActions}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
});
