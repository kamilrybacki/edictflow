'use client';

import { useState, useEffect } from 'react';
import { Rule, TargetLayer, EnforcementMode } from '@/domain/rule';
import { useRuleForm } from '@/hooks/useRuleForm';
import { useDebounce } from '@/hooks/useDebounce';
import { useAuth } from '@/contexts/AuthContext';
import { ModalErrorBoundary } from '@/components/ErrorBoundary';
import { createGlobalRule } from '@/lib/api';
import { targetLayers } from './constants';
import { TriggerEditor } from './TriggerEditor';
import { EnforcementSection } from './EnforcementSection';

interface RuleEditorProps {
  teamId: string;
  rule?: Rule;
  onSave: () => void;
  onCancel: () => void;
}

function RuleEditorContent({ teamId, rule, onSave, onCancel }: RuleEditorProps) {
  const { hasPermission } = useAuth();
  const isAdmin = hasPermission('admin_access');

  // State for global rules
  const [scope, setScope] = useState<'team' | 'global'>('team');
  const [force, setForce] = useState(false);
  const [savingGlobal, setSavingGlobal] = useState(false);
  const [globalError, setGlobalError] = useState<string | null>(null);

  const {
    formData,
    errors,
    categories,
    saving,
    setField,
    setTriggers,
    handleSubmit,
  } = useRuleForm({
    teamId,
    rule,
    onSuccess: onSave,
  });

  // Debounce content field to prevent excessive re-renders during typing
  const debouncedContent = useDebounce(formData.content, 300);

  // Character count uses debounced value for display
  const contentLength = debouncedContent.length;

  // When scope changes to global, auto-set targetLayer to enterprise
  useEffect(() => {
    if (scope === 'global') {
      setField('targetLayer', 'enterprise');
    }
  }, [scope, setField]);

  // Handle form submission
  const onFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setGlobalError(null);

    if (scope === 'global') {
      // Validate required fields
      if (!formData.name.trim()) {
        setGlobalError('Name is required');
        return;
      }
      if (!formData.content.trim()) {
        setGlobalError('Content is required');
        return;
      }

      try {
        setSavingGlobal(true);
        await createGlobalRule({
          name: formData.name.trim(),
          content: formData.content.trim(),
          description: formData.description.trim() || undefined,
          force,
        });
        onSave();
      } catch (err) {
        setGlobalError(err instanceof Error ? err.message : 'Failed to create global rule');
      } finally {
        setSavingGlobal(false);
      }
    } else {
      handleSubmit(e);
    }
  };

  const isSaving = saving || savingGlobal;
  const displayError = globalError || errors.general || errors.name || errors.content;

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

        <form onSubmit={onFormSubmit} className="p-4 overflow-y-auto max-h-[calc(90vh-130px)]">
          {displayError && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-400 text-sm">
              {displayError}
            </div>
          )}

          <div className="space-y-4">
            {/* Scope selector - visible to admins only */}
            {isAdmin && !rule && (
              <div className="space-y-2">
                <label className="block text-sm font-medium">Scope</label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="radio"
                      name="scope"
                      value="team"
                      checked={scope === 'team'}
                      onChange={() => setScope('team')}
                    />
                    <span>Team</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="radio"
                      name="scope"
                      value="global"
                      checked={scope === 'global'}
                      onChange={() => setScope('global')}
                    />
                    <span>Global (Organization-wide)</span>
                  </label>
                </div>
              </div>
            )}

            {/* Force checkbox - visible when scope is global */}
            {scope === 'global' && (
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={force}
                    onChange={(e) => setForce(e.target.checked)}
                  />
                  <span className="font-medium">Force on all teams</span>
                </label>
                <p className="text-sm text-zinc-500 ml-6">
                  Forced rules apply even to teams that opted out of global rule inheritance
                </p>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setField('name', e.target.value)}
                className={`w-full px-3 py-2 border rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.name ? 'border-red-500' : 'border-zinc-300 dark:border-zinc-700'
                }`}
                placeholder="e.g., TypeScript Best Practices"
              />
              {errors.name && (
                <p className="mt-1 text-xs text-red-500">{errors.name}</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Description (optional)</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setField('description', e.target.value)}
                className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Brief explanation for admins/users"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              {/* Target Layer - hidden when scope is global */}
              {scope !== 'global' && (
                <div>
                  <label className="block text-sm font-medium mb-1">Target Layer</label>
                  <select
                    value={formData.targetLayer}
                    onChange={(e) => setField('targetLayer', e.target.value as TargetLayer)}
                    className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {targetLayers.map((layer) => (
                      <option key={layer.value} value={layer.value}>
                        {layer.label}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <div className={scope === 'global' ? 'col-span-2' : ''}>
                <label className="block text-sm font-medium mb-1">Category</label>
                <select
                  value={formData.categoryId}
                  onChange={(e) => setField('categoryId', e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select category...</option>
                  {categories.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name} {cat.isSystem ? '(System)' : ''}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Priority Weight</label>
                <input
                  type="number"
                  value={formData.priorityWeight}
                  onChange={(e) => setField('priorityWeight', parseInt(e.target.value) || 0)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="0"
                />
              </div>

              <div className="flex items-center pt-6">
                <input
                  type="checkbox"
                  id="overridable"
                  checked={formData.overridable}
                  onChange={(e) => setField('overridable', e.target.checked)}
                  className="w-4 h-4 text-blue-600 bg-white border-zinc-300 rounded focus:ring-blue-500"
                />
                <label htmlFor="overridable" className="ml-2 text-sm font-medium">
                  Overridable by lower layers
                </label>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Effective Start (optional)</label>
                <input
                  type="datetime-local"
                  value={formData.effectiveStart}
                  onChange={(e) => setField('effectiveStart', e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Effective End (optional)</label>
                <input
                  type="datetime-local"
                  value={formData.effectiveEnd}
                  onChange={(e) => setField('effectiveEnd', e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Tags (optional, comma-separated)</label>
              <input
                type="text"
                value={formData.tags}
                onChange={(e) => setField('tags', e.target.value)}
                className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="e.g., security, best-practices, typescript"
              />
            </div>

            <div>
              <div className="flex justify-between items-center mb-1">
                <label className="block text-sm font-medium">Content (Markdown)</label>
                <span className="text-xs text-zinc-500">{contentLength} characters</span>
              </div>
              <textarea
                value={formData.content}
                onChange={(e) => setField('content', e.target.value)}
                rows={8}
                className={`w-full px-3 py-2 border rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm ${
                  errors.content ? 'border-red-500' : 'border-zinc-300 dark:border-zinc-700'
                }`}
                placeholder="# Rule Title&#10;&#10;Your rule content in markdown..."
              />
              {errors.content && (
                <p className="mt-1 text-xs text-red-500">{errors.content}</p>
              )}
            </div>

            <EnforcementSection
              enforcementMode={formData.enforcementMode}
              temporaryTimeoutHours={formData.temporaryTimeoutHours}
              onEnforcementModeChange={(mode: EnforcementMode) => setField('enforcementMode', mode)}
              onTemporaryTimeoutChange={(hours: number) => setField('temporaryTimeoutHours', hours)}
            />

            <TriggerEditor
              triggers={formData.triggers}
              onTriggersChange={setTriggers}
            />
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
            onClick={onFormSubmit}
            disabled={isSaving}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : rule ? 'Update Rule' : scope === 'global' ? 'Create Global Rule' : 'Create Rule'}
          </button>
        </div>
      </div>
    </div>
  );
}

/**
 * RuleEditor component wrapped with error boundary for graceful error handling.
 */
export function RuleEditor(props: RuleEditorProps) {
  return (
    <ModalErrorBoundary onClose={props.onCancel}>
      <RuleEditorContent {...props} />
    </ModalErrorBoundary>
  );
}
