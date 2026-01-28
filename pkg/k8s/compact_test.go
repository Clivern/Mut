// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnitCompactRunPicksStatesToRemove(t *testing.T) {
	newest := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

	states := []ClusterState{
		ClusterState{Id: "recent-1", Cluster: ClusterMeta{CollectedAt: newest}},
		ClusterState{Id: "recent-2", Cluster: ClusterMeta{CollectedAt: newest.Add(-5 * time.Minute)}},
		ClusterState{Id: "recent-3", Cluster: ClusterMeta{CollectedAt: newest.Add(-20 * time.Hour)}},
		ClusterState{Id: "hour-a", Cluster: ClusterMeta{CollectedAt: newest.Add(-3*24*time.Hour + 10*time.Minute)}},
		ClusterState{Id: "hour-b", Cluster: ClusterMeta{CollectedAt: newest.Add(-3*24*time.Hour + 40*time.Minute)}},
		ClusterState{Id: "day-a", Cluster: ClusterMeta{CollectedAt: newest.Add(-10*24*time.Hour + 1*time.Hour)}},
		ClusterState{Id: "day-b", Cluster: ClusterMeta{CollectedAt: newest.Add(-10*24*time.Hour + 5*time.Hour)}},
	}

	remove := NewCompact().Run(states)

	assert.ElementsMatch(t, []string{"hour-a", "day-a"}, remove)
}

func TestUnitCompactRunKeepsEverythingWhenRecent(t *testing.T) {
	newest := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

	states := []ClusterState{
		ClusterState{Id: "a", Cluster: ClusterMeta{CollectedAt: newest}},
		ClusterState{Id: "b", Cluster: ClusterMeta{CollectedAt: newest.Add(-time.Minute)}},
	}

	assert.Empty(t, NewCompact().Run(states))
}
