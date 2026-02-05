'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { fetchTeam, updateTeamSettings, fetchGlobalRules } from '@/lib/api';
import { Team } from '@/domain/team';
import { Rule } from '@/domain/rule';

export default function TeamSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const teamId = params.id as string;
  const [team, setTeam] = useState<Team | null>(null);
  const [forcedRulesCount, setForcedRulesCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const [teamData, globalRules] = await Promise.all([
          fetchTeam(teamId),
          fetchGlobalRules(),
        ]);
        setTeam(teamData);
        setForcedRulesCount(globalRules.filter((r: Rule) => r.force).length);
      } catch (err) {
        console.error('Failed to load team settings:', err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [teamId]);

  const handleToggleInherit = async () => {
    if (!team) return;
    setSaving(true);
    try {
      const updated = await updateTeamSettings(teamId, {
        inherit_global_rules: !team.settings.inheritGlobalRules,
      });
      setTeam(updated);
    } catch (err) {
      console.error('Failed to update settings:', err);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="animate-pulse">Loading team settings...</div>
      </div>
    );
  }

  if (!team) {
    return (
      <div className="p-6">
        <div className="text-red-500">Team not found</div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-2xl">
      <button
        onClick={() => router.back()}
        className="text-sm text-zinc-500 hover:text-zinc-700 mb-4"
      >
        ← Back
      </button>

      <h1 className="text-2xl font-bold mb-6">{team.name} Settings</h1>

      <div className="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow">
        <h2 className="text-lg font-semibold mb-4">Global Rules Inheritance</h2>

        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={team.settings.inheritGlobalRules}
            onChange={handleToggleInherit}
            disabled={saving}
            className="w-5 h-5 mt-0.5 rounded"
          />
          <div>
            <span className="font-medium">Inherit Global Rules</span>
            <p className="text-sm text-zinc-500 mt-1">
              When disabled, this team will only receive forced global rules.
              Inheritable global rules will not apply to this team.
            </p>
          </div>
        </label>

        <div className="mt-6 p-4 bg-zinc-50 dark:bg-zinc-700/50 rounded-lg">
          <p className="text-sm">
            <span className="font-medium text-red-600 dark:text-red-400">
              {forcedRulesCount}
            </span>{' '}
            forced rule{forcedRulesCount !== 1 ? 's' : ''} will always apply to this team,
            regardless of this setting.
          </p>
        </div>
      </div>
    </div>
  );
}
