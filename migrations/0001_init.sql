-- Tablon M1 schema.

CREATE TABLE projects (
    id              TEXT PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'planning'
                    CHECK (status IN ('planning', 'active', 'archived', 'done')),
    logo_gradient   INTEGER NOT NULL DEFAULT 0
                    CHECK (logo_gradient BETWEEN 0 AND 7),
    logo_emoji      TEXT NOT NULL DEFAULT '',
    issue_counter   INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

-- Identifiers are never reused, even after a project is deleted.
CREATE TABLE project_identifiers (
    identifier  TEXT PRIMARY KEY,
    created_at  TEXT NOT NULL
);

CREATE TABLE epics (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'backlog'
                CHECK (status IN ('backlog', 'in-progress', 'done')),
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE issues (
    id          TEXT PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    epic_id     TEXT REFERENCES epics(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'backlog'
                CHECK (status IN ('backlog', 'todo', 'in-progress', 'in-review', 'done', 'canceled')),
    priority    TEXT NOT NULL DEFAULT 'none'
                CHECK (priority IN ('urgent', 'high', 'medium', 'low', 'none')),
    agent_id    TEXT,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE INDEX idx_epics_project_id ON epics (project_id, created_at DESC);
CREATE INDEX idx_issues_project_id ON issues (project_id, created_at DESC, id DESC);
CREATE INDEX idx_issues_epic_id ON issues (epic_id);
CREATE INDEX idx_issues_status ON issues (project_id, status);
CREATE INDEX idx_issues_priority ON issues (project_id, priority);
CREATE INDEX idx_issues_updated_at ON issues (project_id, updated_at DESC, id DESC);
CREATE INDEX idx_issues_identifier ON issues (identifier);
