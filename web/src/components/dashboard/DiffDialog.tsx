'use client';

import { X, Clock, User, ArrowRight, Plus, Minus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui';
import { AuditEntry } from '@/lib/api/audit';

interface DiffDialogProps {
  entry: AuditEntry;
  onClose: () => void;
}

const actionLabels: Record<string, { label: string; className: string }> = {
  created: { label: 'Created', className: 'bg-green-500/10 text-green-600 border-green-500/30' },
  updated: { label: 'Updated', className: 'bg-blue-500/10 text-blue-600 border-blue-500/30' },
  deleted: { label: 'Deleted', className: 'bg-red-500/10 text-red-600 border-red-500/30' },
  submitted: { label: 'Submitted', className: 'bg-yellow-500/10 text-yellow-600 border-yellow-500/30' },
  approved: { label: 'Approved', className: 'bg-green-500/10 text-green-600 border-green-500/30' },
  rejected: { label: 'Rejected', className: 'bg-red-500/10 text-red-600 border-red-500/30' },
};

function DiffLine({ type, content }: { type: 'add' | 'remove' | 'context'; content: string }) {
  return (
    <div
      className={cn(
        'font-mono text-xs px-2 py-0.5 flex items-start gap-2',
        type === 'add' && 'bg-green-500/10 text-green-600',
        type === 'remove' && 'bg-red-500/10 text-red-600',
        type === 'context' && 'text-muted-foreground'
      )}
    >
      <span className="w-4 flex-shrink-0 text-center">
        {type === 'add' && <Plus className="w-3 h-3 inline" />}
        {type === 'remove' && <Minus className="w-3 h-3 inline" />}
      </span>
      <span className="whitespace-pre-wrap break-all">{content || '(empty)'}</span>
    </div>
  );
}

function computeLineDiff(oldValue: string, newValue: string): Array<{ type: 'add' | 'remove' | 'context'; content: string }> {
  const oldLines = oldValue.split('\n');
  const newLines = newValue.split('\n');
  const result: Array<{ type: 'add' | 'remove' | 'context'; content: string }> = [];

  // Simple line-by-line diff using longest common subsequence approach
  const lcs = computeLCS(oldLines, newLines);

  let oldIdx = 0;
  let newIdx = 0;
  let lcsIdx = 0;

  while (oldIdx < oldLines.length || newIdx < newLines.length) {
    if (lcsIdx < lcs.length && oldIdx < oldLines.length && oldLines[oldIdx] === lcs[lcsIdx]) {
      if (newIdx < newLines.length && newLines[newIdx] === lcs[lcsIdx]) {
        // Both match LCS - context line
        result.push({ type: 'context', content: oldLines[oldIdx] });
        oldIdx++;
        newIdx++;
        lcsIdx++;
      } else {
        // New line added
        result.push({ type: 'add', content: newLines[newIdx] });
        newIdx++;
      }
    } else if (lcsIdx < lcs.length && newIdx < newLines.length && newLines[newIdx] === lcs[lcsIdx]) {
      // Old line removed
      result.push({ type: 'remove', content: oldLines[oldIdx] });
      oldIdx++;
    } else if (oldIdx < oldLines.length) {
      // Old line removed
      result.push({ type: 'remove', content: oldLines[oldIdx] });
      oldIdx++;
    } else if (newIdx < newLines.length) {
      // New line added
      result.push({ type: 'add', content: newLines[newIdx] });
      newIdx++;
    }
  }

  return result;
}

function computeLCS(a: string[], b: string[]): string[] {
  const m = a.length;
  const n = b.length;
  const dp: number[][] = Array(m + 1).fill(null).map(() => Array(n + 1).fill(0));

  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      if (a[i - 1] === b[j - 1]) {
        dp[i][j] = dp[i - 1][j - 1] + 1;
      } else {
        dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
      }
    }
  }

  // Backtrack to find LCS
  const result: string[] = [];
  let i = m;
  let j = n;
  while (i > 0 && j > 0) {
    if (a[i - 1] === b[j - 1]) {
      result.unshift(a[i - 1]);
      i--;
      j--;
    } else if (dp[i - 1][j] > dp[i][j - 1]) {
      i--;
    } else {
      j--;
    }
  }

  return result;
}

