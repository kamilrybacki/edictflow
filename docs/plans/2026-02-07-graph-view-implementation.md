# Graph View Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a `/graph` page that visualizes relationships between users, teams, and rules using React Flow.

**Architecture:** New Next.js page with React Flow canvas. Custom nodes for teams, users, and rules. Dagre layout for hierarchical positioning. Single API endpoint aggregates all graph data.

**Tech Stack:** React Flow, @dagrejs/dagre, Go chi router, existing RBAC middleware

---

## Task 1: Install React Flow Dependencies

**Files:**
- Modify: `web/package.json`

**Step 1: Install reactflow and dagre**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm install reactflow @dagrejs/dagre @types/d3-hierarchy
```

**Step 2: Verify installation**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm ls reactflow
```

Expected: Shows reactflow version

**Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "feat(web): add reactflow and dagre dependencies"
```

---

## Task 2: Create Graph API Types

**Files:**
- Create: `web/src/lib/api/graph.ts`

**Step 1: Write failing test**

Create: `web/src/__tests__/lib/api/graph.test.ts`

```typescript
import { GraphData, GraphTeam, GraphUser, GraphRule } from '@/lib/api/graph';

describe('Graph API types', () => {
  it('should define GraphTeam with required fields', () => {
    const team: GraphTeam = {
      id: 'team-1',
      name: 'Platform',
      memberCount: 5,
    };
    expect(team.id).toBe('team-1');
    expect(team.name).toBe('Platform');
    expect(team.memberCount).toBe(5);
  });

  it('should define GraphUser with required fields', () => {
    const user: GraphUser = {
      id: 'user-1',
      name: 'Alice',
      email: 'alice@example.com',
      teamId: 'team-1',
    };
    expect(user.id).toBe('user-1');
    expect(user.teamId).toBe('team-1');
  });

  it('should define GraphRule with targeting arrays', () => {
    const rule: GraphRule = {
      id: 'rule-1',
      name: 'No secrets',
      status: 'approved',
      enforcementMode: 'block',
      teamId: 'team-1',
      targetTeams: ['team-2'],
      targetUsers: ['user-1'],
    };
    expect(rule.targetTeams).toContain('team-2');
    expect(rule.targetUsers).toContain('user-1');
  });

  it('should define GraphData combining all types', () => {
    const data: GraphData = {
      teams: [],
      users: [],
      rules: [],
    };
    expect(data.teams).toEqual([]);
  });
});
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm test -- src/__tests__/lib/api/graph.test.ts
```

Expected: FAIL - Cannot find module '@/lib/api/graph'

**Step 3: Write implementation**

Create: `web/src/lib/api/graph.ts`

```typescript
import { RuleStatus, EnforcementMode } from '@/domain/rule';

export interface GraphTeam {
  id: string;
  name: string;
  memberCount: number;
}

export interface GraphUser {
  id: string;
  name: string;
  email: string;
  teamId: string | null;
}

export interface GraphRule {
  id: string;
  name: string;
  status: RuleStatus;
  enforcementMode: EnforcementMode;
  teamId: string | null;
  targetTeams: string[];
  targetUsers: string[];
}

export interface GraphData {
  teams: GraphTeam[];
  users: GraphUser[];
  rules: GraphRule[];
}

export async function fetchGraphData(): Promise<GraphData> {
  const response = await fetch('/api/graph');
  if (!response.ok) {
    throw new Error('Failed to fetch graph data');
  }
  return response.json();
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm test -- src/__tests__/lib/api/graph.test.ts
```

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/lib/api/graph.ts web/src/__tests__/lib/api/graph.test.ts
git commit -m "feat(web): add graph API types and fetch function"
```

---

## Task 3: Create Custom Node Components

**Files:**
- Create: `web/src/components/graph/nodes/TeamNode.tsx`
- Create: `web/src/components/graph/nodes/UserNode.tsx`
- Create: `web/src/components/graph/nodes/RuleNode.tsx`
- Create: `web/src/components/graph/nodes/index.ts`

**Step 1: Create TeamNode**

Create: `web/src/components/graph/nodes/TeamNode.tsx`

```typescript
import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Users } from 'lucide-react';

export interface TeamNodeData {
  name: string;
  memberCount: number;
}

function TeamNodeComponent({ data, selected }: NodeProps<TeamNodeData>) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg border-2 min-w-[140px]
        bg-blue-500 text-white border-blue-600
        ${selected ? 'ring-2 ring-blue-300 ring-offset-2' : ''}
        transition-all duration-200
      `}
    >
      <Handle type="target" position={Position.Top} className="!bg-blue-700" />
      <div className="flex items-center gap-2">
        <Users className="w-4 h-4" />
        <span className="font-medium text-sm">{data.name}</span>
      </div>
      <div className="text-xs text-blue-100 mt-1">
        {data.memberCount} member{data.memberCount !== 1 ? 's' : ''}
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-blue-700" />
    </div>
  );
}

