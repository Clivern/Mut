// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Config configures Protect middleware for a route or route group.
type Config struct {
	Roles []string // platform roles allowed on the route
}

// Protect returns middleware that applies user auth and role checks.
func Protect(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok || user == nil {
				log.Info().
					Str("path", r.URL.Path).
					Msg("User required: principal not in context")
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": locale.TR(r, "not_authenticated"),
				})
				return
			}

			if !user.IsActive {
				log.Info().
					Str("path", r.URL.Path).
					Str("userId", user.Id.String()).
					Msg("Inactive user blocked from protected route")
				util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
					"errorMessage": locale.TR(r, "account_not_active"),
				})
				return
			}

			if len(cfg.Roles) > 0 && !lo.Contains(cfg.Roles, user.Role) {
				log.Info().
					Str("path", r.URL.Path).
					Str("userId", user.Id.String()).
					Str("userRole", user.Role).
					Strs("allowedRoles", cfg.Roles).
					Msg("Insufficient role for route")
				util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
					"errorMessage": locale.TR(r, "insufficient_permissions"),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
