# Notifications

Stay informed about configuration changes, approvals, and system events through multiple notification channels.

## Notification Types

| Type | Description | Default |
|------|-------------|---------|
| `change_blocked` | A change was blocked and reverted | Email, In-App |
| `change_detected` | A change was detected (requires approval) | In-App |
| `change_flagged` | A change was flagged for review | In-App |
| `approval_required` | You have a pending approval | Email, In-App |
| `approval_result` | Your change was approved/rejected | Email, In-App |
| `rule_updated` | A rule you follow was updated | In-App |
| `agent_disconnected` | Your agent lost connection | Email |
| `system_alert` | System maintenance or issues | Email |

## Channels

### In-App Notifications

Notifications appear in the Web UI:

1. Bell icon in header shows unread count
2. Click to view notification list
3. Click notification to view details
4. Mark as read or dismiss

### Email

Configure email notifications:

```bash
# User preference
curl -X PATCH "https://api.example.com/api/v1/users/{id}/preferences" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "notifications": {
      "email": {
        "enabled": true,
        "types": ["change_blocked", "approval_required", "approval_result"]
      }
    }
  }'
```

### Slack

Configure Slack integration:

```bash
# Organization-wide
curl -X POST "https://api.example.com/api/v1/integrations/slack" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "https://hooks.slack.com/services/...",
    "events": ["change_blocked", "approval_required"],
    "channel": "#edictflow-alerts"
  }'

# Per-team
curl -X POST "https://api.example.com/api/v1/teams/{id}/integrations/slack" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "webhook_url": "https://hooks.slack.com/...",
    "channel": "#team-alerts"
  }'
```

Slack message format:

```
🚫 Change Blocked

Rule: Standard CLAUDE.md
User: developer@example.com
File: /project/CLAUDE.md
Time: 2024-01-15 14:30 UTC

View Details →
```

### Webhooks

Send notifications to custom endpoints:

```bash
curl -X POST "https://api.example.com/api/v1/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "url": "https://your-server.com/notifications",
    "events": ["change_blocked", "change_detected"],
    "secret": "webhook-secret",
    "headers": {
      "X-Custom-Header": "value"
    }
  }'
```

Webhook payload:

```json
{
  "event": "change_blocked",
  "timestamp": "2024-01-15T14:30:00Z",
  "data": {
    "change_id": "change-uuid",
    "rule_id": "rule-uuid",
    "rule_name": "Standard CLAUDE.md",
    "user_email": "developer@example.com",
    "file_path": "/project/CLAUDE.md"
  },
  "signature": "sha256=abc123..."
}
```

Verify webhook signature:

```python
import hmac
import hashlib

def verify_signature(payload, signature, secret):
    expected = 'sha256=' + hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### Desktop (Agent)

Agent can show desktop notifications:

```bash
edictflow-agent config set notifications.desktop true
```

Supported on:

- macOS (Notification Center)
- Linux (libnotify)
- Windows (Toast notifications)

## User Preferences

### View Preferences

```bash
curl "https://api.example.com/api/v1/users/{id}/preferences" \
  -H "Authorization: Bearer $TOKEN"
```

### Update Preferences

```bash
curl -X PATCH "https://api.example.com/api/v1/users/{id}/preferences" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "notifications": {
      "email": {
        "enabled": true,
        "types": ["change_blocked", "approval_result"],
        "digest": "daily"
      },
      "slack": {
        "enabled": true,
        "dm": true
      },
      "inapp": {
        "enabled": true,
        "types": ["all"]
      }
    }
  }'
```

### Digest Mode

Receive batched notifications:

| Digest | Description |
|--------|-------------|
| `immediate` | Send each notification immediately |
| `hourly` | Batch notifications every hour |
| `daily` | Daily summary at 9 AM |
| `weekly` | Weekly summary on Monday |

## Team Settings

Admins configure team-wide notification settings:

```bash
curl -X PATCH "https://api.example.com/api/v1/teams/{id}/settings" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "notifications": {
      "required_types": ["change_blocked"],
      "slack_channel": "#team-alerts",
      "escalation": {
        "pending_approval_hours": 24,
        "escalate_to": ["admin@example.com"]
      }
    }
  }'
```

## Quiet Hours

Configure do-not-disturb periods:

```bash
curl -X PATCH "https://api.example.com/api/v1/users/{id}/preferences" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "notifications": {
      "quiet_hours": {
        "enabled": true,
        "start": "22:00",
        "end": "08:00",
        "timezone": "America/New_York"
      }
    }
  }'
```

During quiet hours:

- Non-critical notifications are queued
- Critical alerts (security) still send
- Queued notifications send after quiet hours

## Notification Templates

Customize notification content:

```bash
curl -X PUT "https://api.example.com/api/v1/settings/templates/change_blocked" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "subject": "[Edictflow] Change Blocked: {{rule_name}}",
    "body": "A change to {{file_path}} was blocked.\n\nRule: {{rule_name}}\nUser: {{user_email}}\nTime: {{timestamp}}\n\nView details: {{link}}"
  }'
```

Available variables:

| Variable | Description |
|----------|-------------|
| `{{rule_name}}` | Name of the triggered rule |
| `{{user_email}}` | User who made the change |
| `{{file_path}}` | Path to the modified file |
| `{{timestamp}}` | When the change occurred |
| `{{link}}` | Link to change details |
| `{{team_name}}` | Team name |

## Testing

### Test Email

```bash
curl -X POST "https://api.example.com/api/v1/notifications/test/email" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": "test@example.com"}'
```

### Test Slack

```bash
curl -X POST "https://api.example.com/api/v1/notifications/test/slack" \
  -H "Authorization: Bearer $TOKEN"
```

### Test Webhook

```bash
curl -X POST "https://api.example.com/api/v1/webhooks/{id}/test" \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### Email Not Received

1. Check spam folder
2. Verify email address in profile
3. Check notification preferences
4. Review SMTP configuration (admin)

### Slack Not Working

1. Verify webhook URL is correct
2. Check Slack app permissions
3. Ensure channel exists
4. Test webhook directly

### Webhook Failures

View webhook logs:

```bash
curl "https://api.example.com/api/v1/webhooks/{id}/logs" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "logs": [
    {
      "timestamp": "2024-01-15T14:30:00Z",
      "status": "failed",
      "status_code": 500,
      "error": "Connection timeout"
    }
  ]
}
```

## Best Practices

### 1. Configure Thoughtfully

Don't over-notify:

- Critical: Email + Slack
- Important: Slack or In-App
- Informational: In-App only

### 2. Use Digest Mode

For high-volume teams, use daily digests to reduce noise.

### 3. Set Up Escalation

Ensure critical items don't get missed:

- Pending approvals → escalate after 24h
- Blocked changes → immediate alert to security

### 4. Regular Review

Periodically review notification settings:

- Are you getting too many notifications?
- Are important ones being missed?
- Are the right people receiving alerts?
