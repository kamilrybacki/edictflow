# Web UI

The Edictflow Web UI provides a visual interface for managing rules, users, teams, and monitoring the system. Both administrators and regular users can access the Web UI — available features depend on your assigned role and permissions.

## Accessing the UI

After deployment, access the Web UI at:

- **Development**: `http://localhost:3000`
- **Production**: `https://app.yourdomain.com`

## Dashboard

The dashboard provides an overview of system status:

- **Connected Agents** - Number of agents currently connected
- **Active Rules** - Total rules across all teams
- **Recent Changes** - Latest configuration change events
- **System Health** - Server and database status

## Navigation

The sidebar provides access to all administrative functions:

| Section | Purpose |
|---------|---------|
| Dashboard | System overview |
| Graph | Organization hierarchy visualization |
| Teams | Manage teams |
| Users | Manage users and roles |
| Rules | Create and edit rules |
| Agents | View connected agents |
| Changes | Audit log of changes |
| Approvals | Pending change approvals |
| Settings | System configuration |

### Command Palette

Access the command palette for quick navigation:

1. Press `Ctrl+K` (Windows/Linux) or `Cmd+K` (macOS)
2. Type to search for pages, teams, rules, or actions
3. Press `Enter` to navigate or execute

The command palette provides:

- **Page Navigation** - Jump to any section instantly
- **Team Search** - Find and navigate to specific teams
- **Rule Search** - Locate rules by name
- **Quick Actions** - Common administrative tasks

## Graph View

The Graph View (`/graph`) provides an interactive visualization of your organization's hierarchy.

### Accessing Graph View

1. Click **Graph** in the sidebar
2. Or use the command palette (`Ctrl+K`) and type "graph"

### Graph Features

| Feature | Description |
|---------|-------------|
| **Node Types** | Teams (blue), Users (green), Rules (orange) |
| **Edges** | Connections showing relationships |
| **Zoom** | Mouse wheel or pinch to zoom |
| **Pan** | Click and drag to move around |
| **Fit View** | Click the fit button to center all nodes |

### Node Interactions

- **Click** a node to select and highlight connections
- **Hover** to see basic information
- **Double-click** to navigate to the detail page

### Filtering

Filter the graph to focus on specific elements:

| Filter | Description |
|--------|-------------|
| **Team** | Show only nodes related to a specific team |
| **Status** | Filter rules by status (Draft, Pending, Approved) |
| **Search** | Find nodes by name |

### Graph Controls

The control panel provides:

- **Zoom In/Out** - Adjust zoom level
- **Fit View** - Center and fit all nodes
- **Reset** - Return to default view
- **Fullscreen** - Expand to full screen

### Hierarchical Layout

The graph uses a hierarchical layout:

```
       Teams (Top)
          |
       Users (Middle)
          |
       Rules (Bottom)
```

Rules are connected to their target teams and users, showing the enforcement scope visually.

## Teams

### Create Team

1. Navigate to **Teams**
2. Click **Create Team**
3. Enter team details:
   - **Name**: Display name for the team
   - **Description**: Optional description
4. Click **Create**

### Team Management

From the team detail page:

- **Members**: View and manage team members
- **Rules**: View rules assigned to this team
- **Agents**: View connected agents
- **Settings**: Configure team settings

## Users

### Invite User

1. Navigate to **Users**
2. Click **Invite User**
3. Enter user details:
   - **Email**: User's email address
   - **Team**: Assign to a team
   - **Role**: Select a role
4. Click **Send Invitation**

### User Roles

Assign roles from the user detail page:

- Click the user's current role
- Select a new role from the dropdown
- Changes apply immediately

### Bulk Operations

Select multiple users for bulk actions:

- **Change Role**: Assign role to selected users
- **Change Team**: Move users to different team
- **Deactivate**: Disable selected accounts

## Rules

### Create Rule

1. Navigate to **Rules**
2. Click **Create Rule**
3. Configure the rule:

#### Basic Settings

| Field | Description |
|-------|-------------|
| Name | Descriptive name for the rule |
| Team | Team this rule applies to |
| Description | What this rule enforces |

