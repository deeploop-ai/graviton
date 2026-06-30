CREATE TABLE IF NOT EXISTS project_oauth_providers (
    project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider      TEXT NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    client_id     TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    scopes        TEXT[] NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_project_oauth_providers_project ON project_oauth_providers(project_id);
