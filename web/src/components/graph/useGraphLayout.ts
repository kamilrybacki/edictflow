import { useMemo } from 'react';
import { Node, Edge, Position, MarkerType } from 'reactflow';
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
        markerEnd: { type: MarkerType.ArrowClosed, color: '#3b82f6' },
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
        markerEnd: { type: MarkerType.ArrowClosed, color: '#3b82f6' },
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
