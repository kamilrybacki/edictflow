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
