// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/k8s"

	"github.com/rs/zerolog/log"
)

var (
	// ErrClusterNotFound is returned when a cluster does not exist in the workspace.
	ErrClusterNotFound = errors.New("cluster not found")
	// ErrClusterStateNotFound is returned when a cluster has no stored state yet.
	ErrClusterStateNotFound = errors.New("cluster state not found")
)

// Cluster serves stored Kubernetes state for workspace clusters.
type Cluster struct {
	ClusterRepository      db.ClusterRepository
	ClusterStateRepository db.ClusterStateRepository
	WorkspaceRepository    db.WorkspaceRepository
}

// NewCluster creates a cluster module backed by the given repositories.
func NewCluster(clusters db.ClusterRepository, states db.ClusterStateRepository, workspaces db.WorkspaceRepository) *Cluster {
	return &Cluster{
		ClusterRepository:      clusters,
		ClusterStateRepository: states,
		WorkspaceRepository:    workspaces,
	}
}

// SaveState stores a state pushed by a cluster agent for one of a workspace's clusters.
func (c *Cluster) SaveState(workspaceId, clusterId db.Id, state *k8s.ClusterState) error {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return ErrClusterNotFound
	}

	state.Cluster.Name = cluster.Name

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return c.ClusterStateRepository.Create(&db.ClusterState{
		ClusterId:   clusterId,
		CollectedAt: state.Cluster.CollectedAt,
		State:       data,
	})
}

// ListStates returns a workspace cluster's stored states as cluster states, newest first.
func (c *Cluster) ListStates(workspaceId, clusterId db.Id, limit int) ([]k8s.ClusterState, error) {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return nil, ErrClusterNotFound
	}

	rows, err := c.ClusterStateRepository.ListByClusterId(clusterId, limit)
	if err != nil {
		return nil, err
	}

	states := make([]k8s.ClusterState, 0, len(rows))
	for _, row := range rows {
		var state k8s.ClusterState
		err = json.Unmarshal(row.State, &state)
		if err != nil {
			return nil, err
		}

		state.Id = row.Id.String()
		states = append(states, state)
	}

	return states, nil
}

// CompactStates removes a cluster's stored states that the retention policy no longer keeps and returns how many it removed.
func (c *Cluster) CompactStates(clusterId db.Id) (int, error) {
	rows, err := c.ClusterStateRepository.ListCollectedAtByClusterId(clusterId)
	if err != nil {
		return 0, err
	}

	states := make([]k8s.ClusterState, 0, len(rows))
	for _, row := range rows {
		states = append(states, k8s.ClusterState{
			Id:      row.Id.String(),
			Cluster: k8s.ClusterMeta{CollectedAt: row.CollectedAt},
		})
	}

	remove := k8s.NewCompact().Run(states)
	if len(remove) == 0 {
		return 0, nil
	}

	ids := make([]db.Id, 0, len(remove))
	for _, id := range remove {
		ids = append(ids, db.Id(id))
	}

	err = c.ClusterStateRepository.DeleteByIds(ids)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// GetState returns the latest stored state of a workspace cluster.
func (c *Cluster) GetState(ctx context.Context, workspaceId, clusterId db.Id) (*k8s.ClusterState, error) {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		log.Error().
			Err(err).
			Str("clusterId", clusterId.String()).
			Msg("Failed to load cluster")

		return nil, err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return nil, ErrClusterNotFound
	}

	states, err := c.ClusterStateRepository.ListByClusterId(clusterId, 1)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		log.Info().
			Str("clusterId", clusterId.String()).
			Msg("No stored cluster state yet")

		return nil, ErrClusterStateNotFound
	}

	var state k8s.ClusterState
	err = json.Unmarshal(states[0].State, &state)
	if err != nil {
		return nil, err
	}

	state.Id = states[0].Id.String()

	return &state, nil
}