export const TeamNode = memo(TeamNodeComponent);
```

**Step 2: Create UserNode**

Create: `web/src/components/graph/nodes/UserNode.tsx`

```typescript
import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';

export interface UserNodeData {
  name: string;
  email: string;
  initials: string;
}

function UserNodeComponent({ data, selected }: NodeProps<UserNodeData>) {
  return (
    <div
      className={`
        flex flex-col items-center
        ${selected ? 'scale-110' : ''}
        transition-all duration-200
      `}
    >
      <Handle type="target" position={Position.Top} className="!bg-green-700" />
      <div
        className={`
          w-10 h-10 rounded-full flex items-center justify-center
          bg-green-500 text-white font-medium text-sm
          border-2 border-green-600
          ${selected ? 'ring-2 ring-green-300 ring-offset-2' : ''}
        `}
      >
        {data.initials}
      </div>
      <span className="text-xs mt-1 text-foreground max-w-[80px] truncate">
        {data.name}
      </span>
      <Handle type="source" position={Position.Bottom} className="!bg-green-700" />
    </div>
  );
}

export const UserNode = memo(UserNodeComponent);
```

**Step 3: Create RuleNode**

Create: `web/src/components/graph/nodes/RuleNode.tsx`

```typescript
import { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Shield, Check, X, Clock, FileEdit } from 'lucide-react';
import { RuleStatus } from '@/domain/rule';

export interface RuleNodeData {
  name: string;
  status: RuleStatus;
  enforcementMode: string;
}

const statusConfig: Record<RuleStatus, { border: string; badge: React.ReactNode }> = {
  draft: {
    border: 'border-dashed border-amber-500',
    badge: <FileEdit className="w-3 h-3" />,
  },
  pending: {
    border: 'border-amber-500 animate-pulse',
    badge: <Clock className="w-3 h-3" />,
  },
  approved: {
    border: 'border-solid border-amber-600',
    badge: <Check className="w-3 h-3 text-green-500" />,
  },
  rejected: {
    border: 'border-solid border-red-500',
    badge: <X className="w-3 h-3 text-red-500" />,
  },
};

function RuleNodeComponent({ data, selected }: NodeProps<RuleNodeData>) {
  const config = statusConfig[data.status] || statusConfig.draft;

  return (
    <div
      className={`
        px-3 py-2 min-w-[120px]
        bg-amber-500 text-white border-2 rounded-lg
        ${config.border}
        ${selected ? 'ring-2 ring-amber-300 ring-offset-2' : ''}
        transition-all duration-200
      `}
      style={{
        clipPath: 'polygon(10% 0%, 90% 0%, 100% 50%, 90% 100%, 10% 100%, 0% 50%)',
        padding: '12px 20px',
      }}
    >
      <Handle type="target" position={Position.Top} className="!bg-amber-700" />
      <div className="flex items-center gap-2">
        <Shield className="w-4 h-4" />
        <span className="font-medium text-xs">{data.name}</span>
        <span className="ml-auto">{config.badge}</span>
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-amber-700" />
    </div>
  );
}

export const RuleNode = memo(RuleNodeComponent);
```

**Step 4: Create index export**

Create: `web/src/components/graph/nodes/index.ts`

```typescript
export { TeamNode } from './TeamNode';
export type { TeamNodeData } from './TeamNode';
export { UserNode } from './UserNode';
export type { UserNodeData } from './UserNode';
export { RuleNode } from './RuleNode';
export type { RuleNodeData } from './RuleNode';
```

**Step 5: Commit**

```bash
git add web/src/components/graph/nodes/
git commit -m "feat(web): add custom graph node components"
```

---

## Task 4: Create Graph Layout Utility

**Files:**
- Create: `web/src/components/graph/useGraphLayout.ts`

**Step 1: Write failing test**

Create: `web/src/__tests__/components/graph/useGraphLayout.test.ts`

```typescript
import { buildGraphElements } from '@/components/graph/useGraphLayout';
import { GraphData } from '@/lib/api/graph';

