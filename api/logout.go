// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
)

// LogoutAction logs the user out and revokes their sessions.
func (a *API) LogoutAction(w http.ResponseWriter, r *http.Request) {
	token := util.GetCookie(r, "_mut_session")

	util.DeleteCookie(w, "_mut_session")

	log.Info().
		Msg("New logout request")

	// Logout is a public route, so the session is resolved from the cookie here
	err := a.Auth.LogoutByToken(token)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to revoke session")
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"successMessage": locale.TR(r, "logout_successful"),
	})
}
