// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"sort"
	"time"
)

// Compact works on an array of cluster states and picks the ones to remove.
type Compact struct {
	// KeepAllFor keeps every state newer than this age.
	KeepAllFor time.Duration
	// KeepHourlyFor keeps one state per hour for states older than KeepAllFor and newer than this age.
	KeepHourlyFor time.Duration
}

// NewCompact returns a Compact that keeps everything for a day, one state per hour for a week,
// and one state per day after that.
func NewCompact() *Compact {
	return &Compact{
		KeepAllFor:    24 * time.Hour,
		KeepHourlyFor: 7 * 24 * time.Hour,
	}
}

// Run returns the ids of the states to remove.
func (c *Compact) Run(states []ClusterState) []string {
	if len(states) == 0 {
		return nil
	}

	newest := states[0].Cluster.CollectedAt
	for _, state := range states {
		if state.Cluster.CollectedAt.After(newest) {
			newest = state.Cluster.CollectedAt
		}
	}

	sorted := make([]ClusterState, len(states))
	copy(sorted, states)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Cluster.CollectedAt.After(sorted[j].Cluster.CollectedAt)
	})

	kept := map[string]bool{}
	var remove []string
	for _, state := range sorted {
		collectedAt := state.Cluster.CollectedAt
		age := newest.Sub(collectedAt)

		bucket := state.Id
		switch {
		case age < c.KeepAllFor:
		case age < c.KeepHourlyFor:
			bucket = "hour:" + collectedAt.UTC().Truncate(time.Hour).Format(time.RFC3339)
		default:
			bucket = "day:" + collectedAt.UTC().Truncate(24*time.Hour).Format(time.RFC3339)
		}

		if kept[bucket] {
			remove = append(remove, state.Id)
			continue
		}

		kept[bucket] = true
	}

	return remove
}
