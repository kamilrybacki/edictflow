# Users & Roles API

Manage users, roles, and permissions.

## User Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| <span class="api-method get">GET</span> | `/users` | List users |
| <span class="api-method post">POST</span> | `/users` | Create user |
| <span class="api-method get">GET</span> | `/users/{id}` | Get user |
| <span class="api-method put">PUT</span> | `/users/{id}` | Update user |
| <span class="api-method delete">DELETE</span> | `/users/{id}` | Delete user |
| <span class="api-method get">GET</span> | `/users/{id}/sessions` | List sessions |
| <span class="api-method delete">DELETE</span> | `/users/{id}/sessions` | Revoke all sessions |

## Role Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| <span class="api-method get">GET</span> | `/roles` | List roles |
| <span class="api-method post">POST</span> | `/roles` | Create role |
| <span class="api-method get">GET</span> | `/roles/{id}` | Get role |
| <span class="api-method put">PUT</span> | `/roles/{id}` | Update role |
| <span class="api-method delete">DELETE</span> | `/roles/{id}` | Delete role |

## User Object

```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "team_id": "team-uuid",
  "role_id": "role-uuid",
  "auth_provider": "github",
  "status": "active",
  "created_at": "2024-01-10T10:00:00Z",
  "last_login": "2024-01-15T14:30:00Z",
  "team": {
    "id": "team-uuid",
    "name": "Engineering"
  },
  "role": {
    "id": "role-uuid",
    "name": "admin"
  }
}
```

## List Users

<span class="api-method get">GET</span> `/users`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `team_id` | uuid | Filter by team |
| `role_id` | uuid | Filter by role |
| `status` | string | `active` or `inactive` |
| `search` | string | Search by name/email |
| `page` | integer | Page number |
| `per_page` | integer | Items per page |

**Response:**

```json
{
  "data": [
    {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "User Name",
      "status": "active",
      "team": {"name": "Engineering"},
      "role": {"name": "admin"},
      "last_login": "2024-01-15T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 45
  }
}
```

## Create User

<span class="api-method post">POST</span> `/users`

**Request:**

```json
{
  "email": "newuser@example.com",
  "name": "New User",
  "team_id": "team-uuid",
  "role_id": "role-uuid"
}
```

**Response:** `201 Created`

```json
{
  "id": "new-user-uuid",
  "email": "newuser@example.com",
  "name": "New User",
  "team_id": "team-uuid",
  "role_id": "role-uuid",
  "status": "pending",
  "created_at": "2024-01-15T14:30:00Z"
}
```

## Get User

<span class="api-method get">GET</span> `/users/{id}`

**Response:**

```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "team_id": "team-uuid",
  "role_id": "role-uuid",
  "auth_provider": "github",
  "status": "active",
  "created_at": "2024-01-10T10:00:00Z",
  "updated_at": "2024-01-15T12:00:00Z",
  "last_login": "2024-01-15T14:30:00Z",
  "team": {
    "id": "team-uuid",
    "name": "Engineering"
  },
  "role": {
    "id": "role-uuid",
    "name": "admin",
    "permissions": ["manage_rules", "manage_users"]
  },
  "agents": [
    {
      "id": "agent-uuid",
      "hostname": "dev-laptop",
      "last_seen": "2024-01-15T14:25:00Z"
    }
  ]
}
```

## Update User

<span class="api-method put">PUT</span> `/users/{id}`

**Request:**

```json
{
  "name": "Updated Name",
  "team_id": "new-team-uuid",
  "role_id": "new-role-uuid",
  "status": "active"
}
```

## Delete User

<span class="api-method delete">DELETE</span> `/users/{id}`

**Response:** `204 No Content`

## User Sessions

### List Sessions

<span class="api-method get">GET</span> `/users/{id}/sessions`

**Response:**

```json
{
  "data": [
    {
      "id": "session-uuid",
      "type": "browser",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2024-01-15T10:00:00Z",
      "last_active": "2024-01-15T14:30:00Z"
    },
    {
      "id": "session-uuid-2",
      "type": "agent",
      "ip_address": "192.168.1.100",
      "user_agent": "edictflow-agent/1.0.0",
      "hostname": "dev-laptop",
      "created_at": "2024-01-14T08:00:00Z",
      "last_active": "2024-01-15T14:25:00Z"
    }
  ]
}
```

