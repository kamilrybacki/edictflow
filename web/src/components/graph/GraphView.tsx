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
      const nodeId = node.id.replace(/^(team|user|rule)-/, '');

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
