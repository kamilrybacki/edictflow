'use client';

import { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  variant?: 'default' | 'organization' | 'team' | 'project' | 'warning';
  onClick?: () => void;
}

const variantClasses = {
  default: 'bg-card border',
  organization: 'bg-layer-organization/10 border-layer-organization/20',
  team: 'bg-layer-team/10 border-layer-team/20',
  project: 'bg-layer-project/10 border-layer-project/20',
  warning: 'bg-status-pending/10 border-status-pending/20',
};

export function StatCard({ title, value, icon, trend, variant = 'default', onClick }: StatCardProps) {
  return (
    <div
      className={cn(
        'p-4 rounded-xl border transition-all duration-200 hover:shadow-medium',
        variantClasses[variant],
        onClick && 'cursor-pointer hover:ring-2 hover:ring-primary/20'
      )}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => e.key === 'Enter' && onClick() : undefined}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-caption">{title}</p>
          <p className="text-2xl font-bold mt-1">{value}</p>
        </div>
        <div className="p-2 rounded-lg bg-muted/50">
          {icon}
        </div>
      </div>
      {trend && (
        <div className={cn(
          'mt-2 text-xs flex items-center gap-1',
          trend.isPositive ? 'text-status-approved' : 'text-status-rejected'
        )}>
          <span>{trend.isPositive ? '↑' : '↓'}</span>
          <span>{Math.abs(trend.value)}%</span>
          <span className="text-muted-foreground">vs last week</span>
        </div>
      )}
    </div>
  );
}
