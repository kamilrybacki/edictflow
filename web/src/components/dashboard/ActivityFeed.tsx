'use client';

import { memo, useMemo } from 'react';
import { FilePlus, FileCheck, FileWarning, AlertCircle, ShieldCheck, LucideIcon } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { cn } from '@/lib/utils';

export type ActivityType = 'rule_created' | 'rule_approved' | 'rule_rejected' | 'change_detected' | 'exception_granted';

export interface Activity {
  id: string;
  type: ActivityType;
  message: string;
  timestamp: Date | string;
  userName: string;
}

interface ActivityFeedProps {
  activities: Activity[];
}

const activityIcons: Record<ActivityType, LucideIcon> = {
  rule_created: FilePlus,
  rule_approved: FileCheck,
  rule_rejected: FileWarning,
  change_detected: AlertCircle,
  exception_granted: ShieldCheck,
};

const activityColors: Record<ActivityType, string> = {
  rule_created: 'text-layer-user bg-layer-user/10',
  rule_approved: 'text-status-approved bg-status-approved/10',
  rule_rejected: 'text-status-rejected bg-status-rejected/10',
  change_detected: 'text-status-pending bg-status-pending/10',
  exception_granted: 'text-layer-enterprise bg-layer-enterprise/10',
};

// Memoized activity item to prevent re-renders
const ActivityItem = memo(function ActivityItem({
  activity,
  index
}: {
  activity: Activity;
  index: number;
}) {
  const Icon = activityIcons[activity.type];
  const colorClass = activityColors[activity.type];
  const timestamp = useMemo(
    () => typeof activity.timestamp === 'string' ? new Date(activity.timestamp) : activity.timestamp,
    [activity.timestamp]
  );

  return (
    <div
      className="flex gap-3 p-3 rounded-lg bg-card border hover:border-primary/20 transition-colors animate-fade-in"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0', colorClass)}>
        <Icon className="w-4 h-4" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-foreground">{activity.message}</p>
        <div className="flex items-center gap-2 mt-1 text-caption">
          <span>{activity.userName}</span>
          <span>•</span>
          <span>{formatDistanceToNow(timestamp, { addSuffix: true })}</span>
        </div>
      </div>
    </div>
  );
});

export const ActivityFeed = memo(function ActivityFeed({ activities }: ActivityFeedProps) {
  return (
    <div className="space-y-3">
      {activities.map((activity, index) => (
        <ActivityItem key={activity.id} activity={activity} index={index} />
      ))}
    </div>
  );
});
