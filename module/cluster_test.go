// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/k8s"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClusterRepo serves canned cluster rows for unit tests.
type stubClusterRepo struct {
	clusters map[db.Id]*db.Cluster
}

func (s *stubClusterRepo) Create(cluster *db.Cluster) error { return nil }

func (s *stubClusterRepo) GetById(id db.Id) (*db.Cluster, error) {
	return s.clusters[id], nil
}

func (s *stubClusterRepo) ListAll() ([]*db.Cluster, error) {
	out := make([]*db.Cluster, 0, len(s.clusters))
	for _, cluster := range s.clusters {
		out = append(out, cluster)
	}

	return out, nil
}

func (s *stubClusterRepo) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.Cluster, error) {
	return nil, nil
}

func (s *stubClusterRepo) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return 0, nil
}

func (s *stubClusterRepo) Update(cluster *db.Cluster) error { return nil }

func (s *stubClusterRepo) Delete(id db.Id) error { return nil }

// stubClusterStateRepo keeps stored states in memory for unit tests.
type stubClusterStateRepo struct {
	states []*db.ClusterState
	pruned bool
}

func (s *stubClusterStateRepo) Create(state *db.ClusterState) error {
	s.states = append(s.states, state)

	return nil
}

func (s *stubClusterStateRepo) ListByClusterId(clusterId db.Id, limit int) ([]*db.ClusterState, error) {
	var out []*db.ClusterState
	for i := len(s.states) - 1; i >= 0 && len(out) < limit; i-- {
		if s.states[i].ClusterId == clusterId {
			out = append(out, s.states[i])
		}
	}

	return out, nil
}

func (s *stubClusterStateRepo) ListCollectedAtByClusterId(clusterId db.Id) ([]*db.ClusterState, error) {
	return nil, nil
}

func (s *stubClusterStateRepo) DeleteByIds(ids []db.Id) error {
	return nil
}

func (s *stubClusterStateRepo) DeleteCollectedBefore(before time.Time) error {
	s.pruned = true

	return nil
}

func TestUnitClusterGetStateReadsLatestStored(t *testing.T) {
	older, err := json.Marshal(k8s.ClusterState{Cluster: k8s.ClusterMeta{Name: "old"}})
	require.NoError(t, err)

	latest, err := json.Marshal(k8s.ClusterState{Cluster: k8s.ClusterMeta{Name: "latest"}})
	require.NoError(t, err)

	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	states := &stubClusterStateRepo{states: []*db.ClusterState{
		{ClusterId: "c-1", CollectedAt: time.Now().Add(-time.Minute), State: older},
		{ClusterId: "c-1", CollectedAt: time.Now(), State: latest},
	}}

	c := NewCluster(repo, states, nil)

	state, err := c.GetState(context.Background(), "w-1", "c-1")
	require.NoError(t, err)
	assert.Equal(t, "latest", state.Cluster.Name)
}

func TestUnitClusterGetStateNoStoredState(t *testing.T) {
	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}

	c := NewCluster(repo, &stubClusterStateRepo{}, nil)

	_, err := c.GetState(context.Background(), "w-1", "c-1")
	assert.ErrorIs(t, err, ErrClusterStateNotFound)
}

func TestUnitClusterGetStateCrossWorkspace(t *testing.T) {
	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-2": {Id: "c-2", WorkspaceId: "w-2", Name: "other"},
	}}
	states := &stubClusterStateRepo{states: []*db.ClusterState{
		{ClusterId: "c-2", CollectedAt: time.Now(), State: []byte(`{}`)},
	}}

	c := NewCluster(repo, states, nil)

	_, err := c.GetState(context.Background(), "w-1", "c-2")
	assert.ErrorIs(t, err, ErrClusterNotFound)

	_, err = c.GetState(context.Background(), "w-1", "missing")
	assert.ErrorIs(t, err, ErrClusterNotFound)
}

func TestUnitClusterSaveStoresPushedState(t *testing.T) {
	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	states := &stubClusterStateRepo{}

	c := NewCluster(repo, states, nil)

	state := k8s.NewClusterState("prod", "v1.31.0")
	err := c.SaveState("w-1", "c-1", state)
	require.NoError(t, err)

	require.Len(t, states.states, 1)
	assert.Equal(t, db.Id("c-1"), states.states[0].ClusterId)
	assert.Equal(t, state.Cluster.CollectedAt, states.states[0].CollectedAt)

	var stored k8s.ClusterState
	require.NoError(t, json.Unmarshal(states.states[0].State, &stored))
	assert.Equal(t, "prod", stored.Cluster.Name)
}

func TestUnitClusterSaveRejectsCrossWorkspace(t *testing.T) {
	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	states := &stubClusterStateRepo{}

	c := NewCluster(repo, states, nil)

	err := c.SaveState("w-2", "c-1", k8s.NewClusterState("prod", ""))
	assert.ErrorIs(t, err, ErrClusterNotFound)
	assert.Empty(t, states.states)
}

func TestUnitClusterListStatesBuildsNewestFirst(t *testing.T) {
	older, err := json.Marshal(k8s.ClusterState{Cluster: k8s.ClusterMeta{Name: "older"}})
	require.NoError(t, err)

	latest, err := json.Marshal(k8s.ClusterState{Cluster: k8s.ClusterMeta{Name: "latest"}})
	require.NoError(t, err)

	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}
	states := &stubClusterStateRepo{states: []*db.ClusterState{
		{Id: "s-1", ClusterId: "c-1", CollectedAt: time.Now().Add(-time.Minute), State: older},
		{Id: "s-2", ClusterId: "c-1", CollectedAt: time.Now(), State: latest},
	}}

	c := NewCluster(repo, states, nil)

	list, err := c.ListStates("w-1", "c-1", 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "latest", list[0].Cluster.Name)
	assert.Equal(t, "s-2", list[0].Id)
	assert.Equal(t, "older", list[1].Cluster.Name)
	assert.Equal(t, "s-1", list[1].Id)
}

func TestUnitClusterListStatesRejectsCrossWorkspace(t *testing.T) {
	repo := &stubClusterRepo{clusters: map[db.Id]*db.Cluster{
		"c-1": {Id: "c-1", WorkspaceId: "w-1", Name: "prod"},
	}}

	c := NewCluster(repo, &stubClusterStateRepo{}, nil)

	_, err := c.ListStates("w-2", "c-1", 10)
	assert.ErrorIs(t, err, ErrClusterNotFound)
}
