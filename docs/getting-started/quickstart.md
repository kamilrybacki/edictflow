# Quick Start

<div class="role-selector">
  <button class="role-btn active" data-role="server">I'm setting up a server</button>
  <button class="role-btn" data-role="user">I'm joining as a user</button>
</div>

<div class="role-content" data-role="server" markdown="1">

## Server Setup

Deploy Edictflow to manage CLAUDE.md configurations for your team or organization.

### Prerequisites

- Docker and Docker Compose
- A domain name (for production) or local access
- PostgreSQL 15+ (included in Docker setup)

### Step 1: Get the Server

**Docker Compose (Recommended):**

```bash
# Download the production compose file
curl -O https://raw.githubusercontent.com/kamilrybacki/edictflow/main/docker-compose.prod.yml

# Create environment file
cat > .env << EOF
POSTGRES_PASSWORD=your-secure-database-password
JWT_SECRET=your-secure-jwt-secret-min-32-chars
ADMIN_EMAIL=admin@yourcompany.com
ADMIN_PASSWORD=your-secure-admin-password
EOF

# Start services
docker compose -f docker-compose.prod.yml up -d
```

**Kubernetes:**

```bash
# Add the Helm repository
helm repo add edictflow https://charts.edictflow.dev
helm repo update

# Install with custom values
helm install edictflow edictflow/edictflow \
  --set admin.email=admin@yourcompany.com \
  --set admin.password=your-secure-admin-password \
  --set postgresql.auth.password=your-db-password
```

For manual installation, see [Server Deployment](../admin/deployment.md).

### Step 2: Access the Admin Panel

Once services are running, access the web UI:

- **Local**: `http://localhost:3000`
- **Production**: `https://your-domain.com`

Log in with the admin credentials you configured in your environment.

!!! warning "Secure Your Installation"
    - Use strong, unique passwords for admin and database
    - Enable HTTPS in production (see [Configuration](../admin/configuration.md))
    - Restrict network access to the admin panel

### Step 3: Create Your Organization Structure

1. **Create Teams** - Navigate to Teams and create teams for your organization (e.g., "Frontend", "Backend", "DevOps")

2. **Invite Users** - Go to Users → Invite and send invitations to team members

3. **Create Rules** - Define CLAUDE.md configuration rules for each team:
    - Set the target layer (Enterprise, Team, or Project)
    - Configure enforcement mode (Block, Temporary, or Warning)
    - Add triggers for when rules should apply

### Step 4: Distribute Agent Instructions

Share the agent installation instructions with your team members, providing them with:

- Your server URL
- Their login credentials (or SSO instructions)

### Next Steps

- [Configuration Guide](../admin/configuration.md) - Advanced server settings
- [User Management](../admin/users.md) - Managing users and permissions
- [Roles & Permissions](../admin/rbac.md) - Setting up access control

</div>

<div class="role-content" data-role="user" style="display: none;" markdown="1">

## User Setup

Connect to your organization's Edictflow server to keep your CLAUDE.md files in sync.

### Prerequisites

- Server URL from your administrator
- Your login credentials

### Step 1: Install the Agent

**macOS:**

```bash
brew tap kamilrybacki/edictflow
brew install edictflow-agent
```

**Linux:**

```bash
curl -sSL https://get.edictflow.dev | sh
```

**Windows:**

```powershell
# Using Scoop
scoop bucket add edictflow https://github.com/kamilrybacki/scoop-edictflow
scoop install edictflow-agent
```

**From Source:**

```bash
go install github.com/kamilrybacki/edictflow/agent@latest
```

### Step 2: Connect to Your Server

```bash
# Login to your organization's server
edictflow login https://edictflow.yourcompany.com
```

You'll be prompted for your credentials or redirected to your organization's SSO provider.

### Step 3: Start the Agent

```bash
# Start in foreground (for testing)
edictflow start --foreground

# Or run as a background service
edictflow start
```

The agent will:

1. Connect to your server via WebSocket
2. Download rules assigned to you
3. Watch your project directories for CLAUDE.md changes
4. Sync configurations in real-time

### Step 4: Verify Connection

```bash
# Check agent status
edictflow status
```

You should see:

```
Agent Status: Connected
Server: https://edictflow.yourcompany.com
User: your.email@company.com
Watched Projects: 3
Rules Applied: 12
```

### Working with Rules

Your CLAUDE.md files are now managed by Edictflow. Depending on your organization's enforcement settings:

- **Block Mode**: Changes to managed sections are automatically reverted
- **Temporary Mode**: Changes apply temporarily but revert if not approved
- **Warning Mode**: Changes apply but are flagged for review

To see which rules apply to your current project:

```bash
edictflow rules list
```

### Next Steps

- [CLI Reference](../user/cli.md) - Full command documentation
- [Workflow Guide](../user/workflow.md) - Day-to-day usage patterns
- [Troubleshooting](../user/troubleshooting.md) - Common issues and solutions

</div>

<script>
document.addEventListener('DOMContentLoaded', function() {
  var buttons = document.querySelectorAll('.role-btn');
  var contents = document.querySelectorAll('.role-content');

  buttons.forEach(function(btn) {
    btn.addEventListener('click', function() {
      var role = this.getAttribute('data-role');

      buttons.forEach(function(b) { b.classList.remove('active'); });
      this.classList.add('active');

      contents.forEach(function(content) {
        if (content.getAttribute('data-role') === role) {
          content.style.display = 'block';
        } else {
          content.style.display = 'none';
        }
      });

      if (typeof window.initTabs === 'function') {
        setTimeout(window.initTabs, 50);
      }
    });
  });
});
</script>

<style>
.role-selector {
  display: flex;
  gap: 1rem;
  margin: 2rem 0;
  justify-content: center;
}

.role-btn {
  padding: 1rem 2rem;
  font-size: 1rem;
  font-weight: 600;
  border: 2px solid #374151;
  border-radius: 8px;
  background: #1f2937;
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.2s ease;
}

.role-btn:hover {
  border-color: #3b82f6;
  color: #e5e7eb;
}

.role-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #fff;
}

.role-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
