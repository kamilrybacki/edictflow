# Teams API

Manage teams in Edictflow.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| <span class="api-method get">GET</span> | `/teams` | List teams |
| <span class="api-method post">POST</span> | `/teams` | Create team |
| <span class="api-method get">GET</span> | `/teams/{id}` | Get team |
| <span class="api-method put">PUT</span> | `/teams/{id}` | Update team |
| <span class="api-method delete">DELETE</span> | `/teams/{id}` | Delete team |
| <span class="api-method get">GET</span> | `/teams/{id}/members` | List members |
| <span class="api-method post">POST</span> | `/teams/{id}/members` | Add member |
| <span class="api-method delete">DELETE</span> | `/teams/{id}/members/{user_id}` | Remove member |

## List Teams

<span class="api-method get">GET</span> `/teams`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `per_page` | integer | Items per page |

**Response:**

```json
{
  "data": [
    {
      "id": "team-uuid",
      "name": "Engineering",
      "description": "Engineering team",
      "member_count": 15,
      "rule_count": 5,
      "created_at": "2024-01-10T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 3
  }
}
```

## Create Team

<span class="api-method post">POST</span> `/teams`

**Request:**

```json
{
  "name": "Platform",
  "description": "Platform engineering team"
}
```

**Response:**

```json
{
  "id": "team-uuid",
  "name": "Platform",
  "description": "Platform engineering team",
  "member_count": 0,
  "rule_count": 0,
  "created_at": "2024-01-15T14:30:00Z"
}
```

**Errors:**

| Code | Description |
|------|-------------|
| 400 | Invalid request body |
| 409 | Team name already exists |

## Get Team

<span class="api-method get">GET</span> `/teams/{id}`

**Response:**

```json
{
  "id": "team-uuid",
  "name": "Engineering",
  "description": "Engineering team",
  "member_count": 15,
  "rule_count": 5,
  "settings": {
    "default_enforcement": "block",
    "require_approval": true
  },
  "created_at": "2024-01-10T10:00:00Z",
  "updated_at": "2024-01-15T12:00:00Z"
}
```

## Update Team

<span class="api-method put">PUT</span> `/teams/{id}`

**Request:**

```json
{
  "name": "Engineering Team",
  "description": "Updated description",
  "settings": {
    "default_enforcement": "temporary",
    "require_approval": false
  }
}
```

**Response:**

```json
{
  "id": "team-uuid",
  "name": "Engineering Team",
  "description": "Updated description",
  "settings": {
    "default_enforcement": "temporary",
    "require_approval": false
  },
  "updated_at": "2024-01-15T14:30:00Z"
}
```

## Delete Team

<span class="api-method delete">DELETE</span> `/teams/{id}`

**Response:** `204 No Content`

**Errors:**

| Code | Description |
|------|-------------|
| 400 | Team has members or rules |
| 404 | Team not found |

!!! warning
    Cannot delete teams with active members or rules. Remove them first.

## List Members

<span class="api-method get">GET</span> `/teams/{id}/members`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number |
| `per_page` | integer | Items per page |
| `role` | string | Filter by role |

**Response:**

```json
{
  "data": [
    {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "User Name",
      "role": {
        "id": "role-uuid",
        "name": "admin"
      },
      "joined_at": "2024-01-10T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 15
  }
}
```

## Add Member

<span class="api-method post">POST</span> `/teams/{id}/members`

**Request:**

```json
{
  "user_id": "user-uuid",
  "role_id": "role-uuid"
}
```

**Response:**

```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "role": {
    "id": "role-uuid",
    "name": "developer"
  },
  "joined_at": "2024-01-15T14:30:00Z"
}
```

**Errors:**

| Code | Description |
|------|-------------|
| 400 | Invalid user or role |
| 409 | User already a member |

## Remove Member

<span class="api-method delete">DELETE</span> `/teams/{id}/members/{user_id}`

**Response:** `204 No Content`

## Team Settings

Teams can have custom settings:

| Setting | Type | Description |
|---------|------|-------------|
| `default_enforcement` | string | Default enforcement mode for new rules |
| `require_approval` | boolean | Require approval for changes |
| `notification_email` | string | Team notification email |
| `slack_webhook` | string | Slack webhook URL |

## Examples

### Create Team with Settings

```bash
curl -X POST "https://api.example.com/api/v1/teams" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Security",
    "description": "Security team",
    "settings": {
      "default_enforcement": "block",
      "require_approval": true
    }
  }'
```

### Add Multiple Members

```bash
curl -X POST "https://api.example.com/api/v1/teams/{id}/members/bulk" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "members": [
      {"user_id": "uuid-1", "role_id": "developer-role"},
      {"user_id": "uuid-2", "role_id": "developer-role"},
      {"user_id": "uuid-3", "role_id": "admin-role"}
    ]
  }'
```

### Get Team Statistics

```bash
curl "https://api.example.com/api/v1/teams/{id}/stats" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "member_count": 15,
  "rule_count": 5,
  "agent_count": 12,
  "changes_last_7d": 45,
  "blocked_last_7d": 3
}
```
