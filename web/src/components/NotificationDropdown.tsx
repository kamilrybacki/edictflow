'use client';

import React from 'react';
import { useNotifications } from '@/contexts/NotificationContext';
import { Notification, NotificationType } from '@/domain/notification';
import Link from 'next/link';

interface NotificationDropdownProps {
  onClose: () => void;
}

const notificationTypeIcons: Record<NotificationType, string> = {
  change_detected: '📝',
  approval_required: '⏳',
  change_approved: '✅',
  change_rejected: '❌',
  change_auto_reverted: '↩️',
  exception_granted: '🔓',
  exception_denied: '🔒',
};

const notificationTypeColors: Record<NotificationType, string> = {
  change_detected: 'text-blue-400',
  approval_required: 'text-yellow-400',
  change_approved: 'text-green-400',
  change_rejected: 'text-red-400',
  change_auto_reverted: 'text-orange-400',
  exception_granted: 'text-green-400',
  exception_denied: 'text-red-400',
};

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

function NotificationItem({
  notification,
  onMarkRead,
}: {
  notification: Notification;
  onMarkRead: (id: string) => void;
}) {
  const isUnread = !notification.read_at;
  const changeRequestId = notification.metadata?.change_request_id as string | undefined;

  const handleClick = () => {
    if (isUnread) {
      onMarkRead(notification.id);
    }
  };

  const content = (
    <div
      className={`p-3 hover:bg-gray-700 cursor-pointer ${isUnread ? 'bg-gray-750' : ''}`}
      onClick={handleClick}
    >
      <div className="flex items-start gap-3">
        <span className="text-lg">
          {notificationTypeIcons[notification.type] || '📢'}
        </span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`font-medium ${isUnread ? 'text-white' : 'text-gray-300'}`}>
              {notification.title}
            </span>
            {isUnread && (
              <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0" />
            )}
          </div>
          <p className="text-sm text-gray-400 truncate">{notification.body}</p>
          <p className={`text-xs ${notificationTypeColors[notification.type]} mt-1`}>
            {formatTimeAgo(notification.created_at)}
          </p>
        </div>
      </div>
    </div>
  );

  if (changeRequestId) {
    return (
      <Link href={`/changes/${changeRequestId}`}>
        {content}
      </Link>
    );
  }

  return content;
}

export function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const { notifications, loading, markRead, markAllRead, unreadCount } = useNotifications();

  const handleMarkAllRead = async () => {
    await markAllRead();
  };

  return (
    <div className="absolute right-0 mt-2 w-80 bg-gray-800 rounded-lg shadow-lg border border-gray-700 z-50">
      <div className="p-3 border-b border-gray-700 flex justify-between items-center">
        <h3 className="font-semibold text-white">Notifications</h3>
        {unreadCount > 0 && (
          <button
            onClick={handleMarkAllRead}
            className="text-sm text-blue-400 hover:text-blue-300"
          >
            Mark all read
          </button>
        )}
      </div>

      <div className="max-h-96 overflow-y-auto">
        {loading ? (
          <div className="p-4 text-center text-gray-400">Loading...</div>
        ) : notifications.length === 0 ? (
          <div className="p-4 text-center text-gray-400">No notifications</div>
        ) : (
          <div className="divide-y divide-gray-700">
            {notifications.slice(0, 10).map((notification) => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                onMarkRead={markRead}
              />
            ))}
          </div>
        )}
      </div>

      {notifications.length > 10 && (
        <div className="p-3 border-t border-gray-700 text-center">
          <Link
            href="/notifications"
            className="text-sm text-blue-400 hover:text-blue-300"
            onClick={onClose}
          >
            View all notifications
          </Link>
        </div>
      )}
    </div>
  );
}
