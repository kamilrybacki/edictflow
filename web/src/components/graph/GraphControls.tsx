import { Search, Filter } from 'lucide-react';
import { RuleStatus } from '@/domain/rule';

interface GraphControlsProps {
  teams: { id: string; name: string }[];
  selectedTeamId: string | null;
  onSelectTeam: (teamId: string | null) => void;
  statusFilters: RuleStatus[];
  onToggleStatus: (status: RuleStatus) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
}

const statuses: RuleStatus[] = ['draft', 'pending', 'approved', 'rejected'];

export function GraphControls({
  teams,
  selectedTeamId,
  onSelectTeam,
  statusFilters,
  onToggleStatus,
  searchQuery,
  onSearchChange,
}: GraphControlsProps) {
  return (
    <div className="flex items-center gap-4 p-4 bg-card border-b">
      {/* Team Filter */}
      <div className="flex items-center gap-2">
        <Filter className="w-4 h-4 text-muted-foreground" />
        <select
          value={selectedTeamId || ''}
          onChange={(e) => onSelectTeam(e.target.value || null)}
          className="px-3 py-1.5 rounded border bg-background text-sm"
        >
          <option value="">All Teams</option>
          {teams.map((team) => (
            <option key={team.id} value={team.id}>
              {team.name}
            </option>
          ))}
        </select>
      </div>

      {/* Status Toggles */}
      <div className="flex items-center gap-2">
        {statuses.map((status) => (
          <button
            key={status}
            onClick={() => onToggleStatus(status)}
            className={`
              px-2 py-1 text-xs rounded capitalize
              ${
                statusFilters.includes(status)
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground'
              }
            `}
          >
            {status}
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="flex-1 max-w-xs ml-auto relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search nodes..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="w-full pl-9 pr-3 py-1.5 rounded border bg-background text-sm"
        />
      </div>
    </div>
  );
}
