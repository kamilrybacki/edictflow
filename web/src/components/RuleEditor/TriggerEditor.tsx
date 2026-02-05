'use client';

import { Trigger, TriggerType } from '@/domain/rule';
import { triggerTypes } from './constants';

interface TriggerEditorProps {
  triggers: Trigger[];
  onTriggersChange: (triggers: Trigger[]) => void;
}

export function TriggerEditor({ triggers, onTriggersChange }: TriggerEditorProps) {
  const handleAddTrigger = () => {
    onTriggersChange([...triggers, { type: 'path', pattern: '' }]);
  };

  const handleRemoveTrigger = (index: number) => {
    onTriggersChange(triggers.filter((_, i) => i !== index));
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
    onTriggersChange(newTriggers);
  };

  return (
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
  );
}
