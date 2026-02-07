'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { Network } from 'lucide-react';
import { useRequireAuth } from '@/contexts/AuthContext';
import { GraphData, fetchGraphData } from '@/lib/api/graph';
import { GraphControls } from '@/components/graph/GraphControls';
import { RuleStatus } from '@/domain/rule';

// Dynamic import to avoid SSR issues with ReactFlow
const GraphView = dynamic(
  () => import('@/components/graph/GraphView').then((mod) => mod.GraphView),
  { ssr: false, loading: () => <div className="flex-1 flex items-center justify-center text-muted-foreground">Loading graph...</div> }
);

export default function GraphPage() {
  const auth = useRequireAuth();
  const router = useRouter();

  const [data, setData] = useState<GraphData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filter state
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const [statusFilters, setStatusFilters] = useState<RuleStatus[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    async function loadData() {
      try {
        setIsLoading(true);
        const graphData = await fetchGraphData();
        setData(graphData);
      } catch (err) {
        setError('Failed to load graph data');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }

    if (auth.isAuthenticated) {
      loadData();
    }
  }, [auth.isAuthenticated]);

  const handleToggleStatus = (status: RuleStatus) => {
    setStatusFilters((prev) =>
      prev.includes(status)
        ? prev.filter((s) => s !== status)
        : [...prev, status]
    );
  };

  if (auth.isLoading || !auth.isAuthenticated) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-muted-foreground">Loading graph...</div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-destructive">{error || 'No data available'}</div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Header */}
      <header className="flex items-center gap-3 px-6 py-4 border-b bg-card">
        <button
          onClick={() => router.push('/')}
          className="text-muted-foreground hover:text-foreground"
        >
          ← Back
        </button>
        <Network className="w-6 h-6 text-primary" />
        <h1 className="text-xl font-semibold">Organization Graph</h1>
      </header>

      {/* Controls */}
      <GraphControls
        teams={data.teams}
        selectedTeamId={selectedTeamId}
        onSelectTeam={setSelectedTeamId}
        statusFilters={statusFilters}
        onToggleStatus={handleToggleStatus}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
      />

      {/* Graph */}
      <div className="flex-1">
        <GraphView
          data={data}
          selectedTeamId={selectedTeamId}
          statusFilters={statusFilters}
          searchQuery={searchQuery}
        />
      </div>
    </div>
  );
}