function FieldDiff({ field, oldValue, newValue }: { field: string; oldValue: unknown; newValue: unknown }) {
  const oldStr = oldValue !== null && oldValue !== undefined ? String(oldValue) : '';
  const newStr = newValue !== null && newValue !== undefined ? String(newValue) : '';

  const isMultiline = oldStr.includes('\n') || newStr.includes('\n');
  const diffLines = isMultiline ? computeLineDiff(oldStr, newStr) : null;

  return (
    <div className="border rounded-lg overflow-hidden">
      <div className="bg-muted px-3 py-2 border-b">
        <span className="font-medium text-sm">{field}</span>
      </div>

      {isMultiline && diffLines ? (
        <div className="max-h-64 overflow-y-auto">
          {diffLines.map((line, idx) => (
            <DiffLine key={idx} type={line.type} content={line.content} />
          ))}
        </div>
      ) : (
        <div className="p-3 space-y-2">
          <div className="flex items-start gap-2">
            <div className="flex items-center gap-1 text-xs text-red-500 font-medium w-16 flex-shrink-0">
              <Minus className="w-3 h-3" />
              Before
            </div>
            <div className="flex-1 font-mono text-sm bg-red-500/10 text-red-600 rounded px-2 py-1 break-all">
              {oldStr || '(empty)'}
            </div>
          </div>
          <div className="flex items-center justify-center">
            <ArrowRight className="w-4 h-4 text-muted-foreground" />
          </div>
          <div className="flex items-start gap-2">
            <div className="flex items-center gap-1 text-xs text-green-500 font-medium w-16 flex-shrink-0">
              <Plus className="w-3 h-3" />
              After
            </div>
            <div className="flex-1 font-mono text-sm bg-green-500/10 text-green-600 rounded px-2 py-1 break-all">
              {newStr || '(empty)'}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export function DiffDialog({ entry, onClose }: DiffDialogProps) {
  const actionConfig = actionLabels[entry.action] || { label: entry.action, className: 'bg-muted text-muted-foreground' };
  const hasChanges = entry.changes && Object.keys(entry.changes).length > 0;
  const hasMetadata = entry.metadata && Object.keys(entry.metadata).length > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Dialog */}
      <div className="relative bg-card rounded-xl border shadow-lg w-full max-w-2xl max-h-[80vh] flex flex-col m-4 animate-fade-in">
        {/* Header */}
        <div className="flex items-start justify-between gap-4 p-4 border-b">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Badge variant="outline" className={cn('text-sm', actionConfig.className)}>
                {actionConfig.label}
              </Badge>
              <span className="text-sm text-muted-foreground">Change Details</span>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                {new Date(entry.created_at).toLocaleString()}
              </span>
              <span className="flex items-center gap-1">
                <User className="w-4 h-4" />
                {entry.actor_name || 'System'}
              </span>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md hover:bg-muted transition-colors"
          >
            <X className="w-5 h-5 text-muted-foreground" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {hasChanges ? (
            <>
              <h3 className="text-sm font-medium text-muted-foreground">Changed Fields</h3>
              <div className="space-y-3">
                {Object.entries(entry.changes!).map(([field, change]) => (
                  <FieldDiff
                    key={field}
                    field={field}
                    oldValue={change.old}
                    newValue={change.new}
                  />
                ))}
              </div>
            </>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <p>No detailed changes recorded for this action.</p>
              {entry.action === 'created' && (
                <p className="text-xs mt-1">This was the initial creation of the rule.</p>
              )}
            </div>
          )}

          {hasMetadata && (
            <>
              <h3 className="text-sm font-medium text-muted-foreground mt-6">Metadata</h3>
              <div className="bg-muted/50 rounded-lg p-3">
                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                  {JSON.stringify(entry.metadata, null, 2)}
                </pre>
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
