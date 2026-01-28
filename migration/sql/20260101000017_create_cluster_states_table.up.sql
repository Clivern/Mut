CREATE TABLE IF NOT EXISTS cluster_states (
	id UUID PRIMARY KEY,
	cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
	collected_at TIMESTAMP NOT NULL,
	state JSONB NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_cluster_states_cluster_id_collected_at ON cluster_states (cluster_id, collected_at DESC);
