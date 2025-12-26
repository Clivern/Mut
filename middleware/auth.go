// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Auth authenticates the request via API key or session cookie.
func Auth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ShouldSkipAuth(r.URL.Path) {
				if !IsAssetPath(r.URL.Path) {
					log.Info().
						Str("path", r.URL.Path).
						Msg("Skipping auth for public route")
				}

				next.ServeHTTP(w, r)
				return
			}

			// API Key Check
			apiKey := r.Header.Get("X-API-Key")
			if lo.IsNotEmpty(apiKey) {
				user, err := db.NewUserRepository(db.GetDB()).GetByAPIKey(apiKey)
				if err != nil || user == nil || !user.IsActive {
					log.Info().
						Err(err).
						Str("path", r.URL.Path).
						Msg("API key validation failed")
					util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
						"errorMessage": locale.TR(r, "invalid_api_key"),
					})
					return
				}

				log.Info().
					Str("path", r.URL.Path).
					Msg("API key validated")

				ctx := WithUserContext(r.Context(), user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Session Token Check
			sessionToken := util.GetCookie(r, "_mut_session")
			if lo.IsEmpty(sessionToken) {
				log.Info().
					Str("path", r.URL.Path).
					Msg("No session cookie found")
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": locale.TR(r, "not_authenticated"),
				})
				return
			}

			sessionManager := module.NewSessionManager(
				db.NewSessionRepository(db.GetDB()),
				db.NewUserRepository(db.GetDB()),
			)

			user, _, err := sessionManager.ValidateSession(sessionToken)
			if err != nil {
				util.DeleteCookie(w, "_mut_session")
				if user != nil {
					sessionManager.RevokeUserSessions(user.Id)
				}

				log.Info().
					Err(err).
					Str("path", r.URL.Path).
					Msg("Session validation failed")
				util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"errorMessage": locale.TR(r, "invalid_or_expired_session"),
				})
				return
			}

			log.Info().
				Str("path", r.URL.Path).
				Msg("Session validated")

			ctx := WithUserContext(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BasicAuth creates a basic authentication middleware
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1

			if !ok || !userMatch || !passMatch {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
