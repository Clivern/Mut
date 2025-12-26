// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// GetMeAction returns the authenticated session or API key principal.
func (a *API) GetMeAction(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")

	// Fall back to the session user resolved by the auth middleware
	if lo.IsEmpty(apiKey) {
		if user, ok := a.GetUser(r); ok && user != nil {
			a.WriteJSON(w, http.StatusOK, a.Me.GetBySession(user))
			return
		}

		a.WriteJSON(w, http.StatusForbidden, map[string]any{
			"errorMessage": locale.TR(r, "me_requires_key_header"),
		})
		return
	}

	me, err := a.Me.GetByAPIKey(apiKey)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAPIKeyNotFound, http.StatusUnauthorized, locale.TR(r, "invalid_api_key")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Msg("Failed to resolve API key for /me")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_me"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, me)
}
