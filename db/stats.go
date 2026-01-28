// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// ClusterStat is the analysis of one cluster. Stats holds the JSON-encoded analysis.
type ClusterStat struct {
	Id        Id
	ClusterId Id
	Stats     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClusterStatRepository is the interface for cluster analyses.
type ClusterStatRepository interface {
	Upsert(stat *ClusterStat) error
	GetByClusterId(clusterId Id) (*ClusterStat, error)
}

// ClusterStatRepositoryPostgres implements ClusterStatRepository for PostgreSQL
type ClusterStatRepositoryPostgres struct {
	db *sql.DB
}

// NewClusterStatRepository returns the cluster stat repository for the current driver
func NewClusterStatRepository(db *sql.DB) ClusterStatRepository {
	return &ClusterStatRepositoryPostgres{db: db}
}

// scanClusterStat maps a row into a cluster stat. Column order must match the select list.
func scanClusterStat(s *ClusterStat) []any {
	return []any{
		&s.Id,
		&s.ClusterId,
		&s.Stats,
		&s.CreatedAt,
		&s.UpdatedAt,
	}
}

// Upsert stores the analysis of a cluster, replacing the one it had, and sets the stat's Id and timestamps
func (r *ClusterStatRepositoryPostgres) Upsert(stat *ClusterStat) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	stat.Id = id

	return r.db.QueryRow(
		`INSERT INTO cluster_stats
		(id, cluster_id, stats)
		VALUES ($1, $2, $3)
		ON CONFLICT (cluster_id) DO UPDATE
		SET stats = EXCLUDED.stats, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		RETURNING id, created_at, updated_at`,
		stat.Id.String(),
		stat.ClusterId.String(),
		string(stat.Stats),
	).Scan(
		&stat.Id,
		&stat.CreatedAt,
		&stat.UpdatedAt,
	)
}

// GetByClusterId returns a cluster's analysis, or nil when it has none yet
func (r *ClusterStatRepositoryPostgres) GetByClusterId(clusterId Id) (*ClusterStat, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, cluster_id, stats, created_at, updated_at
		FROM cluster_stats
		WHERE cluster_id = $1`,
		scanClusterStat,
		clusterId.String(),
	)
}
