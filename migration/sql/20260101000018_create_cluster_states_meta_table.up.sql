CREATE TABLE IF NOT EXISTS cluster_states_meta (
	id UUID PRIMARY KEY,
	cluster_state_id UUID NOT NULL REFERENCES cluster_states(id) ON DELETE CASCADE,
	key VARCHAR(60) NOT NULL,
	value JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (cluster_state_id, key)
);
CREATE INDEX IF NOT EXISTS idx_cluster_states_meta_cluster_state_id_key ON cluster_states_meta(cluster_state_id, key);
