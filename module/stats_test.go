// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/k8s"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClusterStatRepo keeps one analysis per cluster for unit tests.
type stubClusterStatRepo struct {
	stats map[db.Id]*db.ClusterStat
}

func (s *stubClusterStatRepo) Upsert(stat *db.ClusterStat) error {
	s.stats[stat.ClusterId] = stat
	return nil
}

func (s *stubClusterStatRepo) GetByClusterId(clusterId db.Id) (*db.ClusterStat, error) {
	return s.stats[clusterId], nil
}

func TestUnitClusterStatsSaveAndGet(t *testing.T) {
	clusters := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	module := NewClusterStats(clusters, &stubClusterStatRepo{stats: map[db.Id]*db.ClusterStat{}})

	require.NoError(t, module.SaveStats("c-1", &k8s.Stats{}))

	stats, err := module.GetStats("w-1", "c-1")
	require.NoError(t, err)
	assert.Equal(t, &k8s.Stats{}, stats)
}

func TestUnitClusterStatsGetNotFound(t *testing.T) {
	clusters := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	module := NewClusterStats(clusters, &stubClusterStatRepo{stats: map[db.Id]*db.ClusterStat{}})

	_, err := module.GetStats("w-1", "c-1")
	assert.ErrorIs(t, err, ErrClusterStatsNotFound)
}

func TestUnitClusterStatsGetRejectsCrossWorkspace(t *testing.T) {
	clusters := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	module := NewClusterStats(clusters, &stubClusterStatRepo{stats: map[db.Id]*db.ClusterStat{}})

	_, err := module.GetStats("w-2", "c-1")
	assert.ErrorIs(t, err, ErrClusterNotFound)
}