describe('buildGraphElements', () => {
  it('should create nodes for teams, users, and rules', () => {
    const data: GraphData = {
      teams: [{ id: 't1', name: 'Team A', memberCount: 2 }],
      users: [{ id: 'u1', name: 'Alice', email: 'a@test.com', teamId: 't1' }],
      rules: [{
        id: 'r1',
        name: 'Rule 1',
        status: 'approved',
        enforcementMode: 'block',
        teamId: 't1',
        targetTeams: [],
        targetUsers: [],
      }],
    };

    const { nodes, edges } = buildGraphElements(data);

    expect(nodes).toHaveLength(3);
    expect(nodes.find(n => n.id === 'team-t1')).toBeDefined();
    expect(nodes.find(n => n.id === 'user-u1')).toBeDefined();
    expect(nodes.find(n => n.id === 'rule-r1')).toBeDefined();
  });

  it('should create edges for user-team membership', () => {
    const data: GraphData = {
      teams: [{ id: 't1', name: 'Team A', memberCount: 1 }],
      users: [{ id: 'u1', name: 'Alice', email: 'a@test.com', teamId: 't1' }],
      rules: [],
    };

    const { edges } = buildGraphElements(data);

    const membershipEdge = edges.find(e => e.id === 'user-u1-belongs-to-team-t1');
    expect(membershipEdge).toBeDefined();
    expect(membershipEdge?.source).toBe('user-u1');
    expect(membershipEdge?.target).toBe('team-t1');
  });

  it('should create edges for rule targeting', () => {
    const data: GraphData = {
      teams: [{ id: 't1', name: 'Team A', memberCount: 0 }],
      users: [{ id: 'u1', name: 'Alice', email: 'a@test.com', teamId: null }],
      rules: [{
        id: 'r1',
        name: 'Rule 1',
        status: 'approved',
        enforcementMode: 'block',
        teamId: null,
        targetTeams: ['t1'],
        targetUsers: ['u1'],
      }],
    };

    const { edges } = buildGraphElements(data);

    expect(edges.find(e => e.id === 'rule-r1-targets-team-t1')).toBeDefined();
    expect(edges.find(e => e.id === 'rule-r1-targets-user-u1')).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm test -- src/__tests__/components/graph/useGraphLayout.test.ts
```

Expected: FAIL - Cannot find module

**Step 3: Write implementation**

Create: `web/src/components/graph/useGraphLayout.ts`

```typescript
import { useMemo } from 'react';
import { Node, Edge, Position } from 'reactflow';
import dagre from '@dagrejs/dagre';
import { GraphData } from '@/lib/api/graph';
import { TeamNodeData } from './nodes/TeamNode';
import { UserNodeData } from './nodes/UserNode';
import { RuleNodeData } from './nodes/RuleNode';

type GraphNode = Node<TeamNodeData | UserNodeData | RuleNodeData>;

const NODE_WIDTH = 150;
const NODE_HEIGHT = 60;

export function buildGraphElements(data: GraphData): { nodes: GraphNode[]; edges: Edge[] } {
  const nodes: GraphNode[] = [];
  const edges: Edge[] = [];

  // Create team nodes
  data.teams.forEach((team) => {
    nodes.push({
      id: `team-${team.id}`,
      type: 'team',
      position: { x: 0, y: 0 },
      data: {
        name: team.name,
        memberCount: team.memberCount,
      },
    });
  });

  // Create user nodes
  data.users.forEach((user) => {
    const initials = user.name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);

    nodes.push({
      id: `user-${user.id}`,
      type: 'user',
      position: { x: 0, y: 0 },
      data: {
        name: user.name,
        email: user.email,
        initials,
      },
    });

    // Edge: user belongs to team
    if (user.teamId) {
      edges.push({
        id: `user-${user.id}-belongs-to-team-${user.teamId}`,
        source: `user-${user.id}`,
        target: `team-${user.teamId}`,
        type: 'default',
        style: { stroke: '#6b7280', strokeWidth: 2 },
        label: 'belongs to',
      });
    }
  });

  // Create rule nodes
  data.rules.forEach((rule) => {
    nodes.push({
      id: `rule-${rule.id}`,
      type: 'rule',
      position: { x: 0, y: 0 },
      data: {
        name: rule.name,
        status: rule.status,
        enforcementMode: rule.enforcementMode,
      },
    });

    // Edge: rule owned by team
    if (rule.teamId) {
      edges.push({
        id: `rule-${rule.id}-owned-by-team-${rule.teamId}`,
        source: `rule-${rule.id}`,
        target: `team-${rule.teamId}`,
        type: 'default',
        style: { stroke: '#9ca3af', strokeWidth: 1, strokeDasharray: '4 2' },
        label: 'owned by',
      });
    }

    // Edges: rule targets teams
    rule.targetTeams.forEach((teamId) => {
      edges.push({
        id: `rule-${rule.id}-targets-team-${teamId}`,
        source: `rule-${rule.id}`,
        target: `team-${teamId}`,
        type: 'default',
        style: { stroke: '#3b82f6', strokeWidth: 2, strokeDasharray: '6 3' },
        markerEnd: { type: 'arrowclosed' as const, color: '#3b82f6' },
        label: 'targets',
      });
    });

    // Edges: rule targets users
    rule.targetUsers.forEach((userId) => {
      edges.push({
        id: `rule-${rule.id}-targets-user-${userId}`,
        source: `rule-${rule.id}`,
        target: `user-${userId}`,
        type: 'default',
        style: { stroke: '#3b82f6', strokeWidth: 2, strokeDasharray: '6 3' },
        markerEnd: { type: 'arrowclosed' as const, color: '#3b82f6' },
        label: 'targets',
      });
    });
  });

  return { nodes, edges };
}

export function applyDagreLayout(
  nodes: GraphNode[],
  edges: Edge[]
): GraphNode[] {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({ rankdir: 'TB', nodesep: 50, ranksep: 100 });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  return nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - NODE_WIDTH / 2,
        y: nodeWithPosition.y - NODE_HEIGHT / 2,
      },
      targetPosition: Position.Top,
      sourcePosition: Position.Bottom,
    };
  });
}

