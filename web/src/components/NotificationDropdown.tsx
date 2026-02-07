'use client';

import React, { useState } from 'react';
import { useNotifications } from '@/contexts/NotificationContext';
import { Notification, NotificationType } from '@/domain/notification';
import { X, Clock, User, FileText, Shield, AlertTriangle, History, CheckCircle } from 'lucide-react';
import { useRouter } from 'next/navigation';

interface NotificationDropdownProps {
  onClose: () => void;
  onViewRuleHistory?: (ruleId: string) => void;
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

const notificationTypeLabels: Record<NotificationType, string> = {
  change_detected: 'Change Detected',
  approval_required: 'Approval Required',
  change_approved: 'Change Approved',
  change_rejected: 'Change Rejected',
  change_auto_reverted: 'Auto-Reverted',
  exception_granted: 'Exception Granted',
  exception_denied: 'Exception Denied',
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

const notificationTypeBgColors: Record<NotificationType, string> = {
  change_detected: 'bg-blue-500/10 border-blue-500/20',
  approval_required: 'bg-yellow-500/10 border-yellow-500/20',
  change_approved: 'bg-green-500/10 border-green-500/20',
  change_rejected: 'bg-red-500/10 border-red-500/20',
  change_auto_reverted: 'bg-orange-500/10 border-orange-500/20',
  exception_granted: 'bg-green-500/10 border-green-500/20',
  exception_denied: 'bg-red-500/10 border-red-500/20',
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

function formatFullDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

interface NotificationDetailDialogProps {
  notification: Notification;
  onClose: () => void;
  onMarkRead: (id: string) => void;
  onViewRuleHistory?: (ruleId: string) => void;
}

function NotificationDetailDialog({ notification, onClose, onMarkRead, onViewRuleHistory }: NotificationDetailDialogProps) {
  const [isMarkedRead, setIsMarkedRead] = useState(!!notification.read_at);
  const isUnread = !isMarkedRead;
  const metadata = (notification.metadata || {}) as Record<string, string | number | boolean | undefined>;
  const ruleId = metadata.rule_id as string | undefined;
  const isApprovalRequired = notification.type === 'approval_required';
  const router = useRouter();

  const handleMarkRead = () => {
    if (isUnread) {
      onMarkRead(notification.id);
      setIsMarkedRead(true);
      onClose();
    }
  };

  const handleViewRuleHistory = () => {
    if (ruleId && onViewRuleHistory) {
      if (isUnread) {
        onMarkRead(notification.id);
      }
      onViewRuleHistory(ruleId);
      onClose();
    }
  };

  const handleGoToApprovals = () => {
    if (isUnread) {
      onMarkRead(notification.id);
    }
    onClose();
    router.push('/approvals');
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-[100] p-4" onClick={onClose}>
      <div
        className="bg-card rounded-2xl shadow-2xl w-full max-w-2xl border border-border/50 overflow-hidden animate-scale-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header with gradient accent */}
        <div className={`relative p-6 border-b ${notificationTypeBgColors[notification.type]}`}>
          {/* Gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-transparent to-card/20" />

          <div className="relative flex items-start justify-between">
            <div className="flex items-center gap-4">
              <div className={`w-14 h-14 rounded-xl flex items-center justify-center text-3xl ${notificationTypeBgColors[notification.type]} border`}>
                {notificationTypeIcons[notification.type] || '📢'}
              </div>
              <div>
                <div className={`text-sm font-semibold uppercase tracking-wide ${notificationTypeColors[notification.type]}`}>
                  {notificationTypeLabels[notification.type]}
                </div>
                <h2 className="text-xl font-bold text-foreground mt-1">{notification.title}</h2>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-xl hover:bg-muted/50 transition-colors"
              aria-label="Close"
            >
              <X className="w-5 h-5 text-muted-foreground" />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="p-6 space-y-6">
          <p className="text-foreground text-base leading-relaxed">{notification.body}</p>

          {/* Metadata */}
          {Object.keys(metadata).length > 0 && (
            <div className="bg-muted/20 rounded-xl p-5 space-y-4 border border-border/50">
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Details</h4>
              <div className="grid gap-3">
                {metadata.project && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <FileText className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Project:</span>
                    <span className="font-medium text-foreground">{String(metadata.project)}</span>
                  </div>
                )}
                {metadata.file_path && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <FileText className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">File:</span>
                    <code className="px-2 py-1 bg-muted rounded-lg text-xs font-mono text-foreground">
                      {String(metadata.file_path)}
                    </code>
                  </div>
                )}
                {(metadata.agent || metadata.requester) && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <User className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">
                      {metadata.agent ? 'Agent:' : 'Requested by:'}
                    </span>
                    <span className="font-medium text-foreground">{String(metadata.agent || metadata.requester)}</span>
                  </div>
                )}
                {metadata.approver && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <User className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Approved by:</span>
                    <span className="font-medium text-foreground">{String(metadata.approver)}</span>
                  </div>
                )}
                {metadata.reviewer && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <User className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Reviewed by:</span>
                    <span className="font-medium text-foreground">{String(metadata.reviewer)}</span>
                  </div>
                )}
                {metadata.rule_id && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <Shield className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Rule ID:</span>
                    <code className="px-2 py-1 bg-muted rounded-lg text-xs font-mono text-foreground">
                      {String(metadata.rule_id).slice(0, 8)}...
                    </code>
                  </div>
                )}
                {metadata.reason && (
                  <div className="flex items-start gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center flex-shrink-0">
                      <AlertTriangle className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Reason:</span>
                    <span className="italic text-foreground">{String(metadata.reason)}</span>
                  </div>
                )}
                {metadata.expires_at && (
                  <div className="flex items-center gap-3 text-sm">
                    <div className="w-8 h-8 rounded-lg bg-muted/50 flex items-center justify-center">
                      <Clock className="w-4 h-4 text-muted-foreground" />
                    </div>
                    <span className="text-muted-foreground">Expires:</span>
                    <span className="font-medium text-foreground">{String(metadata.expires_at)}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Timestamp */}
          <div className="flex items-center gap-2 text-sm text-muted-foreground pt-2 border-t border-border/50">
            <Clock className="w-4 h-4" />
            <span>{formatFullDate(notification.created_at)}</span>
            {notification.read_at && (
              <span className="text-xs bg-muted/50 px-2 py-0.5 rounded-full">Read {formatTimeAgo(notification.read_at)}</span>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-muted/10 flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap gap-3">
            {ruleId && onViewRuleHistory && (
              <button
                onClick={handleViewRuleHistory}
                className="px-5 py-2.5 text-sm font-medium rounded-xl border-2 border-primary text-primary hover:bg-primary hover:text-primary-foreground transition-all flex items-center gap-2"
              >
                <History className="w-4 h-4" />
                View Rule History
              </button>
            )}
            {isApprovalRequired && (
              <button
                onClick={handleGoToApprovals}
                className="px-5 py-2.5 text-sm font-medium rounded-xl bg-gradient-to-r from-yellow-500 to-amber-500 text-white hover:from-yellow-600 hover:to-amber-600 transition-all shadow-lg shadow-yellow-500/25 flex items-center gap-2"
              >
                <CheckCircle className="w-4 h-4" />
                Review Approval
              </button>
            )}
          </div>
          <div className="flex gap-3">
            {isUnread && (
              <button
                onClick={handleMarkRead}
                className="px-5 py-2.5 text-sm font-medium rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 transition-all shadow-lg shadow-primary/25"
              >
                Mark as Read
              </button>
            )}
            <button
              onClick={onClose}
              className="px-5 py-2.5 text-sm font-medium rounded-xl border border-border hover:bg-muted transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function NotificationItem({
  notification,
  onMarkRead,
  onViewDetails,
}: {
  notification: Notification;
  onMarkRead: (id: string) => void;
  onViewDetails: (notification: Notification) => void;
}) {
  const isUnread = !notification.read_at;

  const handleClick = () => {
    onViewDetails(notification);
    if (isUnread) {
      onMarkRead(notification.id);
    }
  };

  return (
    <div
      className={`p-3 hover:bg-muted/50 cursor-pointer transition-colors ${isUnread ? 'bg-primary/5' : ''}`}
      onClick={handleClick}
    >
      <div className="flex items-start gap-3">
        <span className="text-lg">{notificationTypeIcons[notification.type] || '📢'}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`font-medium ${isUnread ? 'text-foreground' : 'text-muted-foreground'}`}>
              {notification.title}
            </span>
            {isUnread && <span className="w-2 h-2 bg-primary rounded-full flex-shrink-0" />}
          </div>
          <p className="text-sm text-muted-foreground truncate">{notification.body}</p>
          <p className={`text-xs ${notificationTypeColors[notification.type]} mt-1`}>
            {formatTimeAgo(notification.created_at)}
          </p>
        </div>
      </div>
    </div>
  );
}

export function NotificationDropdown({ onClose, onViewRuleHistory }: NotificationDropdownProps) {
  const { notifications, loading, markRead, markAllRead, unreadCount } = useNotifications();
  const [selectedNotification, setSelectedNotification] = useState<Notification | null>(null);

  const handleMarkAllRead = async () => {
    await markAllRead();
  };

  const handleViewDetails = (notification: Notification) => {
    setSelectedNotification(notification);
  };

  const handleCloseDetails = () => {
    setSelectedNotification(null);
  };

  return (
    <>
      <div className="absolute right-0 mt-2 w-80 bg-card rounded-lg shadow-lg border z-50">
        <div className="p-3 border-b flex justify-between items-center">
          <h3 className="font-semibold">Notifications</h3>
          {unreadCount > 0 && (
            <button
              onClick={handleMarkAllRead}
              className="text-sm text-primary hover:text-primary/80 transition-colors"
            >
              Mark all read
            </button>
          )}
        </div>

        <div className="max-h-96 overflow-y-auto">
          {loading ? (
            <div className="p-4 text-center text-muted-foreground">Loading...</div>
          ) : !notifications || notifications.length === 0 ? (
            <div className="p-4 text-center text-muted-foreground">No notifications</div>
          ) : (
            <div className="divide-y">
              {notifications.slice(0, 10).map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onMarkRead={markRead}
                  onViewDetails={handleViewDetails}
                />
              ))}
            </div>
          )}
        </div>

        {notifications && notifications.length > 10 && (
          <div className="p-3 border-t text-center">
            <button
              className="text-sm text-primary hover:text-primary/80 transition-colors"
              onClick={onClose}
            >
              View all notifications
            </button>
          </div>
        )}
      </div>

      {selectedNotification && (
        <NotificationDetailDialog
          notification={selectedNotification}
          onClose={handleCloseDetails}
          onMarkRead={markRead}
          onViewRuleHistory={onViewRuleHistory}
        />
      )}
    </>
  );
}