### Revoke Session

<span class="api-method delete">DELETE</span> `/sessions/{id}`

### Revoke All Sessions

<span class="api-method delete">DELETE</span> `/users/{id}/sessions`

---

## Role Object

```json
{
  "id": "role-uuid",
  "name": "admin",
  "description": "Team administrator",
  "permissions": [
    "manage_users",
    "manage_rules",
    "approve_changes",
    "view_audit"
  ],
  "is_system": false,
  "user_count": 5,
  "created_at": "2024-01-10T10:00:00Z"
}
```

## List Roles

<span class="api-method get">GET</span> `/roles`

**Response:**

```json
{
  "data": [
    {
      "id": "role-uuid-1",
      "name": "super_admin",
      "description": "Full system access",
      "is_system": true,
      "user_count": 2
    },
    {
      "id": "role-uuid-2",
      "name": "admin",
      "description": "Team administrator",
      "is_system": true,
      "user_count": 8
    },
    {
      "id": "role-uuid-3",
      "name": "developer",
      "description": "Standard developer",
      "is_system": true,
      "user_count": 35
    }
  ]
}
```

## Create Role

<span class="api-method post">POST</span> `/roles`

**Request:**

```json
{
  "name": "lead_developer",
  "description": "Team lead with rule creation rights",
  "permissions": [
    "view_rules",
    "create_rules",
    "edit_rules",
    "view_changes",
    "request_changes",
    "view_team_users"
  ],
  "parent_role_id": "developer-role-uuid"
}
```

**Response:** `201 Created`

## Get Role

<span class="api-method get">GET</span> `/roles/{id}`

**Response:**

```json
{
  "id": "role-uuid",
  "name": "lead_developer",
  "description": "Team lead with rule creation rights",
  "permissions": [
    "view_rules",
    "create_rules",
    "edit_rules"
  ],
  "inherited_permissions": [
    "view_changes",
    "request_changes"
  ],
  "parent_role": {
    "id": "developer-uuid",
    "name": "developer"
  },
  "user_count": 12,
  "is_system": false,
  "created_at": "2024-01-10T10:00:00Z"
}
```

## Update Role

<span class="api-method put">PUT</span> `/roles/{id}`

**Request:**

```json
{
  "name": "lead_developer",
  "description": "Updated description",
  "permissions": [
    "view_rules",
    "create_rules",
    "edit_rules",
    "delete_rules"
  ]
}
```

## Delete Role

<span class="api-method delete">DELETE</span> `/roles/{id}`

**Errors:**

| Code | Description |
|------|-------------|
| 400 | Role has users assigned |
| 400 | Cannot delete system role |

## Permissions

### List All Permissions

<span class="api-method get">GET</span> `/permissions`

**Response:**

```json
{
  "data": [
    {
      "id": "manage_users",
      "name": "Manage Users",
      "description": "Create, update, and delete users",
      "category": "users"
    },
    {
      "id": "manage_rules",
      "name": "Manage Rules",
      "description": "Full CRUD on rules",
      "category": "rules"
    }
  ]
}
```

### Check Permission

<span class="api-method get">GET</span> `/users/{id}/permissions/{permission}`

**Response:**

```json
{
  "has_permission": true,
  "source": "role",
  "role_name": "admin"
}
```

## Examples

### Create User and Assign Role

```bash
# Create user
curl -X POST "https://api.example.com/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "dev@example.com",
    "name": "Developer",
    "team_id": "team-uuid",
    "role_id": "developer-role-uuid"
  }'
```

### Change User Role

```bash
curl -X PATCH "https://api.example.com/api/v1/users/{id}" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"role_id": "admin-role-uuid"}'
```

### Create Custom Role

```bash
curl -X POST "https://api.example.com/api/v1/roles" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "security_reviewer",
    "description": "Can view and approve security-related changes",
    "permissions": [
      "view_rules",
      "view_changes",
      "approve_changes"
    ]
  }'
```
