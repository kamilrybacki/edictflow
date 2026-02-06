'use client';

import { useEffect, useState } from 'react';
import { X, Wifi, Monitor, Clock, MapPin } from 'lucide-react';
import { ConnectedAgent, fetchConnectedAgents } from '@/lib/api/agents';

interface AgentListModalProps {
  isOpen: boolean;
  onClose: () => void;
}

function formatDuration(connectedAt?: string): string {
  if (!connectedAt) return 'Unknown';

  const connected = new Date(connectedAt);
  const now = new Date();
  const diffMs = now.getTime() - connected.getTime();

  const hours = Math.floor(diffMs / (1000 * 60 * 60));
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function formatRemoteAddr(addr?: string): string {
  if (!addr) return 'Unknown';
  // Remove port if present (e.g., "192.168.1.1:54321" -> "192.168.1.1")
  return addr.split(':')[0];
}

export function AgentListModal({ isOpen, onClose }: AgentListModalProps) {
  const [agents, setAgents] = useState<ConnectedAgent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    async function loadAgents() {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchConnectedAgents();
        setAgents(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load agents');
      } finally {
        setLoading(false);
      }
    }

    loadAgents();
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-card rounded-xl shadow-xl w-full max-w-3xl max-h-[80vh] overflow-hidden border">
        {/* Header */}
        <div className="p-4 border-b flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Wifi className="w-5 h-5 text-layer-user" />
            <h2 className="text-lg font-semibold">Connected Agents</h2>
            <span className="px-2 py-0.5 text-xs font-medium bg-layer-user/10 text-layer-user rounded-full">
              {agents.length} online
            </span>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg hover:bg-muted/50 transition-colors"
            aria-label="Close"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="overflow-y-auto max-h-[calc(80vh-130px)]">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-muted-foreground">Loading agents...</div>
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-status-rejected">{error}</div>
            </div>
          ) : agents.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <Wifi className="w-12 h-12 mb-4 opacity-50" />
              <p>No agents currently connected</p>
            </div>
          ) : (
            <div className="divide-y">
              {agents.map((agent) => (
                <div
                  key={agent.agent_id || agent.user_id}
                  className="p-4 hover:bg-muted/30 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className="p-2 rounded-lg bg-layer-user/10">
                        <Monitor className="w-5 h-5 text-layer-user" />
                      </div>
                      <div>
                        <div className="font-medium">
                          {agent.hostname || 'Unknown Host'}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Agent ID: {agent.agent_id || 'N/A'}
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      {agent.version && (
                        <span className="px-2 py-0.5 text-xs font-medium bg-muted rounded">
                          v{agent.version}
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="mt-3 grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <MapPin className="w-4 h-4" />
                      <span>{formatRemoteAddr(agent.remote_addr)}</span>
                    </div>
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Monitor className="w-4 h-4" />
                      <span>{agent.os || 'Unknown OS'}</span>
                    </div>
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Clock className="w-4 h-4" />
                      <span>{formatDuration(agent.connected_at)}</span>
                    </div>
                    <div className="text-muted-foreground">
                      Team: {agent.team_id || 'None'}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium rounded-lg hover:bg-muted/50 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
