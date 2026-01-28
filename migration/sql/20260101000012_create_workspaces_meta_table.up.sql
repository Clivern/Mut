CREATE TABLE IF NOT EXISTS workspaces_meta (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	key VARCHAR(60) NOT NULL,
	value JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (workspace_id, key)
);
CREATE INDEX IF NOT EXISTS idx_workspaces_meta_workspace_id_key ON workspaces_meta(workspace_id, key);
