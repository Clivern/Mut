// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
)

// LoginAction authenticates a user by email and password and sets the session cookie.
func (a *API) LoginAction(w http.ResponseWriter, r *http.Request) {
	var req module.LoginRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("email", req.Email).
		Msg("New login request")

	result, err := a.Auth.Login(r.Context(), &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrInvalidCredentials, http.StatusUnauthorized, locale.TR(r, "invalid_credentials")},
			{module.ErrAccountNotActive, http.StatusForbidden, locale.TR(r, "account_not_active")},
			{module.ErrMaintenanceModeEnabled, http.StatusServiceUnavailable, locale.TR(r, "maintenance_mode_enabled")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Login failed")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_login"),
		})

		return
	}

	util.SetCookie(w, "_mut_session", result.Session.Token, result.CookieOptions)

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"successMessage": locale.TR(r, "login_successful"),
		"user":           userResponse(result.User),
	})
}
