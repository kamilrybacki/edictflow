# Changes API

Query and manage configuration change events.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| <span class="api-method get">GET</span> | `/changes` | List changes |
| <span class="api-method get">GET</span> | `/changes/{id}` | Get change details |
| <span class="api-method get">GET</span> | `/changes/{id}/diff` | Get change diff |
| <span class="api-method get">GET</span> | `/changes/stats` | Get statistics |
| <span class="api-method get">GET</span> | `/changes/export` | Export changes |

## Change Object

```json
{
  "id": "change-uuid",
  "event_type": "change_blocked",
  "rule_id": "rule-uuid",
  "agent_id": "agent-uuid",
  "user_id": "user-uuid",
  "file_path": "/project/CLAUDE.md",
  "old_content": "# Original...",
  "new_content": "# Modified...",
  "action_taken": "reverted",
  "created_at": "2024-01-15T14:30:00Z",
  "rule": {
    "id": "rule-uuid",
    "name": "Standard CLAUDE.md"
  },
  "agent": {
    "id": "agent-uuid",
    "hostname": "dev-laptop"
  },
  "user": {
    "id": "user-uuid",
    "email": "developer@example.com"
  }
}
```

## Event Types

| Type | Description |
|------|-------------|
| `change_blocked` | Change was reverted (block mode) |
| `change_detected` | Change was allowed (temporary mode) |
| `change_flagged` | Change was logged (warning mode) |
| `sync_complete` | Rule was synced to agent |

## List Changes

<span class="api-method get">GET</span> `/changes`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `team_id` | uuid | Filter by team |
| `event_type` | string | Filter by event type |
| `rule_id` | uuid | Filter by rule |
| `agent_id` | uuid | Filter by agent |
| `user_id` | uuid | Filter by user |
| `from` | datetime | Start date (ISO 8601) |
| `to` | datetime | End date (ISO 8601) |
| `page` | integer | Page number |
| `per_page` | integer | Items per page |
| `sort` | string | Sort field |

**Example:**

```bash
curl "https://api.example.com/api/v1/changes?team_id=team-uuid&event_type=change_blocked&from=2024-01-01" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**

```json
{
  "data": [
    {
      "id": "change-uuid",
      "event_type": "change_blocked",
      "file_path": "/project/CLAUDE.md",
      "rule": {
        "name": "Standard CLAUDE.md"
      },
      "user": {
        "email": "developer@example.com"
      },
      "created_at": "2024-01-15T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 156
  }
}
```

## Get Change Details

<span class="api-method get">GET</span> `/changes/{id}`

**Response:**

```json
{
  "id": "change-uuid",
  "event_type": "change_blocked",
  "rule_id": "rule-uuid",
  "agent_id": "agent-uuid",
  "user_id": "user-uuid",
  "file_path": "/project/CLAUDE.md",
  "old_content": "# CLAUDE.md\n\nOriginal content here...",
  "new_content": "# CLAUDE.md\n\nModified content here...",
  "action_taken": "reverted",
  "created_at": "2024-01-15T14:30:00Z",
  "rule": {
    "id": "rule-uuid",
    "name": "Standard CLAUDE.md",
    "enforcement_mode": "block"
  },
  "agent": {
    "id": "agent-uuid",
    "hostname": "dev-laptop",
    "ip_address": "192.168.1.100"
  },
  "user": {
    "id": "user-uuid",
    "email": "developer@example.com",
    "name": "Developer Name"
  }
}
```

## Get Change Diff

<span class="api-method get">GET</span> `/changes/{id}/diff`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | `unified` (default) or `side_by_side` |
| `context` | integer | Lines of context (default: 3) |

**Response:**

```json
{
  "old_content": "# CLAUDE.md\n\nOriginal content...",
  "new_content": "# CLAUDE.md\n\nModified content...",
  "diff": "--- old\n+++ new\n@@ -1,3 +1,3 @@\n # CLAUDE.md\n \n-Original content...\n+Modified content...",
  "stats": {
    "additions": 5,
    "deletions": 3
  }
}
```

## Statistics

<span class="api-method get">GET</span> `/changes/stats`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `team_id` | uuid | Filter by team |
| `period` | string | `1d`, `7d`, `30d`, `90d` |

**Response:**

```json
{
  "period": "7d",
  "total": 245,
  "by_type": {
    "change_blocked": 45,
    "change_detected": 120,
    "change_flagged": 80
  },
  "by_day": [
    {"date": "2024-01-15", "count": 35},
    {"date": "2024-01-14", "count": 42}
  ],
  "by_team": {
    "Engineering": 180,
    "Platform": 65
  },
  "top_files": [
    {"path": "CLAUDE.md", "count": 89},
    {"path": "src/CLAUDE.md", "count": 34}
  ],
  "top_users": [
    {"email": "dev1@example.com", "count": 45},
    {"email": "dev2@example.com", "count": 32}
  ]
}
```

## Export

<span class="api-method get">GET</span> `/changes/export`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | `json` or `csv` |
| `team_id` | uuid | Filter by team |
| `from` | datetime | Start date |
| `to` | datetime | End date |
| `include_content` | boolean | Include file content |

**Example:**

```bash
curl "https://api.example.com/api/v1/changes/export?format=csv&from=2024-01-01" \
  -H "Authorization: Bearer $TOKEN" \
  -o changes.csv
```

**CSV Format:**

```csv
id,event_type,file_path,rule_name,user_email,created_at
uuid-1,change_blocked,/project/CLAUDE.md,Standard CLAUDE.md,dev@example.com,2024-01-15T14:30:00Z
```

## Examples

### Get Blocked Changes for a Team

```bash
curl "https://api.example.com/api/v1/changes?team_id=team-uuid&event_type=change_blocked&per_page=50" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Changes for a Specific Rule

```bash
curl "https://api.example.com/api/v1/changes?rule_id=rule-uuid" \
  -H "Authorization: Bearer $TOKEN"
```

### Export Last 30 Days

```bash
curl "https://api.example.com/api/v1/changes/export?format=json&from=$(date -d '30 days ago' -I)" \
  -H "Authorization: Bearer $TOKEN" \
  -o changes.json
```

### Get Daily Statistics

```bash
curl "https://api.example.com/api/v1/changes/stats?period=30d&team_id=team-uuid" \
  -H "Authorization: Bearer $TOKEN"
```
