'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { fetchTeam, fetchGlobalRules } from '@/lib/api';
import { Team } from '@/domain/team';
import { Rule } from '@/domain/rule';
import { Shield, Check } from 'lucide-react';

export default function TeamSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const teamId = params.id as string;
  const [team, setTeam] = useState<Team | null>(null);
  const [globalRulesCount, setGlobalRulesCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [teamData, globalRules] = await Promise.all([
          fetchTeam(teamId),
          fetchGlobalRules(),
        ]);
        setTeam(teamData);
        setGlobalRulesCount(globalRules.length);
      } catch (err) {
        console.error('Failed to load team settings:', err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [teamId]);

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
        &larr; Back
      </button>

      <h1 className="text-2xl font-bold mb-6">{team.name} Settings</h1>

      <div className="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow">
        <h2 className="text-lg font-semibold mb-4">Global Rules Inheritance</h2>

        <div className="flex items-start gap-3 p-4 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800">
          <div className="w-8 h-8 rounded-full bg-green-500 flex items-center justify-center flex-shrink-0">
            <Check className="w-5 h-5 text-white" />
          </div>
          <div>
            <span className="font-medium text-green-800 dark:text-green-200">
              Global Rules Active
            </span>
            <p className="text-sm text-green-700 dark:text-green-300 mt-1">
              This team automatically inherits all enterprise-level global rules.
              This ensures consistent governance across the organization.
            </p>
          </div>
        </div>

        <div className="mt-6 p-4 bg-zinc-50 dark:bg-zinc-700/50 rounded-lg flex items-center gap-3">
          <Shield className="w-5 h-5 text-layer-enterprise" />
          <p className="text-sm">
            <span className="font-medium">
              {globalRulesCount}
            </span>{' '}
            global rule{globalRulesCount !== 1 ? 's' : ''} currently apply to this team.
          </p>
        </div>
      </div>
    </div>
  );
}
