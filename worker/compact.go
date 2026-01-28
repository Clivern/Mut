// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/broker"

	"github.com/rs/zerolog/log"
)

// HandleCompactStates removes the stored states of a cluster that the retention policy no longer keeps.
func (h *handlers) HandleCompactStates(_ context.Context, msg *broker.Msg) error {
	var payload map[string]string
	err := json.Unmarshal(msg.Data, &payload)
	if err != nil {
		return ErrInvalidPayload
	}

	taskId := db.Id(payload["taskId"])
	clusterId := db.Id(payload["clusterId"])

	err = h.tasks.MarkRunning(taskId)
	if err != nil {
		return err
	}

	removed, err := h.cluster.CompactStates(clusterId)
	if err != nil {
		return err
	}

	pruned, err := h.tasks.DeleteFinishedBefore(
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1),
	)
	if err != nil {
		return err
	}

	log.Info().
		Str("cluster_id", clusterId.String()).
		Int("removed", removed).
		Int64("pruned_tasks", pruned).
		Msg("Cluster states compacted")

	return h.tasks.Complete(taskId, fmt.Sprintf("removed %d states, pruned %d tasks", removed, pruned))
}