export function useGraphLayout(data: GraphData | null) {
  return useMemo(() => {
    if (!data) {
      return { nodes: [], edges: [] };
    }

    const { nodes, edges } = buildGraphElements(data);
    const layoutedNodes = applyDagreLayout(nodes, edges);

    return { nodes: layoutedNodes, edges };
  }, [data]);
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm test -- src/__tests__/components/graph/useGraphLayout.test.ts
```

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/graph/useGraphLayout.ts web/src/__tests__/components/graph/useGraphLayout.test.ts
git commit -m "feat(web): add graph layout utility with dagre"
```

---

## Task 5: Create GraphPopover Component

**Files:**
- Create: `web/src/components/graph/GraphPopover.tsx`

**Step 1: Create the popover component**

Create: `web/src/components/graph/GraphPopover.tsx`

```typescript
import { X, ExternalLink } from 'lucide-react';
import Link from 'next/link';
import { GraphTeam, GraphUser, GraphRule } from '@/lib/api/graph';

interface BasePopoverProps {
  position: { x: number; y: number };
  onClose: () => void;
}

interface TeamPopoverProps extends BasePopoverProps {
  type: 'team';
  data: GraphTeam;
}

interface UserPopoverProps extends BasePopoverProps {
  type: 'user';
  data: GraphUser;
}

interface RulePopoverProps extends BasePopoverProps {
  type: 'rule';
  data: GraphRule;
}

type GraphPopoverProps = TeamPopoverProps | UserPopoverProps | RulePopoverProps;

export function GraphPopover(props: GraphPopoverProps) {
  const { position, onClose, type, data } = props;

  const getDetailsLink = () => {
    switch (type) {
      case 'team':
        return `/admin/teams/${data.id}/settings`;
      case 'user':
        return `/admin/users/${data.id}`;
      case 'rule':
        return `/rules/${data.id}`;
    }
  };

  const renderContent = () => {
    switch (type) {
      case 'team':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="text-sm text-muted-foreground mt-1">
              {data.memberCount} member{data.memberCount !== 1 ? 's' : ''}
            </div>
          </>
        );
      case 'user':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="text-sm text-muted-foreground mt-1">{data.email}</div>
            {data.teamId && (
              <div className="text-xs text-muted-foreground mt-1">
                Team: {data.teamId}
              </div>
            )}
          </>
        );
      case 'rule':
        return (
          <>
            <div className="font-medium text-foreground">{data.name}</div>
            <div className="flex gap-2 mt-2">
              <span className="text-xs px-2 py-0.5 rounded bg-muted">
                {data.status}
              </span>
              <span className="text-xs px-2 py-0.5 rounded bg-muted">
                {data.enforcementMode}
              </span>
            </div>
          </>
        );
    }
  };

  return (
    <div
      className="absolute z-50 bg-card border rounded-lg shadow-lg p-3 min-w-[180px]"
      style={{
        left: position.x,
        top: position.y,
        transform: 'translate(-50%, -100%) translateY(-10px)',
      }}
    >
      <button
        onClick={onClose}
        className="absolute top-2 right-2 text-muted-foreground hover:text-foreground"
      >
        <X className="w-4 h-4" />
      </button>

      <div className="pr-6">{renderContent()}</div>

      <Link
        href={getDetailsLink()}
        className="flex items-center gap-1 text-xs text-primary hover:underline mt-3"
      >
        View details <ExternalLink className="w-3 h-3" />
      </Link>
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add web/src/components/graph/GraphPopover.tsx
git commit -m "feat(web): add graph popover component"
```

---

## Task 6: Create GraphControls Component

**Files:**
- Create: `web/src/components/graph/GraphControls.tsx`

**Step 1: Create the controls component**

Create: `web/src/components/graph/GraphControls.tsx`

```typescript
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
```

**Step 2: Commit**

```bash
git add web/src/components/graph/GraphControls.tsx
git commit -m "feat(web): add graph controls component"
```

---

## Task 7: Create GraphView Component

**Files:**
- Create: `web/src/components/graph/GraphView.tsx`
- Create: `web/src/components/graph/index.ts`

**Step 1: Create the main GraphView component**

Create: `web/src/components/graph/GraphView.tsx`

```typescript
'use client';

import { useState, useCallback, useRef } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
  NodeMouseHandler,
  useReactFlow,
  ReactFlowProvider,
} from 'reactflow';
import 'reactflow/dist/style.css';

import { GraphData, GraphTeam, GraphUser, GraphRule } from '@/lib/api/graph';
import { useGraphLayout } from './useGraphLayout';
import { TeamNode, UserNode, RuleNode } from './nodes';
import { GraphPopover } from './GraphPopover';
import { RuleStatus } from '@/domain/rule';

const nodeTypes = {
  team: TeamNode,
  user: UserNode,
  rule: RuleNode,
};

interface SelectedNode {
  type: 'team' | 'user' | 'rule';
  data: GraphTeam | GraphUser | GraphRule;
  position: { x: number; y: number };
}

interface GraphViewProps {
  data: GraphData;
  selectedTeamId: string | null;
  statusFilters: RuleStatus[];
  searchQuery: string;
}

function GraphViewInner({
  data,
  selectedTeamId,
  statusFilters,
  searchQuery,
}: GraphViewProps) {
  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { fitView } = useReactFlow();

  // Filter data based on controls
  const filteredData: GraphData = {
    teams: selectedTeamId
      ? data.teams.filter((t) => t.id === selectedTeamId)
      : data.teams,
    users: data.users.filter((u) => {
      if (selectedTeamId && u.teamId !== selectedTeamId) return false;
      if (searchQuery && !u.name.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }
      return true;
    }),
    rules: data.rules.filter((r) => {
      if (statusFilters.length > 0 && !statusFilters.includes(r.status)) {
        return false;
      }
      if (selectedTeamId) {
        const isOwned = r.teamId === selectedTeamId;
        const targetsTeam = r.targetTeams.includes(selectedTeamId);
        if (!isOwned && !targetsTeam) return false;
      }
      if (searchQuery && !r.name.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }
      return true;
    }),
  };

  const { nodes, edges } = useGraphLayout(filteredData);

  // Highlight connected edges on node click
  const [highlightedNodeId, setHighlightedNodeId] = useState<string | null>(null);

  const getHighlightedEdges = useCallback(
    (nodeId: string | null): Edge[] => {
      if (!nodeId) return edges;

      return edges.map((edge) => {
        const isConnected = edge.source === nodeId || edge.target === nodeId;
        return {
          ...edge,
          style: {
            ...edge.style,
            opacity: isConnected ? 1 : 0.2,
          },
        };
      });
    },
    [edges]
  );

  const getHighlightedNodes = useCallback(
    (nodeId: string | null): Node[] => {
      if (!nodeId) return nodes;

      const connectedNodeIds = new Set<string>([nodeId]);
      edges.forEach((edge) => {
        if (edge.source === nodeId) connectedNodeIds.add(edge.target);
        if (edge.target === nodeId) connectedNodeIds.add(edge.source);
      });

      return nodes.map((node) => ({
        ...node,
        style: {
          ...node.style,
          opacity: connectedNodeIds.has(node.id) ? 1 : 0.3,
        },
      }));
    },
    [nodes, edges]
  );

  const handleNodeClick: NodeMouseHandler = useCallback(
    (event, node) => {
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;

      setHighlightedNodeId(node.id);

      // Find original data
      const [type, id] = node.id.split('-').slice(0, 2);
      const nodeId = node.id.replace(`${type}-`, '');

      let nodeData: GraphTeam | GraphUser | GraphRule | undefined;
      let nodeType: 'team' | 'user' | 'rule';

      if (node.id.startsWith('team-')) {
        nodeType = 'team';
        nodeData = data.teams.find((t) => t.id === nodeId);
      } else if (node.id.startsWith('user-')) {
        nodeType = 'user';
        nodeData = data.users.find((u) => u.id === nodeId);
      } else {
        nodeType = 'rule';
        nodeData = data.rules.find((r) => r.id === nodeId);
      }

      if (nodeData) {
        setSelectedNode({
          type: nodeType,
          data: nodeData,
          position: { x: event.clientX - rect.left, y: event.clientY - rect.top },
        });
      }
    },
    [data]
  );

  const handlePaneClick = useCallback(() => {
    setSelectedNode(null);
    setHighlightedNodeId(null);
  }, []);

  const displayedNodes = getHighlightedNodes(highlightedNodeId);
  const displayedEdges = getHighlightedEdges(highlightedNodeId);

  return (
    <div ref={containerRef} className="w-full h-full relative">
      <ReactFlow
        nodes={displayedNodes}
        edges={displayedEdges}
        nodeTypes={nodeTypes}
        onNodeClick={handleNodeClick}
        onPaneClick={handlePaneClick}
        fitView
        minZoom={0.1}
        maxZoom={2}
      >
        <Background />
        <Controls />
        <MiniMap
          nodeColor={(node) => {
            if (node.type === 'team') return '#3b82f6';
            if (node.type === 'user') return '#22c55e';
            return '#f59e0b';
          }}
        />
      </ReactFlow>

      {selectedNode && (
        <GraphPopover
          type={selectedNode.type}
          data={selectedNode.data as any}
          position={selectedNode.position}
          onClose={() => {
            setSelectedNode(null);
            setHighlightedNodeId(null);
          }}
        />
      )}
    </div>
  );
}

export function GraphView(props: GraphViewProps) {
  return (
    <ReactFlowProvider>
      <GraphViewInner {...props} />
    </ReactFlowProvider>
  );
}
```

**Step 2: Create index export**

Create: `web/src/components/graph/index.ts`

```typescript
export { GraphView } from './GraphView';
export { GraphControls } from './GraphControls';
export { GraphPopover } from './GraphPopover';
export { useGraphLayout, buildGraphElements } from './useGraphLayout';
export * from './nodes';
```

**Step 3: Commit**

```bash
git add web/src/components/graph/
git commit -m "feat(web): add main GraphView component with React Flow"
```

---

## Task 8: Create Graph Page

**Files:**
- Create: `web/src/app/graph/page.tsx`

**Step 1: Create the page component**

Create: `web/src/app/graph/page.tsx`

```typescript
'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Network } from 'lucide-react';
import { useRequireAuth } from '@/contexts/AuthContext';
import { GraphData, fetchGraphData } from '@/lib/api/graph';
import { GraphView, GraphControls } from '@/components/graph';
import { RuleStatus } from '@/domain/rule';

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
```

**Step 2: Commit**

```bash
git add web/src/app/graph/
git commit -m "feat(web): add graph page at /graph route"
```

---

## Task 9: Create Next.js API Route Proxy

**Files:**
- Create: `web/src/app/api/graph/route.ts`

**Step 1: Create the API route**

Create: `web/src/app/api/graph/route.ts`

```typescript
import { NextRequest, NextResponse } from 'next/server';

const API_BASE = process.env.API_URL || 'http://localhost:8080';

export async function GET(request: NextRequest) {
  const authHeader = request.headers.get('Authorization');
  const cookieHeader = request.headers.get('Cookie');

  try {
    const response = await fetch(`${API_BASE}/api/v1/graph`, {
      headers: {
        ...(authHeader && { Authorization: authHeader }),
        ...(cookieHeader && { Cookie: cookieHeader }),
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch graph data' },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Graph API error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
```

**Step 2: Commit**

```bash
git add web/src/app/api/graph/
git commit -m "feat(web): add graph API route proxy"
```

---

## Task 10: Create Backend Graph Handler

**Files:**
- Create: `server/entrypoints/api/handlers/graph.go`

**Step 1: Write failing test**

Create: `server/tests/unit/handlers/graph_test.go`

```go
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
)

type mockGraphTeamService struct {
	teams []domain.Team
}

func (m *mockGraphTeamService) List() ([]domain.Team, error) {
	return m.teams, nil
}

type mockGraphUserService struct {
	users []domain.User
}

func (m *mockGraphUserService) List(teamID string, activeOnly bool) ([]domain.User, error) {
	return m.users, nil
}

func (m *mockGraphUserService) CountByTeam(teamID string) (int, error) {
	count := 0
	for _, u := range m.users {
		if u.TeamID != nil && *u.TeamID == teamID {
			count++
		}
	}
	return count, nil
}

type mockGraphRuleService struct {
	rules []domain.Rule
}

func (m *mockGraphRuleService) ListAll() ([]domain.Rule, error) {
	return m.rules, nil
}

func TestGraphHandler_Get(t *testing.T) {
	teamID := "team-1"
	teams := []domain.Team{{ID: teamID, Name: "Platform"}}
	users := []domain.User{{ID: "user-1", Name: "Alice", Email: "alice@test.com", TeamID: &teamID}}
	rules := []domain.Rule{{
		ID:          "rule-1",
		Name:        "Test Rule",
		Status:      domain.RuleStatusApproved,
		TargetTeams: []string{teamID},
	}}

	handler := handlers.NewGraphHandler(
		&mockGraphTeamService{teams: teams},
		&mockGraphUserService{users: users},
		&mockGraphRuleService{rules: rules},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response handlers.GraphResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Teams) != 1 {
		t.Errorf("expected 1 team, got %d", len(response.Teams))
	}
	if len(response.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(response.Users))
	}
	if len(response.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(response.Rules))
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go test ./tests/unit/handlers/graph_test.go -v
```

Expected: FAIL - undefined: handlers.NewGraphHandler

**Step 3: Write implementation**

Create: `server/entrypoints/api/handlers/graph.go`

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kamilrybacki/edictflow/server/domain"
)

