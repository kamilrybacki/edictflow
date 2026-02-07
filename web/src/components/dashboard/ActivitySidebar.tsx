'use client';

import { useState } from 'react';
import { ChevronLeft, ChevronRight, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ActivityFeed, Activity } from './ActivityFeed';

interface ActivitySidebarProps {
  activities: Activity[];
}

export function ActivitySidebar({ activities }: ActivitySidebarProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <>
      {/* Toggle Button - always visible */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={cn(
          'fixed right-0 top-1/2 -translate-y-1/2 z-40',
          'flex items-center gap-1 px-2 py-3 rounded-l-lg',
          'bg-card border border-r-0 shadow-lg',
          'hover:bg-muted transition-colors',
          isExpanded && 'right-80'
        )}
        title={isExpanded ? 'Hide activity' : 'Show activity'}
      >
        <Clock className="w-4 h-4 text-muted-foreground" />
        {isExpanded ? (
          <ChevronRight className="w-4 h-4 text-muted-foreground" />
        ) : (
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        )}
      </button>

      {/* Sidebar Panel */}
      <div
        className={cn(
          'fixed right-0 top-0 h-full w-80 z-30',
          'bg-card border-l shadow-xl',
          'transform transition-transform duration-300 ease-in-out',
          isExpanded ? 'translate-x-0' : 'translate-x-full'
        )}
      >
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b">
            <div className="flex items-center gap-2">
              <Clock className="w-5 h-5 text-primary" />
              <h2 className="text-lg font-semibold">Recent Activity</h2>
            </div>
            <button
              onClick={() => setIsExpanded(false)}
              className="p-1 rounded-md hover:bg-muted transition-colors"
            >
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>

          {/* Activity List */}
          <div className="flex-1 overflow-y-auto p-4">
            <ActivityFeed activities={activities} />
          </div>

          {/* Footer */}
          <div className="p-4 border-t">
            <button className="w-full text-sm text-primary hover:underline">
              View all activity
            </button>
          </div>
        </div>
      </div>

      {/* Backdrop */}
      {isExpanded && (
        <div
          className="fixed inset-0 bg-black/20 z-20 lg:hidden"
          onClick={() => setIsExpanded(false)}
        />
      )}
    </>
  );
}
