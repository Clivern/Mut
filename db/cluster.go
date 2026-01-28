// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Cluster is the DB row for a workspace cluster.
type Cluster struct {
	Id          Id
	WorkspaceId Id
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ClusterRepository is the interface for workspace cluster CRUD.
type ClusterRepository interface {
	Create(cluster *Cluster) error
	Update(cluster *Cluster) error
	GetById(id Id) (*Cluster, error)
	ListAll() ([]*Cluster, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Cluster, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	Delete(id Id) error
}

// ClusterRepositoryPostgres implements ClusterRepository for PostgreSQL
type ClusterRepositoryPostgres struct {
	db *sql.DB
}

// NewClusterRepository returns the cluster repository for the current driver
func NewClusterRepository(db *sql.DB) ClusterRepository {
	return &ClusterRepositoryPostgres{db: db}
}

// scanCluster maps a row into a cluster. Column order must match the select list.
func scanCluster(c *Cluster) []any {
	return []any{
		&c.Id,
		&c.WorkspaceId,
		&c.Name,
		&c.Description,
		&c.CreatedAt,
		&c.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a cluster and sets its Id and timestamps
func (r *ClusterRepositoryPostgres) Create(cluster *Cluster) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	cluster.Id = id

	err = r.db.QueryRow(
		`INSERT INTO clusters
		(id, workspace_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`,
		cluster.Id.String(),
		cluster.WorkspaceId.String(),
		cluster.Name,
		cluster.Description,
	).Scan(
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	return err
}

// Update saves a cluster's name and description and sets its updated time
func (r *ClusterRepositoryPostgres) Update(cluster *Cluster) error {
	return r.db.QueryRow(
		`UPDATE clusters
		SET name = $2, description = $3, updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		WHERE id = $1
		RETURNING updated_at`,
		cluster.Id.String(),
		cluster.Name,
		cluster.Description,
	).Scan(&cluster.UpdatedAt)
}

// GetById returns a cluster by Id, or nil when it does not exist
func (r *ClusterRepositoryPostgres) GetById(id Id) (*Cluster, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, name, description, created_at, updated_at
		FROM clusters
		WHERE id = $1`,
		scanCluster,
		id.String(),
	)
}

// ListAll returns every cluster across workspaces, for the state collector
func (r *ClusterRepositoryPostgres) ListAll() ([]*Cluster, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, name, description, created_at, updated_at
		FROM clusters
		ORDER BY created_at ASC`,
		scanCluster,
	)
}

// ListByWorkspaceId returns a page of clusters for a workspace, newest first
func (r *ClusterRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Cluster, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, name, description, created_at, updated_at
		FROM clusters
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanCluster,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns how many clusters a workspace has
func (r *ClusterRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM clusters WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&total)

	return total, err
}

// Delete removes a cluster
func (r *ClusterRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM clusters WHERE id = $1`,
		id.String(),
	)

	return err
}
