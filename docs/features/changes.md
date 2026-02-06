# Change Requests

Edictflow tracks all configuration changes across your organization, providing complete visibility into CLAUDE.md modifications.

## Change Events

Every file modification generates a change event:

| Field | Description |
|-------|-------------|
| `id` | Unique event identifier |
| `rule_id` | Rule that was triggered |
| `agent_id` | Agent that detected the change |
| `user_id` | User who made the change |
| `event_type` | Type of event (blocked, detected, flagged) |
| `old_content` | Previous file content |
| `new_content` | New file content |
| `file_path` | Path to the modified file |
| `created_at` | When the change occurred |

## Event Types

### change_blocked

Generated when block mode reverts a change.

```json
{
  "event_type": "change_blocked",
  "rule_id": "rule-uuid",
  "file_path": "/project/CLAUDE.md",
  "old_content": "# Original content...",
  "new_content": "# User's attempted change...",
  "action_taken": "reverted"
}
```

### change_detected

Generated when temporary mode allows a change.

```json
{
  "event_type": "change_detected",
  "rule_id": "rule-uuid",
  "file_path": "/project/CLAUDE.md",
  "old_content": "# Original content...",
  "new_content": "# User's change...",
  "action_taken": "allowed",
  "requires_approval": true
}
```

### change_flagged

Generated when warning mode logs a change.

```json
{
  "event_type": "change_flagged",
  "rule_id": "rule-uuid",
  "file_path": "/project/CLAUDE.md",
  "old_content": "# Original content...",
  "new_content": "# User's change...",
  "action_taken": "none"
}
```

## Viewing Changes

### Web UI

1. Navigate to **Changes** in the sidebar
2. See list of recent change events
3. Filter by:
   - Team
   - Event type
   - Date range
   - User
   - Agent

### API

```bash
# List recent changes
curl "https://api.example.com/api/v1/changes?team_id=team-uuid" \
  -H "Authorization: Bearer $TOKEN"

# Filter by type
curl "https://api.example.com/api/v1/changes?event_type=change_blocked" \
  -H "Authorization: Bearer $TOKEN"

# Filter by date range
curl "https://api.example.com/api/v1/changes?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"

# Get specific change
curl "https://api.example.com/api/v1/changes/{change_id}" \
  -H "Authorization: Bearer $TOKEN"
```

### Response

```json
{
  "data": [
    {
      "id": "change-uuid",
      "event_type": "change_detected",
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
      },
      "file_path": "/project/CLAUDE.md",
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

## Change Details

### View Diff

Each change includes before/after content:

```bash
curl "https://api.example.com/api/v1/changes/{id}/diff" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "old_content": "# CLAUDE.md\n\nOriginal content...",
  "new_content": "# CLAUDE.md\n\nModified content...",
  "diff": "--- old\n+++ new\n@@ -1,3 +1,3 @@\n # CLAUDE.md\n \n-Original content...\n+Modified content..."
}
```

### Web UI Diff View

The web UI provides a visual diff:

- Side-by-side comparison
- Inline changes highlighted
- Line numbers
- Expandable context

## Change Metrics

### Dashboard Metrics

The dashboard shows:

- Total changes (last 24h, 7d, 30d)
- Changes by type (blocked, detected, flagged)
- Changes by team
- Top changed files

### API Metrics

```bash
curl "https://api.example.com/api/v1/changes/stats?period=7d" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "period": "7d",
  "total": 245,
  "by_type": {
    "change_blocked": 45,
    "change_detected": 120,
    "change_flagged": 80
  },
  "by_team": {
    "Engineering": 180,
    "Platform": 65
  },
  "top_files": [
    {
      "path": "CLAUDE.md",
      "count": 89
    }
  ]
}
```

## Export

### CSV Export

```bash
curl "https://api.example.com/api/v1/changes/export?format=csv&from=2024-01-01" \
  -H "Authorization: Bearer $TOKEN" \
  -o changes.csv
```

### JSON Export

```bash
curl "https://api.example.com/api/v1/changes/export?format=json" \
  -H "Authorization: Bearer $TOKEN" \
  -o changes.json
```

## Real-time Updates

### WebSocket

Subscribe to real-time change events:

```javascript
const ws = new WebSocket('wss://api.example.com/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'changes',
    team_id: 'team-uuid'
  }));
};

ws.onmessage = (event) => {
  const change = JSON.parse(event.data);
  console.log('New change:', change);
};
```

### Webhooks

Configure webhooks for change events:

```bash
curl -X POST "https://api.example.com/api/v1/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "url": "https://your-server.com/webhook",
    "events": ["change_blocked", "change_detected"],
    "secret": "your-webhook-secret"
  }'
```

Webhook payload:

```json
{
  "event": "change_blocked",
  "timestamp": "2024-01-15T14:30:00Z",
  "data": {
    "change_id": "change-uuid",
    "rule_name": "Standard CLAUDE.md",
    "file_path": "/project/CLAUDE.md",
    "user_email": "developer@example.com"
  }
}
```

## Retention

### Default Retention

| Content | Retention |
|---------|-----------|
| Change metadata | 1 year |
| File diffs | 90 days |
| Full content | 30 days |

### Configure Retention

```bash
CHANGES_METADATA_RETENTION=365d
CHANGES_DIFF_RETENTION=90d
CHANGES_CONTENT_RETENTION=30d
```

### Archive

Export before cleanup:

```bash
curl "https://api.example.com/api/v1/changes/archive?older_than=90d" \
  -H "Authorization: Bearer $TOKEN" \
  -o archive.json

# Then cleanup
curl -X DELETE "https://api.example.com/api/v1/changes?older_than=90d" \
  -H "Authorization: Bearer $TOKEN"
```

## Integration

### Slack Notifications

Configure Slack integration:

```bash
curl -X POST "https://api.example.com/api/v1/integrations/slack" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "https://hooks.slack.com/...",
    "events": ["change_blocked"],
    "channel": "#security-alerts"
  }'
```

### JIRA Integration

Create JIRA tickets for changes:

```bash
curl -X POST "https://api.example.com/api/v1/integrations/jira" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "host": "https://company.atlassian.net",
    "project": "SEC",
    "issue_type": "Task",
    "events": ["change_blocked"]
  }'
```

## Best Practices

### 1. Regular Review

Schedule regular review of change logs:

- Daily: Blocked changes (security concern)
- Weekly: All changes (compliance)
- Monthly: Trends and patterns

### 2. Set Up Alerts

Configure alerts for critical events:

- Blocked changes in production
- Unusual change patterns
- Changes from unexpected sources

### 3. Use Filters

Create saved filters for common queries:

- "My team's changes today"
- "All blocked changes this week"
- "Changes to security rules"

### 4. Maintain Archives

Export important changes before retention cleanup:

- Compliance requirements
- Incident investigation
- Audit trails
