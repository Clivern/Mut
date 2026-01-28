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
	"github.com/stretchr/testify/require"
)

type fakeAPIAccessKeys struct {
	db.AccessKeyRepository
	keys    map[db.Id]*db.AccessKey
	err     error
	deleted []db.Id
}

func (f *fakeAPIAccessKeys) Create(key *db.AccessKey) error {
	if f.err != nil {
		return f.err
	}

	key.Id = db.Id("k-new")
	f.keys[key.Id] = key

	return nil
}

func (f *fakeAPIAccessKeys) GetById(id db.Id) (*db.AccessKey, error) {
	return f.keys[id], f.err
}

func (f *fakeAPIAccessKeys) ListByWorkspaceId(workspaceId db.Id, _, _ int) ([]*db.AccessKey, error) {
	var list []*db.AccessKey
	for _, key := range f.keys {
		if key.WorkspaceId == workspaceId {
			list = append(list, key)
		}
	}

	return list, f.err
}

func (f *fakeAPIAccessKeys) CountByWorkspaceId(_ db.Id) (int64, error) {
	return int64(len(f.keys)), f.err
}

func (f *fakeAPIAccessKeys) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func TestUnitAccessKeyActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1")}
	stored := &db.AccessKey{
		Id:          db.Id("k-1"),
		WorkspaceId: db.Id("ws-1"),
		Name:        "CI",
		KeySuffix:   "abcd",
		Permissions: module.AccessKeyPermissionWorkspaceRead,
	}

	newAPI := func(role string, keys *fakeAPIAccessKeys) *API {
		roles := map[db.Id]string{}
		if role != "" {
			roles["ws-1"] = role
		}
		workspaces := &fakeAPIWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"ws-1": {Id: "ws-1", Name: "Platform"}},
			roles:      roles,
		}

		return &API{AccessKey: module.NewAccessKey(keys, workspaces)}
	}

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Post("/api/v1/workspaces/{workspaceId}/accessKeys", api.CreateWorkspaceAccessKeyAction)
		router.Get("/api/v1/workspaces/{workspaceId}/accessKeys", api.ListWorkspaceAccessKeysAction)
		router.Get("/api/v1/workspaces/{workspaceId}/accessKeys/{keyId}", api.GetWorkspaceAccessKeyAction)
		router.Delete("/api/v1/workspaces/{workspaceId}/accessKeys/{keyId}", api.DeleteWorkspaceAccessKeyAction)

		return router
	}

	serve := func(api *API, method, target, payload string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(payload))
		req = req.WithContext(middleware.WithUserContext(req.Context(), user))
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, req)

		return w
	}

	newKeys := func() *fakeAPIAccessKeys {
		return &fakeAPIAccessKeys{keys: map[db.Id]*db.AccessKey{"k-1": stored}}
	}

	t.Run("create returns 201 with the raw key", func(t *testing.T) {
		w := serve(
			newAPI("owner", newKeys()),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/accessKeys",
			`{"name":"CI","permissions":["workspace:read"]}`,
		)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"key":"mut_`)
	})

	t.Run("create returns 400 for a bad permission", func(t *testing.T) {
		w := serve(
			newAPI("owner", newKeys()),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/accessKeys",
			`{"name":"CI","permissions":["root"]}`,
		)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		keys := newKeys()
		keys.err = errors.New("boom")

		w := serve(
			newAPI("owner", keys),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/accessKeys",
			`{"name":"CI","permissions":["workspace:read"]}`,
		)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list returns keys with paging meta", func(t *testing.T) {
		w := serve(newAPI("owner", newKeys()), http.MethodGet, "/api/v1/workspaces/ws-1/accessKeys", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"CI"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("get returns one key", func(t *testing.T) {
		w := serve(newAPI("owner", newKeys()), http.MethodGet, "/api/v1/workspaces/ws-1/accessKeys/k-1", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"k-1"`)
	})

	t.Run("get returns 404 for a key from another workspace", func(t *testing.T) {
		keys := newKeys()
		keys.keys["k-9"] = &db.AccessKey{Id: "k-9", WorkspaceId: "ws-9", Permissions: "workspace:read"}

		w := serve(newAPI("owner", keys), http.MethodGet, "/api/v1/workspaces/ws-1/accessKeys/k-9", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "access_key_not_found")
	})

	t.Run("delete returns 204", func(t *testing.T) {
		keys := newKeys()

		w := serve(newAPI("owner", keys), http.MethodDelete, "/api/v1/workspaces/ws-1/accessKeys/k-1", "")

		require.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, []db.Id{"k-1"}, keys.deleted)
	})

	t.Run("delete returns 404 for a missing key", func(t *testing.T) {
		w := serve(newAPI("owner", newKeys()), http.MethodDelete, "/api/v1/workspaces/ws-1/accessKeys/k-9", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