// GraphTeamService defines the team operations needed for graph data
type GraphTeamService interface {
	List() ([]domain.Team, error)
}

// GraphUserService defines the user operations needed for graph data
type GraphUserService interface {
	List(teamID string, activeOnly bool) ([]domain.User, error)
	CountByTeam(teamID string) (int, error)
}

// GraphRuleService defines the rule operations needed for graph data
type GraphRuleService interface {
	ListAll() ([]domain.Rule, error)
}

// GraphTeam represents a team in the graph response
type GraphTeam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

// GraphUser represents a user in the graph response
type GraphUser struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	TeamID *string `json:"teamId"`
}

// GraphRule represents a rule in the graph response
type GraphRule struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Status          domain.RuleStatus `json:"status"`
	EnforcementMode string            `json:"enforcementMode"`
	TeamID          *string           `json:"teamId"`
	TargetTeams     []string          `json:"targetTeams"`
	TargetUsers     []string          `json:"targetUsers"`
}

// GraphResponse is the complete graph data response
type GraphResponse struct {
	Teams []GraphTeam `json:"teams"`
	Users []GraphUser `json:"users"`
	Rules []GraphRule `json:"rules"`
}

// GraphHandler handles graph-related API requests
type GraphHandler struct {
	teamService GraphTeamService
	userService GraphUserService
	ruleService GraphRuleService
}

