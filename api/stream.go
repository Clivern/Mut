// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EventStream starts a server-sent event response and returns a function that sends one named event.
func (a *API) EventStream(w http.ResponseWriter, timeout time.Duration) func(event string, data any) {
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(timeout))

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	return func(event string, data any) {
		body, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, body)
		_ = controller.Flush()
	}
}
