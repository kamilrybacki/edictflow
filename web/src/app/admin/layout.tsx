'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState, useEffect } from 'react';
import { useRequirePermission } from '@/contexts/AuthContext';
import { fetchConnectedAgents, ConnectedAgent } from '@/lib/api';

const navItems = [
  { href: '/admin/users', label: 'Users', icon: 'users' },
  { href: '/admin/roles', label: 'Roles', icon: 'shield' },
  { href: '/admin/audit', label: 'Audit Log', icon: 'list' },
];

function AgentIcon() {
  return (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
    </svg>
  );
}

function ConnectedAgentsPanel() {
  const [agents, setAgents] = useState<ConnectedAgent[]>([]);
  const [loading, setLoading] = useState(true);
  const [expanded, setExpanded] = useState(false);

  useEffect(() => {
    const loadAgents = async () => {
      try {
        const data = await fetchConnectedAgents();
        setAgents(data);
      } catch {
        // Silently fail - worker might not be available
        setAgents([]);
      } finally {
        setLoading(false);
      }
    };

    loadAgents();
    // Refresh every 30 seconds
    const interval = setInterval(loadAgents, 30000);
    return () => clearInterval(interval);
  }, []);

  const agentCount = agents.length;

  return (
    <div className="border-t border-zinc-200 dark:border-zinc-700 p-4">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center justify-between w-full text-left"
      >
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${agentCount > 0 ? 'bg-green-500' : 'bg-zinc-400'}`} />
          <span className="text-sm font-medium text-zinc-600 dark:text-zinc-400">
            Connected Agents
          </span>
        </div>
        <span className="text-sm font-semibold text-zinc-900 dark:text-white">
          {loading ? '...' : agentCount}
        </span>
      </button>

      {expanded && !loading && (
        <div className="mt-3 space-y-2 max-h-48 overflow-y-auto">
          {agents.length === 0 ? (
            <p className="text-xs text-zinc-500 italic">No agents connected</p>
          ) : (
            agents.map((agent) => (
              <div
                key={agent.agent_id}
                className="flex items-center gap-2 px-2 py-1.5 bg-zinc-50 dark:bg-zinc-700/50 rounded text-xs"
              >
                <AgentIcon />
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-zinc-700 dark:text-zinc-300 truncate">
                    {agent.agent_id.slice(0, 8)}...
                  </p>
                  <p className="text-zinc-500 dark:text-zinc-400 truncate">
                    Team: {agent.team_id.slice(0, 8)}...
                  </p>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

function NavIcon({ name }: { name: string }) {
  switch (name) {
    case 'users':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m12 5.197v-1a6 6 0 00-6-6" />
        </svg>
      );
    case 'shield':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      );
    case 'list':
      return (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      );
    default:
      return null;
  }
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const auth = useRequirePermission('admin_access');

  if (auth.isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-900">
        <div className="text-zinc-600 dark:text-zinc-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-900">
      {/* Top Navigation */}
      <header className="bg-white dark:bg-zinc-800 border-b border-zinc-200 dark:border-zinc-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link href="/" className="text-xl font-bold text-zinc-900 dark:text-white">
                Edictflow
              </Link>
              <span className="ml-4 px-2 py-1 text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded">
                Admin
              </span>
            </div>
            <div className="flex items-center">
              <Link
                href="/"
                className="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white"
              >
                Back to Dashboard
              </Link>
            </div>
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <aside className="w-64 min-h-[calc(100vh-4rem)] bg-white dark:bg-zinc-800 border-r border-zinc-200 dark:border-zinc-700 flex flex-col">
          <nav className="p-4 space-y-1 flex-1">
            {navItems.map((item) => {
              const isActive = pathname.startsWith(item.href);
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400'
                      : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-700'
                  }`}
                >
                  <NavIcon name={item.icon} />
                  {item.label}
                </Link>
              );
            })}
          </nav>
          <ConnectedAgentsPanel />
        </aside>

        {/* Main Content */}
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
