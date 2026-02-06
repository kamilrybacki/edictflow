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
