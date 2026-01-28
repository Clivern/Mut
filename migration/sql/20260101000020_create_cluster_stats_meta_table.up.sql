CREATE TABLE IF NOT EXISTS cluster_stats_meta (
	id UUID PRIMARY KEY,
	cluster_stat_id UUID NOT NULL REFERENCES cluster_stats(id) ON DELETE CASCADE,
	key VARCHAR(60) NOT NULL,
	value JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (cluster_stat_id, key)
);
CREATE INDEX IF NOT EXISTS idx_cluster_stats_meta_cluster_stat_id_key ON cluster_stats_meta(cluster_stat_id, key);
