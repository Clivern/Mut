// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/k8s"
)

// ErrClusterStatsNotFound is returned when a cluster has no analysis yet.
var ErrClusterStatsNotFound = errors.New("cluster stats not found")

// ClusterStats serves the analysis of workspace clusters.
type ClusterStats struct {
	ClusterRepository     db.ClusterRepository
	ClusterStatRepository db.ClusterStatRepository
}

// NewClusterStats creates a cluster stats module backed by the given repositories.
func NewClusterStats(clusters db.ClusterRepository, stats db.ClusterStatRepository) *ClusterStats {
	return &ClusterStats{
		ClusterRepository:     clusters,
		ClusterStatRepository: stats,
	}
}

// SaveStats stores the analysis of a cluster, replacing the previous one.
func (c *ClusterStats) SaveStats(clusterId db.Id, stats *k8s.Stats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	return c.ClusterStatRepository.Upsert(&db.ClusterStat{
		ClusterId: clusterId,
		Stats:     data,
	})
}

// GetStats returns the analysis of a workspace cluster.
func (c *ClusterStats) GetStats(workspaceId, clusterId db.Id) (*k8s.Stats, error) {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return nil, ErrClusterNotFound
	}

	stat, err := c.ClusterStatRepository.GetByClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	if stat == nil {
		return nil, ErrClusterStatsNotFound
	}

	var stats k8s.Stats
	err = json.Unmarshal(stat.Stats, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
