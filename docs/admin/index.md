# Administrator Guide

This guide covers deploying, configuring, and managing a Edictflow server for your organization.

## Overview

As an administrator, you are responsible for:

- **Deployment** - Setting up and maintaining the server infrastructure
- **Configuration** - Tuning server settings for your environment
- **User Management** - Creating users, assigning roles, managing teams
- **Rule Management** - Defining and enforcing CLAUDE.md configurations
- **Monitoring** - Tracking agent connections and configuration changes

## Quick Navigation

<div class="grid" markdown>

<div class="card" markdown>

### [Deployment](deployment.md)

Deploy Edictflow using Docker Compose, Kubernetes, or manual installation.

</div>

<div class="card" markdown>

### [Configuration](configuration.md)

Configure the server via environment variables and config files.

</div>

<div class="card" markdown>

### [Database Setup](database.md)

Set up PostgreSQL, run migrations, and manage backups.

</div>

<div class="card" markdown>

### [Web UI](../web-ui/index.md)

Navigate and use the web interface for managing rules, teams, and agents.

</div>

<div class="card" markdown>

### [User Management](users.md)

Create users, manage teams, and configure authentication.

</div>

<div class="card" markdown>

### [Roles & Permissions](rbac.md)

Configure role-based access control for your organization.

</div>

<div class="card" markdown>

### [Audit Logging](audit.md)

Monitor changes and maintain compliance with audit trails.

</div>

</div>

## Deployment Options

| Method | Best For | Complexity |
|--------|----------|------------|
| [Docker Compose](deployment.md#docker-compose-recommended) | Small teams, evaluation | Low |
| [Kubernetes](deployment.md#kubernetes) | Production, scale | Medium |
| [Manual](deployment.md#manual-installation) | Custom environments | High |

## First Steps

After deploying Edictflow:

1. **Access the Web UI** at `https://your-server:3000`
2. **Create an admin account** or configure OAuth
3. **Create your first team** to organize users
4. **Define rules** for CLAUDE.md management
5. **Invite users** and distribute agent installation instructions

## Security Considerations

!!! warning "Production Checklist"

    Before going to production, ensure:

    - [ ] TLS/HTTPS is enabled
    - [ ] Default credentials are changed
    - [ ] Database is secured and backed up
    - [ ] OAuth/SSO is configured (recommended)
    - [ ] Firewall rules are in place
    - [ ] Audit logging is enabled
    - [ ] Log rotation is configured

## Support

For issues and questions:

- [GitHub Issues](https://github.com/kamilrybacki/edictflow/issues)
- [Documentation](../index.md)
