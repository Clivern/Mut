// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/middleware"
	"github.com/clivern/mut/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type fakeMeAPIKeys struct {
	db.APIKeyRepository
	key *db.APIKey
	err error
}

func (f *fakeMeAPIKeys) GetByKey(_ string) (*db.APIKey, error) {
	return f.key, f.err
}

type fakeMeUsers struct {
	db.UserRepository
	user    *db.User
	updated []*db.User
	err     error
}

func (f *fakeMeUsers) GetById(_ db.Id) (*db.User, error) {
	return f.user, f.err
}

func (f *fakeMeUsers) Update(user *db.User) error {
	f.updated = append(f.updated, user)

	return f.err
}

func TestUnitMeActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1"), Email: "ops@example.com", Name: "Ops", Role: db.UserRoleAdmin}
	key := &db.APIKey{Id: db.Id("key-1"), UserId: db.Id("user-1"), Name: "CI"}

	newAPI := func(keys *fakeMeAPIKeys, users *fakeMeUsers) *API {
		return &API{Me: module.NewMe(keys, users)}
	}

	serve := func(api *API, withUser bool, apiKey string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		if withUser {
			req = req.WithContext(middleware.WithUserContext(req.Context(), user))
		}
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		w := httptest.NewRecorder()
		router := chi.NewRouter()
		router.Get("/api/v1/me", api.GetMeAction)
		router.ServeHTTP(w, req)

		return w
	}

	t.Run("returns the session user", func(t *testing.T) {
		w := serve(newAPI(&fakeMeAPIKeys{}, &fakeMeUsers{}), true, "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"principalType":"session"`)
		assert.Contains(t, w.Body.String(), `"email":"ops@example.com"`)
	})

	t.Run("returns the api key owner", func(t *testing.T) {
		w := serve(newAPI(&fakeMeAPIKeys{key: key}, &fakeMeUsers{user: user}), false, "secret")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"principalType":"api_key"`)
	})

	t.Run("returns 401 for an unknown api key", func(t *testing.T) {
		w := serve(newAPI(&fakeMeAPIKeys{}, &fakeMeUsers{}), false, "nope")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_api_key")
	})

	t.Run("returns 500 when the key lookup fails", func(t *testing.T) {
		w := serve(newAPI(&fakeMeAPIKeys{err: errors.New("boom")}, &fakeMeUsers{}), false, "secret")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("returns 403 without a session or key", func(t *testing.T) {
		w := serve(newAPI(&fakeMeAPIKeys{}, &fakeMeUsers{}), false, "")

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUnitProfileActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1"), Email: "ops@example.com", Name: "Ops", Role: db.UserRoleRegular}

	newAPI := func(users *fakeMeUsers) *API {
		return &API{Profile: module.NewProfile(users)}
	}

	serve := func(api *API, withUser bool, method, target, payload string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(payload))
		if withUser {
			req = req.WithContext(middleware.WithUserContext(req.Context(), user))
		}
		w := httptest.NewRecorder()
		router := chi.NewRouter()
		router.Get("/api/v1/action/profile", api.GetProfileAction)
		router.Put("/api/v1/action/profile", api.UpdateProfileAction)
		router.ServeHTTP(w, req)

		return w
	}

	t.Run("get returns the profile", func(t *testing.T) {
		w := serve(newAPI(&fakeMeUsers{user: user}), true, http.MethodGet, "/api/v1/action/profile", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"ops@example.com"`)
	})

	t.Run("get returns 401 without a user", func(t *testing.T) {
		w := serve(newAPI(&fakeMeUsers{}), false, http.MethodGet, "/api/v1/action/profile", "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		w := serve(newAPI(&fakeMeUsers{err: errors.New("boom")}), true, http.MethodGet, "/api/v1/action/profile", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update returns the updated profile", func(t *testing.T) {
		users := &fakeMeUsers{user: &db.User{Id: "user-1", Email: "ops@example.com", Name: "Ops"}}

		w := serve(
			newAPI(users),
			true,
			http.MethodPut,
			"/api/v1/action/profile",
			`{"name":"Bo","language":"fr","theme":"dark"}`,
		)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Bo"`)
		assert.Contains(t, w.Body.String(), "profile_updated_successfully")
	})

	t.Run("update returns 400 for a bad theme", func(t *testing.T) {
		w := serve(
			newAPI(&fakeMeUsers{user: user}),
			true,
			http.MethodPut,
			"/api/v1/action/profile",
			`{"language":"en","theme":"ghost"}`,
		)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 401 without a user", func(t *testing.T) {
		w := serve(
			newAPI(&fakeMeUsers{}),
			false,
			http.MethodPut,
			"/api/v1/action/profile",
			`{"language":"en","theme":"dark"}`,
		)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
