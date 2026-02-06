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
