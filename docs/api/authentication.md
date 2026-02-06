# Authentication

Edictflow supports multiple authentication methods for the API and agent CLI.

## Overview

| Method | Use Case |
|--------|----------|
| JWT Token | API requests |
| OAuth 2.0 | Web UI login |
| Device Code | Agent CLI login |
| API Keys | Automation/CI |

## JWT Tokens

All API requests require a JWT bearer token.

### Using Tokens

```bash
curl https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer <your-token>"
```

### Token Structure

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "team_id": "team-uuid",
  "role": "admin",
  "exp": 1705417200,
  "iat": 1705330800
}
```

### Token Expiration

| Token Type | Expiration |
|------------|------------|
| Access Token | 24 hours |
| Refresh Token | 7 days |

## Login (Password)

For local authentication:

<span class="api-method post">POST</span> `/auth/login`

**Request:**

```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**

```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 86400,
  "token_type": "Bearer",
  "user": {
    "id": "user-uuid",
    "email": "user@example.com",
    "name": "User Name"
  }
}
```

## Refresh Token

<span class="api-method post">POST</span> `/auth/refresh`

**Request:**

```json
{
  "refresh_token": "eyJhbG..."
}
```

**Response:**

```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 86400
}
```

## OAuth 2.0

### Supported Providers

- GitHub
- Google
- Custom OIDC

### OAuth Flow

1. **Initiate**: Redirect to `/auth/{provider}`
2. **Callback**: Provider redirects to `/auth/{provider}/callback`
3. **Token**: Server issues JWT tokens

### GitHub OAuth

<span class="api-method get">GET</span> `/auth/github`

Redirects to GitHub for authorization.

<span class="api-method get">GET</span> `/auth/github/callback`

Handles callback and issues tokens.

### Google OAuth

<span class="api-method get">GET</span> `/auth/google`

Redirects to Google for authorization.

## Device Code Flow

For CLI authentication without a browser on the same device.

### Step 1: Request Device Code

<span class="api-method post">POST</span> `/auth/device/code`

**Request:**

```json
{
  "client_id": "edictflow-agent"
}
```

**Response:**

```json
{
  "device_code": "ABCD-EFGH-IJKL",
  "user_code": "ABC-123",
  "verification_uri": "https://app.example.com/device",
  "verification_uri_complete": "https://app.example.com/device?code=ABC-123",
  "expires_in": 1800,
  "interval": 5
}
```

### Step 2: User Verification

User visits `verification_uri` and enters `user_code`.

### Step 3: Poll for Token

<span class="api-method post">POST</span> `/auth/device/token`

**Request:**

```json
{
  "client_id": "edictflow-agent",
  "device_code": "ABCD-EFGH-IJKL",
  "grant_type": "urn:ietf:params:oauth:grant-type:device_code"
}
```

**Response (pending):**

```json
{
  "error": "authorization_pending"
}
```

**Response (success):**

```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 86400,
  "token_type": "Bearer"
}
```

**Response (denied):**

```json
{
  "error": "access_denied"
}
```

## API Keys

For automation and CI/CD.

### Create API Key

<span class="api-method post">POST</span> `/users/{id}/api-keys`

**Request:**

```json
{
  "name": "CI/CD Pipeline",
  "expires_at": "2025-01-15T00:00:00Z",
  "permissions": ["read:rules", "write:changes"]
}
```

**Response:**

```json
{
  "id": "key-uuid",
  "key": "ccp_xxxxxxxxxxxxxxxxxxxx",
  "name": "CI/CD Pipeline",
  "created_at": "2024-01-15T10:00:00Z",
  "expires_at": "2025-01-15T00:00:00Z"
}
```

!!! warning "Store Securely"
    The API key is only shown once. Store it securely.

### Using API Keys

```bash
curl https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer ccp_xxxxxxxxxxxx"
```

### List API Keys

<span class="api-method get">GET</span> `/users/{id}/api-keys`

### Revoke API Key

<span class="api-method delete">DELETE</span> `/api-keys/{id}`

## Logout

<span class="api-method post">POST</span> `/auth/logout`

Revokes the current token.

**Request:**

```json
{
  "refresh_token": "eyJhbG..."
}
```

## Current User

<span class="api-method get">GET</span> `/auth/me`

Get information about the authenticated user.

**Response:**

```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "team": {
    "id": "team-uuid",
    "name": "Engineering"
  },
  "role": {
    "id": "role-uuid",
    "name": "admin"
  },
  "permissions": ["manage_rules", "manage_users", ...]
}
```

## Security Best Practices

### Token Storage

| Environment | Recommendation |
|-------------|----------------|
| Browser | HTTP-only cookies |
| Mobile | Secure storage |
| CLI | Keychain/credential manager |
| CI/CD | Environment variables/secrets |

### Token Rotation

- Rotate API keys regularly
- Use short-lived tokens when possible
- Revoke tokens on logout

### Rate Limiting

Auth endpoints have stricter rate limits:

| Endpoint | Limit |
|----------|-------|
| `/auth/login` | 10/minute |
| `/auth/device/*` | 30/minute |
| Token refresh | 60/minute |

## Errors

| Error | Description |
|-------|-------------|
| `invalid_credentials` | Wrong email/password |
| `invalid_token` | Token is malformed |
| `expired_token` | Token has expired |
| `revoked_token` | Token was revoked |
| `authorization_pending` | Device code not yet authorized |
| `expired_token` | Device code expired |
| `access_denied` | User denied authorization |
