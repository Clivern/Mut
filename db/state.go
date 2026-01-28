// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ClusterState is one stored state of a cluster. State holds the JSON-encoded
type ClusterState struct {
	Id          Id
	ClusterId   Id
	CollectedAt time.Time
	State       []byte
	CreatedAt   time.Time
}

// ClusterStateRepository is the interface for stored cluster states.
type ClusterStateRepository interface {
	Create(state *ClusterState) error
	ListByClusterId(clusterId Id, limit int) ([]*ClusterState, error)
	ListCollectedAtByClusterId(clusterId Id) ([]*ClusterState, error)
	DeleteByIds(ids []Id) error
	DeleteCollectedBefore(before time.Time) error
}

// ClusterStateRepositoryPostgres implements ClusterStateRepository for PostgreSQL
type ClusterStateRepositoryPostgres struct {
	db *sql.DB
}

// NewClusterStateRepository returns the cluster state repository for the current driver
func NewClusterStateRepository(db *sql.DB) ClusterStateRepository {
	return &ClusterStateRepositoryPostgres{db: db}
}

// scanClusterState maps a row into a state. Column order must match the select list.
func scanClusterState(s *ClusterState) []any {
	return []any{
		&s.Id,
		&s.ClusterId,
		&s.CollectedAt,
		&s.State,
		&s.CreatedAt,
	}
}

// scanCollectedAt maps a row of id, cluster id and collection time into a state without its body.
func scanCollectedAt(s *ClusterState) []any {
	return []any{
		&s.Id,
		&s.ClusterId,
		&s.CollectedAt,
	}
}

// Create inserts a stored cluster state and sets its Id and created time
func (r *ClusterStateRepositoryPostgres) Create(state *ClusterState) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	state.Id = id

	return r.db.QueryRow(
		`INSERT INTO cluster_states
		(id, cluster_id, collected_at, state)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		state.Id.String(),
		state.ClusterId.String(),
		state.CollectedAt,
		string(state.State),
	).Scan(&state.CreatedAt)
}

// ListByClusterId returns a cluster's stored states, most recently collected first
func (r *ClusterStateRepositoryPostgres) ListByClusterId(clusterId Id, limit int) ([]*ClusterState, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, cluster_id, collected_at, state, created_at
		FROM cluster_states
		WHERE cluster_id = $1
		ORDER BY collected_at DESC
		LIMIT $2`,
		scanClusterState,
		clusterId.String(),
		limit,
	)
}

// ListCollectedAtByClusterId returns a cluster's state ids and collection times without the state bodies, newest first
func (r *ClusterStateRepositoryPostgres) ListCollectedAtByClusterId(clusterId Id) ([]*ClusterState, error) {
	return QueryMany(
		r.db,
		`SELECT id, cluster_id, collected_at FROM cluster_states WHERE cluster_id = $1 ORDER BY collected_at DESC`,
		scanCollectedAt,
		clusterId.String(),
	)
}

// DeleteByIds removes the stored states with the given ids
func (r *ClusterStateRepositoryPostgres) DeleteByIds(ids []Id) error {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id.String()
	}

	_, err := r.db.Exec(
		`DELETE FROM cluster_states WHERE id IN (`+strings.Join(placeholders, ", ")+`)`,
		args...,
	)

	return err
}

// DeleteCollectedBefore removes stored states collected before the given time
func (r *ClusterStateRepositoryPostgres) DeleteCollectedBefore(before time.Time) error {
	_, err := r.db.Exec(
		`DELETE FROM cluster_states WHERE collected_at < $1`,
		before,
	)

	return err
}
