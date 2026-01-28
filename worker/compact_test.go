// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package worker

import (
	"context"
	"testing"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/broker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClusterStates serves state rows and records the ids that get deleted.
type fakeClusterStates struct {
	db.ClusterStateRepository
	rows    []*db.ClusterState
	deleted []db.Id
}

func (f *fakeClusterStates) ListCollectedAtByClusterId(clusterId db.Id) ([]*db.ClusterState, error) {
	return f.rows, nil
}

func (f *fakeClusterStates) DeleteByIds(ids []db.Id) error {
	f.deleted = append(f.deleted, ids...)
	return nil
}

func TestUnitHandleCompactStates(t *testing.T) {
	base := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("removes states the retention policy no longer keeps", func(t *testing.T) {
		tasks := &recordingTasks{}
		states := &fakeClusterStates{rows: []*db.ClusterState{
			{Id: "new", CollectedAt: base},
			{Id: "old-1", CollectedAt: base.Add(-3*24*time.Hour + 10*time.Minute)},
			{Id: "old-2", CollectedAt: base.Add(-3*24*time.Hour + 40*time.Minute)},
		}}
		h := &handlers{tasks: tasks, cluster: module.NewCluster(nil, states, nil)}

		err := h.HandleCompactStates(context.Background(), &broker.Msg{
			Data: []byte(`{"taskId":"task-1","clusterId":"c-1"}`),
		})
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"old-1"}, states.deleted)
		assert.Equal(t, []db.Id{"task-1"}, tasks.running)
		assert.Equal(t, []db.Id{"task-1"}, tasks.completed)
		require.Len(t, tasks.pruneCutoffs, 1)
		cutoff := tasks.pruneCutoffs[0]
		assert.Equal(t, time.UTC, cutoff.Location())
		assert.Equal(t, 0, cutoff.Hour()+cutoff.Minute()+cutoff.Second())
		assert.True(t, cutoff.Before(time.Now().Add(-24*time.Hour)))
		assert.True(t, cutoff.After(time.Now().Add(-48*time.Hour)))
	})

	t.Run("deletes nothing when every state is kept", func(t *testing.T) {
		states := &fakeClusterStates{rows: []*db.ClusterState{
			{Id: "new", CollectedAt: base},
			{Id: "older", CollectedAt: base.Add(-time.Minute)},
		}}
		h := &handlers{tasks: &recordingTasks{}, cluster: module.NewCluster(nil, states, nil)}

		err := h.HandleCompactStates(context.Background(), &broker.Msg{
			Data: []byte(`{"taskId":"task-2","clusterId":"c-1"}`),
		})
		require.NoError(t, err)
		assert.Empty(t, states.deleted)
	})
}
