// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/mut/pkg/github"
	"github.com/clivern/mut/pkg/google"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitOAuthActions(t *testing.T) {
	newAPI := func() *API {
		return &API{
			GitHub: github.NewOAuth(github.OAuthConfig{
				ClientID:     "github-id",
				ClientSecret: "github-secret",
				RedirectURL:  "http://localhost:8000/api/v1/public/action/oauth/github/callback",
			}),
			Google: google.NewOAuth(google.OAuthConfig{
				ClientID:     "google-id",
				ClientSecret: "google-secret",
				RedirectURL:  "http://localhost:8000/api/v1/public/action/oauth/google/callback",
			}),
		}
	}

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/public/action/oauth/github", api.GitHubOAuthStartAction)
		router.Get("/api/v1/public/action/oauth/github/callback", api.GitHubOAuthCallbackAction)
		router.Get("/api/v1/public/action/oauth/google", api.GoogleOAuthStartAction)
		router.Get("/api/v1/public/action/oauth/google/callback", api.GoogleOAuthCallbackAction)

		return router
	}

	serve := func(api *API, target string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, req)

		return w
	}

	t.Run("github start redirects with a state cookie", func(t *testing.T) {
		w := serve(newAPI(), "/api/v1/public/action/oauth/github")

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "github.com/login/oauth/authorize")
		assert.Contains(t, w.Header().Get("Set-Cookie"), OAuthStateCookie+"=")
	})

	t.Run("google start redirects with a state cookie", func(t *testing.T) {
		w := serve(newAPI(), "/api/v1/public/action/oauth/google")

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "accounts.google.com")
		assert.Contains(t, w.Header().Get("Set-Cookie"), OAuthStateCookie+"=")
	})

	t.Run("github callback rejects a mismatched state", func(t *testing.T) {
		w := serve(
			newAPI(),
			"/api/v1/public/action/oauth/github/callback?code=abc&state=forged",
			&http.Cookie{Name: OAuthStateCookie, Value: "real"},
		)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login?oauth_error=github")
	})

	t.Run("google callback rejects a mismatched state", func(t *testing.T) {
		w := serve(
			newAPI(),
			"/api/v1/public/action/oauth/google/callback?code=abc&state=forged",
			&http.Cookie{Name: OAuthStateCookie, Value: "real"},
		)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login?oauth_error=google")
	})
}
