# Audit Logging

Edictflow maintains comprehensive audit logs for compliance, debugging, and security monitoring.

## What's Logged

### User Actions

| Action | Description |
|--------|-------------|
| `user.login` | User authenticated |
| `user.logout` | User logged out |
| `user.created` | New user created |
| `user.updated` | User profile modified |
| `user.deleted` | User removed |
| `user.role_changed` | User's role modified |

### Rule Actions

| Action | Description |
|--------|-------------|
| `rule.created` | New rule created |
| `rule.updated` | Rule content modified |
| `rule.deleted` | Rule removed |
| `rule.enforcement_changed` | Enforcement mode changed |

### Change Events

| Action | Description |
|--------|-------------|
| `change.blocked` | Unauthorized change reverted |
| `change.detected` | Change detected (temporary mode) |
| `change.flagged` | Change flagged (warning mode) |
| `change.approved` | Change request approved |
| `change.rejected` | Change request rejected |

### Agent Actions

| Action | Description |
|--------|-------------|
| `agent.connected` | Agent established connection |
| `agent.disconnected` | Agent connection closed |
| `agent.sync` | Rules synced to agent |

### System Actions

| Action | Description |
|--------|-------------|
| `system.startup` | Server started |
| `system.shutdown` | Server stopped |
| `role.created` | New role created |
| `role.updated` | Role permissions modified |
| `team.created` | New team created |

## Log Structure

Each audit entry contains:

```json
{
  "id": "audit-uuid",
  "timestamp": "2024-01-15T14:30:00Z",
  "action": "rule.updated",
  "actor": {
    "id": "user-uuid",
    "email": "admin@example.com",
    "type": "user"
  },
  "target": {
    "type": "rule",
    "id": "rule-uuid",
    "name": "Standard CLAUDE.md"
  },
  "details": {
    "changes": {
      "enforcement_mode": {
        "old": "warning",
        "new": "block"
      }
    }
  },
  "metadata": {
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "request_id": "req-uuid"
  }
}
```

## Viewing Audit Logs

### Web UI

1. Navigate to **Changes** in the sidebar
2. Use filters to narrow results:
   - Date range
   - Action type
   - User
   - Team
   - Target resource

### API

```bash
# Get recent logs
curl "https://api.example.com/api/v1/audit" \
  -H "Authorization: Bearer $TOKEN"

# Filter by action
curl "https://api.example.com/api/v1/audit?action=rule.updated" \
  -H "Authorization: Bearer $TOKEN"

# Filter by date range
curl "https://api.example.com/api/v1/audit?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"

# Filter by user
curl "https://api.example.com/api/v1/audit?user_id=user-uuid" \
  -H "Authorization: Bearer $TOKEN"
```

### Response

```json
{
  "data": [
    {
      "id": "audit-uuid",
      "timestamp": "2024-01-15T14:30:00Z",
      "action": "rule.updated",
      "actor": { "email": "admin@example.com" },
      "target": { "type": "rule", "name": "Standard CLAUDE.md" }
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 1234
  }
}
```

## Retention

### Default Retention

| Log Type | Retention |
|----------|-----------|
| Security events | 1 year |
| Change events | 6 months |
| Access logs | 3 months |
| Debug logs | 30 days |

### Configure Retention

```bash
# Environment variables
AUDIT_RETENTION_SECURITY=365d
AUDIT_RETENTION_CHANGES=180d
AUDIT_RETENTION_ACCESS=90d
```

### Manual Cleanup

```bash
# Delete logs older than 90 days
curl -X DELETE "https://api.example.com/api/v1/audit?older_than=90d" \
  -H "Authorization: Bearer $TOKEN"
```

## Export

### CSV Export

```bash
curl "https://api.example.com/api/v1/audit/export?format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit_export.csv
```

### JSON Export

```bash
curl "https://api.example.com/api/v1/audit/export?format=json" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit_export.json
```

### Scheduled Exports

Configure automated exports:

```yaml
# In config.yaml
audit:
  export:
    enabled: true
    schedule: "0 2 * * *"  # Daily at 2 AM
    destination: s3://bucket/audit/
    format: json
    retention: 365d
```

## External Integration

### Webhook

Send audit events to external systems:

```bash
curl -X POST "https://api.example.com/api/v1/settings/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://siem.example.com/ingest",
    "events": ["change.blocked", "user.login", "rule.updated"],
    "secret": "webhook-signing-secret"
  }'
```

Webhook payload:

```json
{
  "event": "change.blocked",
  "timestamp": "2024-01-15T14:30:00Z",
  "data": { ... },
  "signature": "sha256=..."
}
```

### SIEM Integration

Forward logs to SIEM systems:

#### Splunk

```yaml
audit:
  splunk:
    enabled: true
    hec_url: https://splunk.example.com:8088
    hec_token: ${SPLUNK_HEC_TOKEN}
    index: edictflow
```

#### Elasticsearch

```yaml
audit:
  elasticsearch:
    enabled: true
    url: https://elastic.example.com:9200
    index: edictflow-audit
    username: ${ES_USER}
    password: ${ES_PASSWORD}
```

#### Datadog

```yaml
audit:
  datadog:
    enabled: true
    api_key: ${DD_API_KEY}
    site: datadoghq.com
```

## Compliance

### SOC 2

Edictflow audit logs support SOC 2 compliance:

- **CC6.1**: Logical access controls are logged
- **CC6.2**: Authentication events are tracked
- **CC7.2**: System events are monitored

### GDPR

For GDPR compliance:

- Audit logs can be exported for data subject requests
- User data can be anonymized in historical logs
- Retention policies can be configured

```bash
# Anonymize user data in logs
curl -X POST "https://api.example.com/api/v1/audit/anonymize" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id": "user-uuid"}'
```

### HIPAA

For HIPAA environments:

- Enable encryption at rest for audit logs
- Configure appropriate retention (6 years)
- Enable detailed access logging

## Monitoring

### Alerts

Configure alerts for critical events:

```yaml
audit:
  alerts:
    - event: user.login_failed
      threshold: 5
      window: 5m
      action: email:security@example.com

    - event: rule.deleted
      threshold: 1
      window: 1m
      action: slack:security-channel
```

### Dashboards

Create dashboards for audit data:

- Login activity trends
- Rule change frequency
- Blocked changes by team
- User activity heatmaps

## Troubleshooting

### Missing Logs

If logs are missing:

1. Check retention policy hasn't deleted them
2. Verify logging is enabled for that event type
3. Check disk space on log storage
4. Review log filtering rules

### Performance Impact

If audit logging affects performance:

1. Enable async logging:
   ```yaml
   audit:
     async: true
     buffer_size: 1000
   ```
2. Reduce log verbosity for non-critical events
3. Use log sampling for high-volume events

### Log Corruption

If logs are corrupted:

1. Stop the server
2. Check database integrity
3. Restore from backup if needed
4. Review disk health
