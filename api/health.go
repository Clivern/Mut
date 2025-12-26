// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// HealthAction returns a simple health check (status ok).
func (a *API) HealthAction(w http.ResponseWriter, _ *http.Request) {
	log.Debug().
		Msg("Health check")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}
