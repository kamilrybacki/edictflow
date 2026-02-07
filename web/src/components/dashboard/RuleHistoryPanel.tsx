'use client';

import { useState, useEffect } from 'react';
import { X, Clock, User, ArrowLeft, ExternalLink } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge, Button } from '@/components/ui';
import { fetchEntityHistory, AuditEntry } from '@/lib/api/audit';
import { Rule } from '@/domain/rule';
import { getLayerConfig } from '@/lib/layerConfig';
import { DiffDialog } from './DiffDialog';

interface RuleHistoryPanelProps {
  rule: Rule;
  onClose: () => void;
  onBack: () => void;
}

const actionLabels: Record<string, { label: string; className: string }> = {
  created: { label: 'Created', className: 'bg-green-500/10 text-green-600 border-green-500/30' },
  updated: { label: 'Updated', className: 'bg-blue-500/10 text-blue-600 border-blue-500/30' },
  deleted: { label: 'Deleted', className: 'bg-red-500/10 text-red-600 border-red-500/30' },
  submitted: { label: 'Submitted', className: 'bg-yellow-500/10 text-yellow-600 border-yellow-500/30' },
  approved: { label: 'Approved', className: 'bg-green-500/10 text-green-600 border-green-500/30' },
  rejected: { label: 'Rejected', className: 'bg-red-500/10 text-red-600 border-red-500/30' },
};

function HistoryEntry({ entry, onClick }: { entry: AuditEntry; onClick: () => void }) {
  const actionConfig = actionLabels[entry.action] || { label: entry.action, className: 'bg-muted text-muted-foreground' };
  const hasChanges = entry.changes && Object.keys(entry.changes).length > 0;
  const changeCount = hasChanges ? Object.keys(entry.changes!).length : 0;

  return (
    <button
      onClick={onClick}
      className="w-full text-left border-l-2 border-muted pl-4 pb-4 last:pb-0 relative hover:bg-muted/50 rounded-r-lg transition-colors cursor-pointer group"
    >
      {/* Timeline dot */}
      <div className="absolute -left-[5px] top-0 w-2 h-2 rounded-full bg-muted-foreground group-hover:bg-primary transition-colors" />

      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <Badge variant="outline" className={cn('text-xs', actionConfig.className)}>
              {actionConfig.label}
            </Badge>
            {hasChanges && (
              <span className="text-xs text-muted-foreground">
                {changeCount} field{changeCount !== 1 ? 's' : ''} changed
              </span>
            )}
          </div>
          <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {new Date(entry.created_at).toLocaleString()}
            </span>
            <span className="flex items-center gap-1">
              <User className="w-3 h-3" />
              {entry.actor_name || 'System'}
            </span>
          </div>
        </div>

        <ExternalLink className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" />
      </div>
    </button>
  );
}

export function RuleHistoryPanel({ rule, onClose, onBack }: RuleHistoryPanelProps) {
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedEntry, setSelectedEntry] = useState<AuditEntry | null>(null);

  const layer = getLayerConfig(rule.targetLayer);
  const LayerIcon = layer.icon;

  useEffect(() => {
    async function loadHistory() {
      try {
        setLoading(true);
        setError(null);
        const history = await fetchEntityHistory('rule', rule.id);
        setEntries(history);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load history');
      } finally {
        setLoading(false);
      }
    }

    loadHistory();
  }, [rule.id]);

  return (
    <div className="bg-card rounded-xl border h-full flex flex-col">
      {/* Header */}
      <div className={cn('p-4 rounded-t-xl', layer.bgClassName)}>
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <button
              onClick={onBack}
              className="p-1.5 rounded-md hover:bg-white/10 transition-colors"
              title="Back to details"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <div className={cn('w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 shadow-medium', layer.className)}>
              <LayerIcon className="w-5 h-5" />
            </div>
            <div className="min-w-0">
              <h2 className="font-semibold text-lg truncate">{rule.name}</h2>
              <p className="text-sm text-muted-foreground">History</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md hover:bg-muted transition-colors"
          >
            <X className="w-5 h-5 text-muted-foreground" />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="flex items-center justify-center py-8 text-muted-foreground">
            Loading history...
          </div>
        ) : error ? (
          <div className="p-4 rounded-lg bg-destructive/10 text-destructive text-sm">
            {error}
          </div>
        ) : entries.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <Clock className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>No history available for this rule.</p>
            <p className="text-xs mt-1">Changes will appear here once audit logging is enabled.</p>
          </div>
        ) : (
          <div className="space-y-0">
            {entries.map((entry) => (
              <HistoryEntry
                key={entry.id}
                entry={entry}
                onClick={() => setSelectedEntry(entry)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="p-4 border-t">
        <Button
          size="sm"
          variant="outline"
          onClick={onBack}
          className="gap-1.5"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Details
        </Button>
      </div>

      {/* Diff Dialog */}
      {selectedEntry && (
        <DiffDialog
          entry={selectedEntry}
          onClose={() => setSelectedEntry(null)}
        />
      )}
    </div>
  );
}
