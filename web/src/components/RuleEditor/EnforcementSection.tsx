'use client';

import { EnforcementMode } from '@/domain/rule';
import { enforcementModes, DEFAULT_TIMEOUT_HOURS, MIN_TIMEOUT_HOURS, MAX_TIMEOUT_HOURS } from './constants';

interface EnforcementSectionProps {
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
  onEnforcementModeChange: (mode: EnforcementMode) => void;
  onTemporaryTimeoutChange: (hours: number) => void;
}

export function EnforcementSection({
  enforcementMode,
  temporaryTimeoutHours,
  onEnforcementModeChange,
  onTemporaryTimeoutChange,
}: EnforcementSectionProps) {
  return (
    <div className="border-t border-zinc-200 dark:border-zinc-700 pt-4 mt-4">
      <h3 className="text-sm font-medium mb-3">Enforcement</h3>

      <div className="mb-4">
        <label className="block text-sm font-medium mb-1">Enforcement Mode</label>
        <select
          value={enforcementMode}
          onChange={(e) => onEnforcementModeChange(e.target.value as EnforcementMode)}
          className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {enforcementModes.map((mode) => (
            <option key={mode.value} value={mode.value}>
              {mode.label} - {mode.description}
            </option>
          ))}
        </select>
        <p className="mt-1 text-xs text-zinc-500">
          {enforcementModes.find((m) => m.value === enforcementMode)?.description}
        </p>
      </div>

      {enforcementMode === 'temporary' && (
        <div>
          <label className="block text-sm font-medium mb-1">Timeout (hours)</label>
          <input
            type="number"
            value={temporaryTimeoutHours}
            onChange={(e) => onTemporaryTimeoutChange(parseInt(e.target.value) || DEFAULT_TIMEOUT_HOURS)}
            min={MIN_TIMEOUT_HOURS}
            max={MAX_TIMEOUT_HOURS}
            className="w-32 px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <p className="mt-1 text-xs text-zinc-500">
            Changes will auto-revert after this many hours if not approved.
          </p>
        </div>
      )}
    </div>
  );
}
