'use client';

import { useState } from 'react';
import { Rule, TargetLayer, TriggerType, Trigger } from '@/domain/rule';
import { createRule, CreateRuleRequest } from '@/lib/api';

interface RuleEditorProps {
  teamId: string;
  rule?: Rule;
  onSave: () => void;
  onCancel: () => void;
}

const targetLayers: TargetLayer[] = ['enterprise', 'global', 'project', 'local'];
const triggerTypes: TriggerType[] = ['path', 'context', 'tag'];

export function RuleEditor({ teamId, rule, onSave, onCancel }: RuleEditorProps) {
  const [name, setName] = useState(rule?.name || '');
  const [content, setContent] = useState(rule?.content || '');
  const [targetLayer, setTargetLayer] = useState<TargetLayer>(rule?.targetLayer || 'project');
  const [priorityWeight, setPriorityWeight] = useState(rule?.priorityWeight || 0);
  const [triggers, setTriggers] = useState<Trigger[]>(rule?.triggers || []);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAddTrigger = () => {
    setTriggers([...triggers, { type: 'path', pattern: '' }]);
  };

  const handleRemoveTrigger = (index: number) => {
    setTriggers(triggers.filter((_, i) => i !== index));
  };

  const handleTriggerChange = (index: number, field: keyof Trigger, value: string | string[]) => {
    const newTriggers = [...triggers];
    const trigger = { ...newTriggers[index] };

    if (field === 'type') {
      trigger.type = value as TriggerType;
      // Reset other fields when type changes
      trigger.pattern = undefined;
      trigger.contextTypes = undefined;
      trigger.tags = undefined;
    } else if (field === 'pattern') {
      trigger.pattern = value as string;
    } else if (field === 'contextTypes') {
      trigger.contextTypes = (value as string).split(',').map((s) => s.trim()).filter(Boolean);
    } else if (field === 'tags') {
      trigger.tags = (value as string).split(',').map((s) => s.trim()).filter(Boolean);
    }

    newTriggers[index] = trigger;
    setTriggers(newTriggers);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    if (!content.trim()) {
      setError('Content is required');
      return;
    }

    try {
      setSaving(true);
      setError(null);

      const request: CreateRuleRequest = {
        name: name.trim(),
        content: content.trim(),
        target_layer: targetLayer,
        team_id: teamId,
        triggers: triggers.map((t) => ({
          type: t.type,
          pattern: t.pattern,
          context_types: t.contextTypes,
          tags: t.tags,
        })),
      };

      await createRule(request);
      onSave();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save rule');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-zinc-900 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden">
        <div className="p-4 border-b border-zinc-200 dark:border-zinc-700 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{rule ? 'Edit Rule' : 'Create New Rule'}</h2>
          <button onClick={onCancel} className="text-zinc-400 hover:text-zinc-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 overflow-y-auto max-h-[calc(90vh-130px)]">
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-400 text-sm">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="e.g., TypeScript Best Practices"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Target Layer</label>
                <select
                  value={targetLayer}
                  onChange={(e) => setTargetLayer(e.target.value as TargetLayer)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {targetLayers.map((layer) => (
                    <option key={layer} value={layer}>
                      {layer.charAt(0).toUpperCase() + layer.slice(1)}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Priority Weight</label>
                <input
                  type="number"
                  value={priorityWeight}
                  onChange={(e) => setPriorityWeight(parseInt(e.target.value) || 0)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="0"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Content (Markdown)</label>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={8}
                className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                placeholder="# Rule Title&#10;&#10;Your rule content in markdown..."
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="block text-sm font-medium">Triggers</label>
                <button
                  type="button"
                  onClick={handleAddTrigger}
                  className="text-sm text-blue-600 hover:underline"
                >
                  + Add Trigger
                </button>
              </div>

              {triggers.length === 0 ? (
                <p className="text-sm text-zinc-500">No triggers. Rule will always apply.</p>
              ) : (
                <div className="space-y-3">
                  {triggers.map((trigger, index) => (
                    <div
                      key={index}
                      className="flex gap-2 p-3 bg-zinc-50 dark:bg-zinc-800 rounded-md"
                    >
                      <select
                        value={trigger.type}
                        onChange={(e) => handleTriggerChange(index, 'type', e.target.value)}
                        className="px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700"
                      >
                        {triggerTypes.map((type) => (
                          <option key={type} value={type}>
                            {type}
                          </option>
                        ))}
                      </select>

                      {trigger.type === 'path' && (
                        <input
                          type="text"
                          value={trigger.pattern || ''}
                          onChange={(e) => handleTriggerChange(index, 'pattern', e.target.value)}
                          placeholder="e.g., *.tsx, src/**/*.ts"
                          className="flex-1 px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700"
                        />
                      )}

                      {trigger.type === 'context' && (
                        <input
                          type="text"
                          value={trigger.contextTypes?.join(', ') || ''}
                          onChange={(e) => handleTriggerChange(index, 'contextTypes', e.target.value)}
                          placeholder="e.g., frontend, debug"
                          className="flex-1 px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700"
                        />
                      )}

                      {trigger.type === 'tag' && (
                        <input
                          type="text"
                          value={trigger.tags?.join(', ') || ''}
                          onChange={(e) => handleTriggerChange(index, 'tags', e.target.value)}
                          placeholder="e.g., react, typescript"
                          className="flex-1 px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700"
                        />
                      )}

                      <button
                        type="button"
                        onClick={() => handleRemoveTrigger(index)}
                        className="p-1 text-zinc-400 hover:text-red-500"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </form>

        <div className="p-4 border-t border-zinc-200 dark:border-zinc-700 flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 text-sm font-medium text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-md"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={saving}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? 'Saving...' : rule ? 'Update Rule' : 'Create Rule'}
          </button>
        </div>
      </div>
    </div>
  );
}
