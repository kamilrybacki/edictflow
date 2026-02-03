import { Team } from '@/domain/team';
import { Rule } from '@/domain/rule';
import { User, LoginRequest, RegisterRequest, AuthResponse } from '@/domain/user';

// API URL is determined at runtime in the browser
function getApiUrl(): string {
  if (typeof window !== 'undefined') {
    // Client-side: check for runtime config
    const runtimeUrl = (window as unknown as { __API_URL__?: string }).__API_URL__;
    if (runtimeUrl) return runtimeUrl;

    // Use environment variable if available (set at build time)
    if (process.env.NEXT_PUBLIC_API_URL) {
      return process.env.NEXT_PUBLIC_API_URL;
    }

    // Default fallback - assume API is on port 8081
    return `http://${window.location.hostname}:8081`;
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
}

// Cache the API URL
let cachedApiUrl: string | null = null;
function getApiUrlCached(): string {
  if (!cachedApiUrl) {
    cachedApiUrl = getApiUrl();
  }
  return cachedApiUrl;
}

const TOKEN_KEY = 'auth_token';

// Get auth headers with token from localStorage
function getAuthHeaders(): HeadersInit {
  const token = typeof window !== 'undefined' ? localStorage.getItem(TOKEN_KEY) : null;
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

// Auth API
export async function login(request: LoginRequest): Promise<AuthResponse> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || 'Login failed');
  }
  return res.json();
}

export async function register(request: RegisterRequest): Promise<AuthResponse> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || 'Registration failed');
  }
  return res.json();
}

export async function fetchCurrentUser(): Promise<User> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/auth/me`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error('Failed to fetch current user');
  }
  return res.json();
}

// Teams API
export async function fetchTeams(): Promise<Team[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch teams: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchTeam(id: string): Promise<Team> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${id}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch team: ${res.statusText}`);
  }
  return res.json();
}

export async function createTeam(name: string): Promise<Team> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    throw new Error(`Failed to create team: ${res.statusText}`);
  }
  return res.json();
}

export async function deleteTeam(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/teams/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete team: ${res.statusText}`);
  }
}

// Rules API
export async function fetchRules(teamId: string, status?: string): Promise<Rule[]> {
  let url = `${getApiUrlCached()}/api/v1/rules/?team_id=${teamId}`;
  if (status) {
    url += `&status=${status}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch rules: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchRule(id: string): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch rule: ${res.statusText}`);
  }
  return res.json();
}

export interface CreateRuleRequest {
  name: string;
  content: string;
  target_layer: string;
  team_id: string;
  triggers: Array<{
    type: string;
    pattern?: string;
    context_types?: string[];
    tags?: string[];
  }>;
}

export async function createRule(request: CreateRuleRequest): Promise<Rule> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error(`Failed to create rule: ${res.statusText}`);
  }
  return res.json();
}

export async function updateRule(id: string, request: CreateRuleRequest): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error(`Failed to update rule: ${res.statusText}`);
  }
}

export async function deleteRule(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/rules/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete rule: ${res.statusText}`);
  }
}

// Approvals API
export async function submitRuleForApproval(ruleId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}/submit`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to submit rule: ${res.statusText}`);
  }
}

export async function approveRule(ruleId: string, comment?: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}/approve`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ comment }),
  });
  if (!res.ok) {
    throw new Error(`Failed to approve rule: ${res.statusText}`);
  }
}

export async function rejectRule(ruleId: string, comment: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}/reject`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ comment }),
  });
  if (!res.ok) {
    throw new Error(`Failed to reject rule: ${res.statusText}`);
  }
}

export interface ApprovalStatus {
  rule_id: string;
  status: string;
  required_count: number;
  current_count: number;
  approvals: Array<{
    id: string;
    user_id: string;
    user_name?: string;
    decision: string;
    comment?: string;
    created_at: string;
  }>;
}

export async function getApprovalStatus(ruleId: string): Promise<ApprovalStatus> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/approvals/rules/${ruleId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to get approval status: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchPendingApprovals(scope?: string): Promise<Rule[]> {
  let url = `${getApiUrlCached()}/api/v1/approvals/pending`;
  if (scope) {
    url += `?scope=${scope}`;
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch pending approvals: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

// Users API (Admin)
export async function fetchUsers(): Promise<User[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch users: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchUser(id: string): Promise<User> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/${id}`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch user: ${res.statusText}`);
  }
  return res.json();
}

export async function deactivateUser(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/${id}/deactivate`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to deactivate user: ${res.statusText}`);
  }
}

export async function assignUserRole(userId: string, roleId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/users/${userId}/roles`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ role_id: roleId }),
  });
  if (!res.ok) {
    throw new Error(`Failed to assign role: ${res.statusText}`);
  }
}

// Roles API (Admin)
export interface Role {
  id: string;
  name: string;
  description: string;
  hierarchy_level: number;
  permissions: string[];
  created_at: string;
}

export async function fetchRoles(): Promise<Role[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch roles: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function createRole(data: { name: string; description: string; hierarchy_level: number }): Promise<Role> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    throw new Error(`Failed to create role: ${res.statusText}`);
  }
  return res.json();
}

export async function addRolePermission(roleId: string, permissionId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/${roleId}/permissions`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ permission_id: permissionId }),
  });
  if (!res.ok) {
    throw new Error(`Failed to add permission: ${res.statusText}`);
  }
}

export async function removeRolePermission(roleId: string, permissionId: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/roles/${roleId}/permissions/${permissionId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to remove permission: ${res.statusText}`);
  }
}

// Permissions API (Admin)
export interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
}

export async function fetchPermissions(): Promise<Permission[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/permissions/`, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch permissions: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

// Audit API (Admin)
export interface AuditEntry {
  id: string;
  entity_type: string;
  entity_id: string;
  action: string;
  actor_id?: string;
  actor_name?: string;
  changes?: Record<string, { old: unknown; new: unknown }>;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditListResponse {
  entries: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}

export async function fetchAuditLogs(params: {
  entity_type?: string;
  action?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}): Promise<AuditListResponse> {
  const searchParams = new URLSearchParams();
  if (params.entity_type) searchParams.set('entity_type', params.entity_type);
  if (params.action) searchParams.set('action', params.action);
  if (params.from) searchParams.set('from', params.from);
  if (params.to) searchParams.set('to', params.to);
  if (params.limit) searchParams.set('limit', params.limit.toString());
  if (params.offset) searchParams.set('offset', params.offset.toString());

  const res = await fetch(`${getApiUrlCached()}/api/v1/audit/?${searchParams}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch audit logs: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchEntityHistory(entityType: string, entityId: string): Promise<AuditEntry[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/audit/entity/${entityType}/${entityId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch entity history: ${res.statusText}`);
  }
  return res.json();
}

// Health check (no auth required)
export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${getApiUrlCached()}/health`);
    return res.ok;
  } catch {
    return false;
  }
}

export async function fetchServiceInfo(): Promise<{ service: string; version: string; status: string } | null> {
  try {
    const res = await fetch(`${getApiUrlCached()}/`);
    if (res.ok) {
      return res.json();
    }
    return null;
  } catch {
    return null;
  }
}
