// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/broker"
)

// HandleAnalyzeStates analyzes a cluster's stored states after a new state arrives. It does nothing yet.
func (h *handlers) HandleAnalyzeStates(_ context.Context, msg *broker.Msg) error {
	var payload map[string]string
	err := json.Unmarshal(msg.Data, &payload)
	if err != nil {
		return ErrInvalidPayload
	}

	taskId := db.Id(payload["taskId"])

	err = h.tasks.MarkRunning(taskId)
	if err != nil {
		return err
	}

	return h.tasks.Complete(taskId, "")
}
