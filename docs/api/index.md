# API Reference

Edictflow provides a comprehensive REST API and WebSocket interface for integration and automation.

## Base URL

All API requests use the base URL:

```
https://api.example.com/api/v1
```

## Authentication

All endpoints require authentication via JWT bearer token.

```bash
curl https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer <your-token>"
```

See [Authentication](authentication.md) for details on obtaining tokens.

## Content Type

All requests and responses use JSON:

```bash
curl https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

## API Sections

<div class="grid" markdown>

<div class="card" markdown>

### [Authentication](authentication.md)

Login, tokens, OAuth, and device code flow.

</div>

<div class="card" markdown>

### [Teams](teams.md)

Create and manage teams.

</div>

<div class="card" markdown>

### [Rules](rules.md)

CRUD operations for rules.

</div>

<div class="card" markdown>

### [Changes](changes.md)

Query and export change events.

</div>

<div class="card" markdown>

### [Users & Roles](users-roles.md)

User management and RBAC.

</div>

<div class="card" markdown>

### [WebSocket](websocket.md)

Real-time event streaming.

</div>

</div>

## Response Format

All API responses use a standardized format with `success`, `data`, and `error` fields.

### Success Response

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Example",
    ...
  }
}
```

### List Response

```json
{
  "success": true,
  "data": [
    { "id": "uuid-1", ... },
    { "id": "uuid-2", ... }
  ]
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "rule not found"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | `true` for successful requests, `false` for errors |
| `data` | object/array | Response payload (present on success) |
| `error` | object | Error details (present on failure) |
| `error.code` | string | Machine-readable error code |
| `error.message` | string | Human-readable error message |

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Validation Error |
| 429 | Rate Limited |
| 500 | Server Error |

## Error Codes

| Code | Description |
|------|-------------|
| `BAD_REQUEST` | Invalid request format or parameters |
| `VALIDATION_FAILED` | Request validation failed |
| `UNAUTHORIZED` | Invalid or missing token |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource already exists |
| `INTERNAL_ERROR` | Server error |

## Pagination

List endpoints support pagination:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | 1 | Page number |
| `per_page` | 20 | Items per page (max 100) |

Example:

```bash
curl "https://api.example.com/api/v1/rules?page=2&per_page=50" \
  -H "Authorization: Bearer <token>"
```

## Filtering

Many endpoints support filtering:

```bash
# Filter by team
curl "https://api.example.com/api/v1/rules?team_id=uuid" \
  -H "Authorization: Bearer <token>"

# Filter by date range
curl "https://api.example.com/api/v1/changes?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer <token>"

# Filter by status
curl "https://api.example.com/api/v1/approvals?status=pending" \
  -H "Authorization: Bearer <token>"
```

## Sorting

Use `sort` parameter:

```bash
# Sort by created date descending
curl "https://api.example.com/api/v1/rules?sort=-created_at" \
  -H "Authorization: Bearer <token>"

# Sort by name ascending
curl "https://api.example.com/api/v1/rules?sort=name" \
  -H "Authorization: Bearer <token>"
```

## Rate Limiting

Default limits:

| Tier | Requests | Window |
|------|----------|--------|
| Standard | 100 | 1 minute |
| Burst | 10 | 1 second |

Rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705330800
```

## Versioning

API version is included in the URL path:

```
/api/v1/rules
```

When a new version is released, v1 remains supported for a deprecation period.

## OpenAPI Specification

Download the OpenAPI spec:

```bash
curl https://api.example.com/api/v1/openapi.json -o openapi.json
```

Or view in Swagger UI at `/api/docs`.

## SDKs

Official SDKs:

- **Go**: `go get github.com/kamilrybacki/edictflow/sdk-go`
- **Python**: `pip install edictflow`
- **TypeScript**: `npm install @edictflow/sdk`

Example (Go):

```go
import "github.com/kamilrybacki/edictflow/sdk-go"

client := edictflow.NewClient("https://api.example.com", "token")
rules, err := client.Rules.List(ctx, &edictflow.RulesListParams{
    TeamID: "team-uuid",
})
```

## Webhooks

Configure outbound webhooks for real-time events:

```bash
curl -X POST "https://api.example.com/api/v1/webhooks" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "url": "https://your-server.com/webhook",
    "events": ["change_blocked", "approval_required"],
    "secret": "your-secret"
  }'
```

See [WebSocket](websocket.md) for real-time streaming.