// NewGraphHandler creates a new GraphHandler
func NewGraphHandler(
	teamService GraphTeamService,
	userService GraphUserService,
	ruleService GraphRuleService,
) *GraphHandler {
	return &GraphHandler{
		teamService: teamService,
		userService: userService,
		ruleService: ruleService,
	}
}

// Get returns all graph data in a single response
func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Fetch all teams
	teams, err := h.teamService.List()
	if err != nil {
		http.Error(w, "Failed to fetch teams", http.StatusInternalServerError)
		return
	}

	// Fetch all users
	users, err := h.userService.List("", true)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Fetch all rules
	rules, err := h.ruleService.ListAll()
	if err != nil {
		http.Error(w, "Failed to fetch rules", http.StatusInternalServerError)
		return
	}

	// Build response
	response := GraphResponse{
		Teams: make([]GraphTeam, 0, len(teams)),
		Users: make([]GraphUser, 0, len(users)),
		Rules: make([]GraphRule, 0, len(rules)),
	}

	// Transform teams with member counts
	for _, team := range teams {
		count, _ := h.userService.CountByTeam(team.ID)
		response.Teams = append(response.Teams, GraphTeam{
			ID:          team.ID,
			Name:        team.Name,
			MemberCount: count,
		})
	}

	// Transform users
	for _, user := range users {
		response.Users = append(response.Users, GraphUser{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			TeamID: user.TeamID,
		})
	}

	// Transform rules
	for _, rule := range rules {
		targetTeams := rule.TargetTeams
		if targetTeams == nil {
			targetTeams = []string{}
		}
		targetUsers := rule.TargetUsers
		if targetUsers == nil {
			targetUsers = []string{}
		}

		response.Rules = append(response.Rules, GraphRule{
			ID:              rule.ID,
			Name:            rule.Name,
			Status:          rule.Status,
			EnforcementMode: string(rule.EnforcementMode),
			TeamID:          rule.TeamID,
			TargetTeams:     targetTeams,
			TargetUsers:     targetUsers,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go test ./tests/unit/handlers/graph_test.go -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add server/entrypoints/api/handlers/graph.go server/tests/unit/handlers/graph_test.go
git commit -m "feat(server): add graph handler with tests"
```

---

## Task 11: Wire Up Graph Route in Router

**Files:**
- Modify: `server/entrypoints/api/router.go`

**Step 1: Add GraphService to Config struct**

In `server/entrypoints/api/router.go`, find the `Config` struct and add:

```go
GraphTeamService    handlers.GraphTeamService
GraphUserService    handlers.GraphUserService
GraphRuleService    handlers.GraphRuleService
```

**Step 2: Add graph route**

In the `NewRouter` function, after the audit routes (around line 255), add:

```go
// Graph route
if cfg.GraphTeamService != nil && cfg.GraphUserService != nil && cfg.GraphRuleService != nil {
	r.Route("/graph", func(r chi.Router) {
		h := handlers.NewGraphHandler(cfg.GraphTeamService, cfg.GraphUserService, cfg.GraphRuleService)
		r.Get("/", h.Get)
	})
}
```

**Step 3: Run server tests**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go test ./... -v -count=1 2>&1 | tail -30
```

Expected: All tests pass

**Step 4: Commit**

```bash
git add server/entrypoints/api/router.go
git commit -m "feat(server): wire up graph route in router"
```

---

## Task 12: Add ListAll Method to Rules Service

**Files:**
- Modify: `server/services/rules/service.go`

**Step 1: Check if ListAll exists**

Search for existing `ListAll` method in rules service. If it doesn't exist, add it.

**Step 2: Add ListAll method if needed**

Add to `server/services/rules/service.go`:

```go
// ListAll returns all rules across all teams
func (s *Service) ListAll() ([]domain.Rule, error) {
	return s.repo.ListAll()
}
```

**Step 3: Add to repository if needed**

Check `server/adapters/postgres/rule_db.go` for `ListAll`. If missing, add:

```go
// ListAll returns all rules
func (r *RuleRepository) ListAll() ([]domain.Rule, error) {
	var rules []domain.Rule
	err := r.db.Select(&rules, `
		SELECT id, name, content, description, target_layer, category_id,
			   priority_weight, overridable, effective_start, effective_end,
			   target_teams, target_users, tags, triggers, team_id, force,
			   status, enforcement_mode, temporary_timeout_hours,
			   created_by, submitted_at, approved_at, created_at, updated_at
		FROM rules
		ORDER BY created_at DESC
	`)
	return rules, err
}
```

**Step 4: Run tests**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go test ./... -v -count=1 2>&1 | tail -30
```

**Step 5: Commit**

```bash
git add server/services/rules/ server/adapters/postgres/
git commit -m "feat(server): add ListAll method to rules service"
```

---

## Task 13: Wire Graph Handler in Main

**Files:**
- Modify: `server/cmd/server/main.go` or `server/cmd/master/services.go`

**Step 1: Find where router config is built**

Check `server/cmd/master/services.go` for the router configuration.

**Step 2: Add graph services to router config**

Add the graph services to the router config:

```go
GraphTeamService:  teamService,
GraphUserService:  userService,
GraphRuleService:  ruleService,
```

**Step 3: Run server**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go build ./cmd/server
```

Expected: Build succeeds

**Step 4: Commit**

```bash
git add server/cmd/
git commit -m "feat(server): wire graph handler in main initialization"
```

---

## Task 14: Add Navigation Link to Graph

**Files:**
- Modify: `web/src/components/dashboard/DashboardLayout.tsx`

**Step 1: Add graph link to navigation**

Find the navigation section in `DashboardLayout.tsx` and add a link to the graph page:

```typescript
import { Network } from 'lucide-react';

// In the navigation items, add:
{ href: '/graph', icon: Network, label: 'Graph View' }
```

**Step 2: Verify the link works**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm run dev
```

Navigate to http://localhost:3000 and verify the Graph View link appears and works.

**Step 3: Commit**

```bash
git add web/src/components/dashboard/DashboardLayout.tsx
git commit -m "feat(web): add graph view link to dashboard navigation"
```

---

## Task 15: Final Integration Test

**Step 1: Run all tests**

Run:
```bash
cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/web
npm test

cd /Users/kamilrybacki/Projects/Personal/edictflow/.worktrees/graph-view/server
go test ./...
```

Expected: All tests pass

**Step 2: Manual smoke test**

1. Start the server and web app
2. Navigate to `/graph`
3. Verify nodes appear for teams, users, and rules
4. Click a node - verify popover appears
5. Use filters - verify graph updates
6. Search for a node - verify highlighting works

**Step 3: Final commit**

If any fixes were needed, commit them:

```bash
git add -A
git commit -m "fix: address integration issues from smoke test"
```

---

## Summary

This plan implements:

1. **Frontend (8 tasks):**
   - React Flow + dagre dependencies
   - Graph API types
   - Custom node components (Team, User, Rule)
   - Layout utility with dagre
   - Popover component
   - Controls component
   - Main GraphView component
   - Graph page at `/graph`

2. **Backend (4 tasks):**
   - Graph handler with tests
   - Router wiring
   - ListAll method for rules
   - Main initialization

3. **Integration (2 tasks):**
   - Navigation link
   - Final testing

Total: 15 tasks
