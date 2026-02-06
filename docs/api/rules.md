# Rules API

CRUD operations for managing configuration rules.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| <span class="api-method get">GET</span> | `/rules` | List rules |
| <span class="api-method post">POST</span> | `/rules` | Create rule |
| <span class="api-method get">GET</span> | `/rules/{id}` | Get rule |
| <span class="api-method put">PUT</span> | `/rules/{id}` | Update rule |
| <span class="api-method patch">PATCH</span> | `/rules/{id}` | Partial update |
| <span class="api-method delete">DELETE</span> | `/rules/{id}` | Delete rule |
| <span class="api-method get">GET</span> | `/rules/{id}/versions` | List versions |
| <span class="api-method post">POST</span> | `/rules/{id}/rollback` | Rollback version |

## Rule Object

```json
{
  "id": "rule-uuid",
  "name": "Standard CLAUDE.md",
  "team_id": "team-uuid",
  "content": "# CLAUDE.md\n\nProject guidelines...",
  "enforcement_mode": "block",
  "triggers": [
    {"type": "path", "pattern": "CLAUDE.md"},
    {"type": "glob", "pattern": "**/CLAUDE.md"}
  ],
  "description": "Standard configuration for all projects",
  "priority": 100,
  "enabled": true,
  "version": 3,
  "created_at": "2024-01-10T10:00:00Z",
  "updated_at": "2024-01-15T14:30:00Z",
  "created_by": {
    "id": "user-uuid",
    "email": "admin@example.com"
  }
}
```

## List Rules

<span class="api-method get">GET</span> `/rules`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `team_id` | uuid | Filter by team |
| `enabled` | boolean | Filter by enabled status |
| `enforcement_mode` | string | Filter by mode |
| `page` | integer | Page number |
| `per_page` | integer | Items per page |
| `sort` | string | Sort field (e.g., `-created_at`) |

**Response:**

```json
{
  "data": [
    {
      "id": "rule-uuid",
      "name": "Standard CLAUDE.md",
      "team_id": "team-uuid",
      "enforcement_mode": "block",
      "enabled": true,
      "trigger_count": 2,
      "version": 3,
      "created_at": "2024-01-10T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 5
  }
}
```

## Create Rule

<span class="api-method post">POST</span> `/rules`

**Request:**

```json
{
  "name": "Project Guidelines",
  "team_id": "team-uuid",
  "content": "# CLAUDE.md\n\n## Guidelines\n\n- Follow best practices\n- Write tests",
  "enforcement_mode": "block",
  "triggers": [
    {"type": "glob", "pattern": "**/CLAUDE.md"}
  ],
  "description": "Guidelines for all projects",
  "priority": 100
}
```

**Response:** `201 Created`

```json
{
  "id": "new-rule-uuid",
  "name": "Project Guidelines",
  "team_id": "team-uuid",
  "content": "# CLAUDE.md\n\n## Guidelines\n\n- Follow best practices\n- Write tests",
  "enforcement_mode": "block",
  "triggers": [
    {"type": "glob", "pattern": "**/CLAUDE.md"}
  ],
  "enabled": true,
  "version": 1,
  "created_at": "2024-01-15T14:30:00Z"
}
```

**Validation:**

| Field | Requirements |
|-------|--------------|
| `name` | Required, 1-255 characters |
| `team_id` | Required, valid team UUID |
| `content` | Required |
| `enforcement_mode` | Required: `block`, `temporary`, `warning` |
| `triggers` | Required, at least one trigger |

## Get Rule

<span class="api-method get">GET</span> `/rules/{id}`

**Response:**

```json
{
  "id": "rule-uuid",
  "name": "Standard CLAUDE.md",
  "team_id": "team-uuid",
  "content": "# CLAUDE.md\n\nFull content...",
  "enforcement_mode": "block",
  "triggers": [
    {"type": "path", "pattern": "CLAUDE.md"}
  ],
  "description": "Standard configuration",
  "priority": 100,
  "enabled": true,
  "version": 3,
  "created_at": "2024-01-10T10:00:00Z",
  "updated_at": "2024-01-15T14:30:00Z",
  "created_by": {
    "id": "user-uuid",
    "email": "admin@example.com"
  },
  "updated_by": {
    "id": "user-uuid",
    "email": "admin@example.com"
  }
}
```

