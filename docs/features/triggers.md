# Triggers

Triggers define which files a rule applies to. Each rule can have multiple triggers using paths, patterns, or glob expressions.

## Trigger Types

| Type | Description | Example |
|------|-------------|---------|
| `path` | Exact file path | `CLAUDE.md` |
| `glob` | Glob pattern matching | `**/CLAUDE.md` |

## Path Triggers

Match exact file paths relative to the workspace root.

### Syntax

```json
{
  "type": "path",
  "pattern": "CLAUDE.md"
}
```

### Examples

| Pattern | Matches |
|---------|---------|
| `CLAUDE.md` | `/project/CLAUDE.md` |
| `src/CLAUDE.md` | `/project/src/CLAUDE.md` |
| `.claude/config.md` | `/project/.claude/config.md` |

### Use Cases

- Single file enforcement
- Specific file locations
- Exact path matching

## Glob Triggers

Match files using glob patterns.

### Syntax

```json
{
  "type": "glob",
  "pattern": "**/CLAUDE.md"
}
```

### Glob Patterns

| Pattern | Meaning |
|---------|---------|
| `*` | Match any characters except `/` |
| `**` | Match any characters including `/` |
| `?` | Match single character |
| `[abc]` | Match character in set |
| `[!abc]` | Match character not in set |

### Examples

| Pattern | Matches |
|---------|---------|
| `**/CLAUDE.md` | Any `CLAUDE.md` in any directory |
| `src/**/*.md` | Any `.md` file under `src/` |
| `CLAUDE*.md` | `CLAUDE.md`, `CLAUDE-dev.md`, etc. |
| `**/[Cc]laude*.md` | Case-insensitive Claude files |

### Use Cases

- Recursive matching across directories
- Pattern-based file selection
- Multiple file types

## Configuring Triggers

### Via Web UI

1. Open rule editor
2. Scroll to **Triggers** section
3. Click **Add Trigger**
4. Select type and enter pattern
5. Click **Save**

### Via API

```bash
curl -X POST https://api.example.com/api/v1/rules \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Multi-trigger Rule",
    "team_id": "team-uuid",
    "content": "...",
    "enforcement_mode": "block",
    "triggers": [
      {"type": "path", "pattern": "CLAUDE.md"},
      {"type": "glob", "pattern": "**/CLAUDE.md"},
      {"type": "glob", "pattern": ".claude/*.md"}
    ]
  }'
```

### Update Triggers

```bash
curl -X PATCH https://api.example.com/api/v1/rules/{id} \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "triggers": [
      {"type": "glob", "pattern": "**/CLAUDE.md"}
    ]
  }'
```

## Multiple Triggers

Rules can have multiple triggers. A file matches if it matches **any** trigger.

### Example

```json
{
  "triggers": [
    {"type": "path", "pattern": "CLAUDE.md"},
    {"type": "path", "pattern": "GUIDELINES.md"},
    {"type": "glob", "pattern": "docs/**/*.md"}
  ]
}
```

This rule applies to:

- `CLAUDE.md` in root
- `GUIDELINES.md` in root
- Any `.md` file under `docs/`

## Trigger Priority

When multiple rules match the same file:

1. **Explicit path** takes precedence over glob
2. **More specific glob** wins over general
3. **Rule priority** breaks ties

### Example

| Rule | Trigger | Priority |
|------|---------|----------|
| Rule A | `src/CLAUDE.md` (path) | 50 |
| Rule B | `**/CLAUDE.md` (glob) | 100 |

For `/project/src/CLAUDE.md`:

- Both rules match
- Rule A wins (path > glob, despite lower priority)

## Exclusions

Exclude files from matching:

### Exclude Pattern

```json
{
  "triggers": [
    {"type": "glob", "pattern": "**/CLAUDE.md"}
  ],
  "exclusions": [
    {"type": "glob", "pattern": "**/test/**"}
  ]
}
```

This matches all `CLAUDE.md` files except those in `test/` directories.

### Common Exclusions

```json
{
  "exclusions": [
    {"type": "glob", "pattern": "**/node_modules/**"},
    {"type": "glob", "pattern": "**/vendor/**"},
    {"type": "glob", "pattern": "**/.git/**"},
    {"type": "glob", "pattern": "**/dist/**"}
  ]
}
```

## Workspace Configuration

Triggers are relative to the user's workspace root.

### How Root is Determined

1. Git repository root (if in git repo)
2. Current working directory
3. Home directory (fallback)

### Agent Configuration

Override workspace root:

```bash
edictflow-agent config set workspace.root /custom/path
```

## Testing Triggers

### Via CLI

```bash
# Check which rules match a file
edictflow-agent validate /path/to/CLAUDE.md
```

### Via API

```bash
curl "https://api.example.com/api/v1/rules/match?path=src/CLAUDE.md&team_id=team-uuid" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "matches": [
    {
      "rule_id": "rule-uuid",
      "rule_name": "Source Rules",
      "trigger": {"type": "glob", "pattern": "src/**/*.md"}
    }
  ]
}
```

## Common Patterns

### All CLAUDE.md Files

```json
{"type": "glob", "pattern": "**/CLAUDE.md"}
```

### Root CLAUDE.md Only

```json
{"type": "path", "pattern": "CLAUDE.md"}
```

### All Markdown in Directory

```json
{"type": "glob", "pattern": "docs/**/*.md"}
```

### Multiple Config Files

```json
{
  "triggers": [
    {"type": "path", "pattern": "CLAUDE.md"},
    {"type": "path", "pattern": ".claude/config.md"},
    {"type": "path", "pattern": ".claude/guidelines.md"}
  ]
}
```

### Case-Insensitive

```json
{"type": "glob", "pattern": "**/[Cc][Ll][Aa][Uu][Dd][Ee].md"}
```

Or use the case-insensitive flag (if supported):

```json
{"type": "glob", "pattern": "**/claude.md", "case_insensitive": true}
```

## Best Practices

### 1. Be Specific

Prefer specific patterns over broad ones:

```json
// Good - specific
{"type": "path", "pattern": "CLAUDE.md"}

// Less good - too broad
{"type": "glob", "pattern": "**/*.md"}
```

### 2. Test Before Deploying

Verify triggers match expected files:

```bash
edictflow-agent validate --dry-run
```

### 3. Document Trigger Intent

Include descriptions:

```json
{
  "triggers": [
    {
      "type": "glob",
      "pattern": "src/**/CLAUDE.md",
      "description": "Source code guidelines"
    }
  ]
}
```

### 4. Use Exclusions Sparingly

Exclusions add complexity. Consider separate rules instead.

### 5. Consider Performance

Very broad patterns (like `**/*`) can impact file watching performance.

## Troubleshooting

### File Not Being Watched

1. Check trigger matches: `edictflow-agent validate <path>`
2. Verify rule is enabled
3. Check rule applies to your team
4. Force sync: `edictflow-agent sync --force`

### Wrong Rule Applied

1. Check all matching rules: API match endpoint
2. Review trigger specificity
3. Check rule priorities
4. Look for path vs glob precedence

### Pattern Not Matching

1. Verify syntax (glob vs path)
2. Check path is relative to workspace root
3. Test pattern online (globster.xyz)
4. Check for case sensitivity
