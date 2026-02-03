'use client';

import { useState, useEffect } from 'react';
import { Team } from '@/domain/team';
import { fetchTeams, createTeam, deleteTeam } from '@/lib/api';

interface TeamListProps {
  selectedTeamId: string | null;
  onSelectTeam: (team: Team | null) => void;
}

export function TeamList({ selectedTeamId, onSelectTeam }: TeamListProps) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newTeamName, setNewTeamName] = useState('');
  const [creating, setCreating] = useState(false);

  const loadTeams = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchTeams();
      setTeams(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load teams');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTeams();
  }, []);

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTeamName.trim()) return;

    try {
      setCreating(true);
      const team = await createTeam(newTeamName.trim());
      setTeams([...teams, team]);
      setNewTeamName('');
      onSelectTeam(team);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create team');
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteTeam = async (team: Team) => {
    if (!confirm(`Delete team "${team.name}"? This will also delete all associated rules.`)) {
      return;
    }

    try {
      await deleteTeam(team.id);
      setTeams(teams.filter((t) => t.id !== team.id));
      if (selectedTeamId === team.id) {
        onSelectTeam(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete team');
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="p-4 border-b border-zinc-200 dark:border-zinc-800">
        <h2 className="text-lg font-semibold mb-3">Teams</h2>
        <form onSubmit={handleCreateTeam} className="flex gap-2">
          <input
            type="text"
            value={newTeamName}
            onChange={(e) => setNewTeamName(e.target.value)}
            placeholder="New team name"
            className="flex-1 px-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={creating || !newTeamName.trim()}
            className="px-3 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {creating ? '...' : 'Add'}
          </button>
        </form>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="p-4 text-center text-zinc-500">Loading teams...</div>
        ) : error ? (
          <div className="p-4">
            <div className="text-red-500 text-sm mb-2">{error}</div>
            <button
              onClick={loadTeams}
              className="text-sm text-blue-600 hover:underline"
            >
              Retry
            </button>
          </div>
        ) : teams.length === 0 ? (
          <div className="p-4 text-center text-zinc-500 text-sm">
            No teams yet. Create one above.
          </div>
        ) : (
          <ul className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {teams.map((team) => (
              <li
                key={team.id}
                className={`flex items-center justify-between p-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                  selectedTeamId === team.id
                    ? 'bg-blue-50 dark:bg-blue-900/20 border-l-2 border-blue-600'
                    : ''
                }`}
                onClick={() => onSelectTeam(team)}
              >
                <div className="flex-1 min-w-0">
                  <div className="font-medium truncate">{team.name}</div>
                  <div className="text-xs text-zinc-500">
                    Drift: {team.settings.driftThresholdMinutes}min
                  </div>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeleteTeam(team);
                  }}
                  className="ml-2 p-1 text-zinc-400 hover:text-red-500"
                  title="Delete team"
                >
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
