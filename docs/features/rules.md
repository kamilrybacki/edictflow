# Rules

Rules are the core building blocks of Edictflow. Each rule defines the content, enforcement policy, and scope for CLAUDE.md configurations.

## Three-File Model

Edictflow manages exactly three CLAUDE.md files at fixed locations:

| Target Layer | File Path | Description |
|--------------|-----------|-------------|
| **Enterprise** | `/etc/claude-code/CLAUDE.md` | Organization-wide policies |
| **User** | `~/.claude/CLAUDE.md` | User-specific preferences |
| **Project** | `./CLAUDE.md` | Project-specific guidelines |

The agent merges rules from all layers into these fixed files, maintaining a managed section for Edictflow content while preserving any manually-added content outside the managed section.

### Managed Sections

Each CLAUDE.md file contains a managed section delimited by special markers:

```markdown
# My Project CLAUDE.md

Some manual content I added...

<!-- MANAGED BY EDICTFLOW - DO NOT EDIT -->

## Security

[Enterprise] **No Hardcoded Secrets**
Never commit API keys or secrets to the repository.

## Coding Standards

[Enterprise] **TypeScript Strict Mode** (overridable)
Use TypeScript strict mode in all projects.

<!-- END EDICTFLOW -->

More manual content...
```

Content outside the managed section is preserved during updates. If the managed section is tampered with, Edictflow will restore it and notify administrators.

## Rule Structure

A rule consists of:

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier |
| `name` | String | Human-readable name |
| `description` | String | Optional detailed description |
| `team_id` | UUID | Team this rule belongs to |
| `category_id` | UUID | Category for grouping (Security, Testing, etc.) |
| `content` | Text | The rule content to enforce |
| `target_layer` | Enum | Where to apply: enterprise, user, project |
| `enforcement_mode` | Enum | How to enforce: block, temporary, warning |
| `overridable` | Boolean | Can lower layers override this rule |
| `effective_start` | Timestamp | When the rule becomes active (optional) |
| `effective_end` | Timestamp | When the rule expires (optional) |
| `target_teams` | Array | Specific teams to target (optional) |
| `target_users` | Array | Specific users to target (optional) |
| `tags` | Array | Tags for filtering and organization |
| `triggers` | Array | Files and patterns this rule applies to |
| `priority_weight` | Integer | Priority when multiple rules match |
| `status` | Enum | draft, pending, approved, rejected |
| `created_at` | Timestamp | When the rule was created |
| `updated_at` | Timestamp | When the rule was last modified |

## Categories

Rules are organized into categories for better management and display. Categories appear as sections in the merged CLAUDE.md files.

### Default Categories

| Category | Description |
|----------|-------------|
| **Security** | Security policies and requirements |
| **Coding Standards** | Code style and quality guidelines |
| **Testing** | Testing requirements and practices |
| **Documentation** | Documentation standards |

### Custom Categories

Teams can create custom categories:

```bash
curl -X POST https://api.example.com/api/v1/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Performance",
    "display_order": 5
  }'
```

## Creating a Rule

### Via Web UI

1. Navigate to **Rules** in the sidebar
2. Click **Create Rule**
3. Fill in the form:
   - **Name**: Descriptive name
   - **Description**: Optional details
   - **Team**: Select the team
   - **Target Layer**: Enterprise, User, or Project
   - **Category**: Select a category
   - **Content**: The rule content
   - **Overridable**: Allow lower layers to override
   - **Effective Dates**: Optional start/end dates
   - **Tags**: Add tags for organization
   - **Enforcement Mode**: Block, Temporary, or Warning
   - **Triggers**: Add paths or patterns
4. Click **Create**

### Via API

```bash
curl -X POST https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "No Hardcoded Secrets",
    "description": "Prevent API keys and secrets from being committed",
    "team_id": "team-uuid",
    "category_id": "security-category-uuid",
    "target_layer": "enterprise",
    "content": "Never commit API keys, passwords, or other secrets to the repository. Use environment variables or secret management tools.",
    "overridable": false,
    "enforcement_mode": "block",
    "triggers": [
      {"type": "path", "pattern": "*.go"},
      {"type": "path", "pattern": "*.ts"}
    ],
    "tags": ["security", "secrets"]
  }'
```

## Target Layers

### Enterprise Layer

Enterprise rules apply to all users and projects in the organization.

- Stored at `/etc/claude-code/CLAUDE.md`
- Managed by administrators
- Typically contain organization-wide policies

### User Layer

User rules apply to all projects for a specific user.

- Stored at `~/.claude/CLAUDE.md`
- User-specific preferences and overrides
- Can override enterprise rules marked as overridable

### Project Layer

Project rules apply to a specific project directory.

- Stored at `./CLAUDE.md` in the project root
- Project-specific guidelines
- Can override user and enterprise rules marked as overridable

## Overridable Rules

Rules can be marked as `overridable: true` to allow lower layers to override them.

### Override Hierarchy

1. **Enterprise** rules set the baseline
2. **User** rules can override enterprise rules (if overridable)
3. **Project** rules can override user and enterprise rules (if overridable)

### Example

Enterprise rule (overridable):
```
Use 2 spaces for indentation.
```

Project rule override:
```
Use 4 spaces for indentation (overrides enterprise policy).
```

## Effective Dates

Rules can have optional start and end dates:

```bash
curl -X POST https://api.example.com/api/v1/rules \
  -d '{
    "name": "Holiday Code Freeze",
    "content": "No deployments during the holiday period.",
    "effective_start": "2024-12-20T00:00:00Z",
    "effective_end": "2025-01-02T00:00:00Z"
  }'
```

Rules outside their effective date range are not included in merged content.

## Getting Merged Content

Retrieve the merged CLAUDE.md content for any target layer:

```bash
# Get enterprise-level merged content
curl "https://api.example.com/api/v1/rules/merged?level=enterprise" \
  -H "Authorization: Bearer $TOKEN"

# Get user-level merged content
curl "https://api.example.com/api/v1/rules/merged?level=user" \
  -H "Authorization: Bearer $TOKEN"

# Get project-level merged content
curl "https://api.example.com/api/v1/rules/merged?level=project" \
  -H "Authorization: Bearer $TOKEN"
```

## Rule Priority

When multiple rules match the same layer and category, priority determines order:

- **Higher priority_weight wins** (100 > 50)
- **Equal priority**: More recently updated wins

## Rule Approval Workflow

Rules go through an approval workflow before becoming active:

1. **Draft**: Initial creation, can be edited
2. **Pending**: Submitted for approval
3. **Approved**: Active and enforced
4. **Rejected**: Returned with feedback

See [Approvals](approvals.md) for details.

## Tampering Detection

Edictflow monitors managed sections for unauthorized changes:

1. Agent periodically checks file contents
2. If managed section differs from expected, it's restored
3. Administrator is notified of tampering
4. Audit log records the incident

## Best Practices

### 1. Use Categories

Organize rules into logical categories for better management.

### 2. Set Appropriate Layers

- **Enterprise**: Organization-wide policies (security, compliance)
- **User**: Personal preferences (editor settings, shortcuts)
- **Project**: Project-specific guidelines (dependencies, architecture)

### 3. Use Overridable Wisely

Mark rules as overridable when flexibility is appropriate:
- ✅ Coding style preferences
- ❌ Security requirements

### 4. Set Effective Dates

Use effective dates for temporary policies like code freezes.

### 5. Document with Descriptions

Add clear descriptions to help users understand rule purposes.

### 6. Use Tags

Add tags for filtering and organization:
- `security`, `compliance`, `style`, `testing`, `documentation`