## Update Rule

<span class="api-method put">PUT</span> `/rules/{id}`

Full update - replaces all fields.

**Request:**

```json
{
  "name": "Updated Rule Name",
  "team_id": "team-uuid",
  "content": "# Updated content...",
  "enforcement_mode": "temporary",
  "triggers": [
    {"type": "glob", "pattern": "**/*.md"}
  ],
  "description": "Updated description",
  "priority": 150,
  "enabled": true
}
```

## Partial Update

<span class="api-method patch">PATCH</span> `/rules/{id}`

Update specific fields only.

**Request:**

```json
{
  "enforcement_mode": "warning"
}
```

**Response:**

```json
{
  "id": "rule-uuid",
  "enforcement_mode": "warning",
  "version": 4,
  "updated_at": "2024-01-15T14:35:00Z"
}
```

## Delete Rule

<span class="api-method delete">DELETE</span> `/rules/{id}`

**Response:** `204 No Content`

**Errors:**

| Code | Description |
|------|-------------|
| 404 | Rule not found |
| 403 | Insufficient permissions |

## Rule Versions

### List Versions

<span class="api-method get">GET</span> `/rules/{id}/versions`

**Response:**

```json
{
  "data": [
    {
      "version": 3,
      "content": "# Latest content...",
      "enforcement_mode": "block",
      "changed_by": {
        "email": "admin@example.com"
      },
      "changed_at": "2024-01-15T14:30:00Z",
      "change_summary": "Updated guidelines section"
    },
    {
      "version": 2,
      "content": "# Previous content...",
      "changed_at": "2024-01-10T10:00:00Z"
    }
  ]
}
```

### Rollback

<span class="api-method post">POST</span> `/rules/{id}/rollback`

**Request:**

```json
{
  "version": 2
}
```

**Response:**

```json
{
  "id": "rule-uuid",
  "content": "# Rolled back content...",
  "version": 4,
  "message": "Rolled back to version 2"
}
```

## Bulk Operations

### Bulk Update Enforcement

<span class="api-method patch">PATCH</span> `/rules/bulk`

**Request:**

```json
{
  "rule_ids": ["uuid-1", "uuid-2", "uuid-3"],
  "enforcement_mode": "block"
}
```

### Bulk Enable/Disable

<span class="api-method patch">PATCH</span> `/rules/bulk`

**Request:**

```json
{
  "rule_ids": ["uuid-1", "uuid-2"],
  "enabled": false
}
```

## Match Rules

Find rules that match a given path.

<span class="api-method get">GET</span> `/rules/match`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `path` | string | File path to match |
| `team_id` | uuid | Team to check |

**Response:**

```json
{
  "matches": [
    {
      "rule_id": "rule-uuid",
      "rule_name": "Standard CLAUDE.md",
      "trigger": {"type": "glob", "pattern": "**/CLAUDE.md"},
      "priority": 100
    }
  ]
}
```

## Examples

### Create Rule with Multiple Triggers

```bash
curl -X POST "https://api.example.com/api/v1/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Multi-path Rule",
    "team_id": "team-uuid",
    "content": "# Configuration\n\nContent here...",
    "enforcement_mode": "block",
    "triggers": [
      {"type": "path", "pattern": "CLAUDE.md"},
      {"type": "path", "pattern": ".claude/config.md"},
      {"type": "glob", "pattern": "docs/**/*.md"}
    ]
  }'
```

### Update Only Enforcement Mode

```bash
curl -X PATCH "https://api.example.com/api/v1/rules/{id}" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"enforcement_mode": "warning"}'
```

### Compare Versions

```bash
curl "https://api.example.com/api/v1/rules/{id}/versions/diff?from=2&to=3" \
  -H "Authorization: Bearer $TOKEN"
```
