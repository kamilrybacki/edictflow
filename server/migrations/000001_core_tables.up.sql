-- 000001_core_tables.up.sql
-- Core tables: teams, users, rules, projects, agents

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teams_name ON teams(name);

-- Users (with all fields from later migrations)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    auth_provider VARCHAR(50) NOT NULL,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    email_verified BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_by ON users(created_by);

-- Rules (with all fields from later migrations)
CREATE TABLE rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    target_layer VARCHAR(50) NOT NULL CHECK (target_layer IN ('organization', 'team', 'project')),
    priority_weight INTEGER NOT NULL DEFAULT 0,
    triggers JSONB NOT NULL DEFAULT '[]',
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected')),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    submitted_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    enforcement_mode TEXT NOT NULL DEFAULT 'block',
    temporary_timeout_hours INTEGER NOT NULL DEFAULT 24,
    category_id UUID,
    overridable BOOLEAN NOT NULL DEFAULT TRUE,
    effective_start TIMESTAMP WITH TIME ZONE,
    effective_end TIMESTAMP WITH TIME ZONE,
    target_teams UUID[] DEFAULT '{}',
    target_users UUID[] DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    force BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT rules_force_global_only CHECK (force = false OR team_id IS NULL),
    CONSTRAINT rules_global_organization_only CHECK (team_id IS NOT NULL OR target_layer = 'organization')
);

CREATE INDEX idx_rules_team_id ON rules(team_id);
CREATE INDEX idx_rules_target_layer ON rules(target_layer);
CREATE INDEX idx_rules_status ON rules(status);
CREATE INDEX idx_rules_created_by ON rules(created_by);
CREATE INDEX idx_rules_category_id ON rules(category_id);
CREATE INDEX idx_rules_effective_dates ON rules(effective_start, effective_end);
CREATE INDEX idx_rules_target_teams ON rules USING GIN(target_teams);
CREATE INDEX idx_rules_target_users ON rules USING GIN(target_users);
CREATE INDEX idx_rules_tags ON rules USING GIN(tags);
CREATE INDEX idx_rules_global ON rules(force) WHERE team_id IS NULL;
CREATE INDEX idx_rules_force ON rules(force) WHERE force = true;

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    path_pattern VARCHAR(500) NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_team_id ON projects(team_id);
CREATE INDEX idx_projects_tags ON projects USING GIN(tags);

-- Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    machine_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'offline',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cached_config_version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(machine_id, user_id)
);

CREATE INDEX idx_agents_user_id ON agents(user_id);
CREATE INDEX idx_agents_status ON agents(status);