#### Content

The CLAUDE.md content to enforce. Use the Monaco editor for syntax highlighting:

```markdown
# CLAUDE.md

## Project Guidelines

- Follow TypeScript best practices
- Write tests for all new features
- Keep functions under 50 lines
```

#### Enforcement Mode

| Mode | Behavior |
|------|----------|
| Block | Revert unauthorized changes immediately |
| Temporary | Allow changes but flag for review |
| Warning | Log changes without intervention |

#### Triggers

Define what files this rule applies to:

| Trigger Type | Example |
|--------------|---------|
| Path | `CLAUDE.md` |
| Pattern | `**/CLAUDE.md` |
| Directory | `src/` |

### Edit Rule

1. Navigate to **Rules**
2. Click on the rule name
3. Make changes
4. Click **Save**

Changes are pushed to connected agents immediately.

### Rule History

View the change history for a rule:

1. Open the rule detail page
2. Click **History** tab
3. See all modifications with:
   - Timestamp
   - User who made the change
   - Before/after diff

## Agents

### View Connected Agents

The Agents page shows all registered agents:

| Column | Description |
|--------|-------------|
| Hostname | Agent's machine hostname |
| User | User who owns the agent |
| Team | Team the agent belongs to |
| Status | Online/Offline indicator |
| Last Seen | Last heartbeat time |
| Version | Agent version |

### Agent Details

Click an agent to see:

- Connection history
- Applied rules
- Recent change events
- Sync status

### Disconnect Agent

To forcibly disconnect an agent:

1. Open agent details
2. Click **Disconnect**
3. Confirm the action

The agent will attempt to reconnect unless deauthorized.

## Changes

The Changes page shows all configuration change events:

### Event Types

| Type | Description | Icon |
|------|-------------|------|
| `change_blocked` | Change was reverted | :material-block-helper: |
| `change_detected` | Change was detected (temporary mode) | :material-eye: |
| `change_flagged` | Change was flagged (warning mode) | :material-flag: |
| `sync_complete` | Rule was synced to agent | :material-sync: |

### Filtering

Filter changes by:

- **Team**: Select specific team
- **Agent**: Filter by agent
- **Event Type**: Filter by event type
- **Date Range**: Custom date range

### Change Details

Click a change to view:

- Full content diff
- Agent and user information
- Rule that was triggered
- Timestamp and metadata

## Approvals

For rules with approval workflows:

### Pending Approvals

View changes awaiting approval:

1. Navigate to **Approvals**
2. See list of pending requests
3. Click to review

### Approve/Reject

Review the change and:

- **Approve**: Accept the change
- **Reject**: Deny with reason
- **Request Changes**: Ask for modifications

## Settings

### General

- **Site Name**: Customize the UI title
- **Logo**: Upload custom logo
- **Theme**: Light/dark mode default

### Authentication

- **OAuth Providers**: Configure GitHub, Google SSO
- **Session Duration**: JWT expiration settings
- **2FA**: Enable two-factor authentication

### Notifications

Configure notification settings:

- **Email**: SMTP settings for email notifications
- **Webhooks**: Outbound webhook URLs
- **Slack**: Slack integration

### API Keys

Manage API keys for automation:

1. Navigate to **Settings** → **API Keys**
2. Click **Generate Key**
3. Set permissions and expiration
4. Copy the key (shown only once)

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` / `Cmd+K` | Open command palette |
| `g` `d` | Go to Dashboard |
| `g` `g` | Go to Graph View |
| `g` `t` | Go to Teams |
| `g` `u` | Go to Users |
| `g` `r` | Go to Rules |
| `g` `c` | Go to Changes |
| `g` `a` | Go to Approvals |
| `?` | Show keyboard shortcuts |
| `/` | Focus search |
| `Esc` | Close modal / command palette |

## Dark Mode

Toggle dark mode:

1. Click the theme toggle in the header
2. Or use system preference

## Mobile Support

The Web UI is responsive and supports:

- Tablets in portrait and landscape
- Mobile phones (limited functionality)

For full administrative capabilities, use a desktop browser.
