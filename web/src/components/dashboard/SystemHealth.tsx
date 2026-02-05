'use client';

import { CheckCircle2, Wifi, AlertTriangle, XCircle } from 'lucide-react';

interface SystemHealthProps {
  syncStatus: 'synced' | 'syncing' | 'error';
  agentsOnline: number;
  pendingExceptions: number;
  errorsLast24h: number;
  lastUpdated?: string;
}

export function SystemHealth({
  syncStatus,
  agentsOnline,
  pendingExceptions,
  errorsLast24h,
  lastUpdated = '2 min ago'
}: SystemHealthProps) {
  return (
    <div className="bg-card rounded-xl border p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-heading">System Health</h2>
        <span className="text-xs text-muted-foreground">Updated {lastUpdated}</span>
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
          <CheckCircle2 className={`w-5 h-5 ${syncStatus === 'synced' ? 'text-status-approved' : syncStatus === 'error' ? 'text-status-rejected' : 'text-status-pending'}`} />
          <div>
            <p className="text-sm font-medium">Sync Status</p>
            <p className="text-caption">{syncStatus === 'synced' ? 'All synced' : syncStatus === 'syncing' ? 'Syncing...' : 'Sync error'}</p>
          </div>
        </div>
        <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
          <Wifi className="w-5 h-5 text-layer-user" />
          <div>
            <p className="text-sm font-medium">{agentsOnline} Agents</p>
            <p className="text-caption">Online</p>
          </div>
        </div>
        <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
          <AlertTriangle className="w-5 h-5 text-status-pending" />
          <div>
            <p className="text-sm font-medium">{pendingExceptions} Pending</p>
            <p className="text-caption">Exceptions</p>
          </div>
        </div>
        <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
          <XCircle className="w-5 h-5 text-muted-foreground" />
          <div>
            <p className="text-sm font-medium">{errorsLast24h} Errors</p>
            <p className="text-caption">Last 24h</p>
          </div>
        </div>
      </div>
    </div>
  );
}
