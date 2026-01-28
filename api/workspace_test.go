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

type fakeAPIWorkspaces struct {
	db.WorkspaceRepository
	workspaces map[db.Id]*db.Workspace
	roles      map[db.Id]string
	err        error
	updated    []*db.Workspace
	deleted    []db.Id
}

func (f *fakeAPIWorkspaces) CreateWithOwner(workspace *db.Workspace, _ db.Id) error {
	if f.err != nil {
		return f.err
	}

	workspace.Id = db.Id("w-new")
	workspace.Role = db.WorkspaceRoleOwner
	workspace.MembersCount = 1
	f.workspaces[workspace.Id] = workspace

	return nil
}

func (f *fakeAPIWorkspaces) GetById(id db.Id) (*db.Workspace, error) {
	workspace, ok := f.workspaces[id]
	if !ok {
		return nil, f.err
	}

	return workspace, f.err
}

func (f *fakeAPIWorkspaces) GetForUser(id db.Id, _ db.Id) (*db.Workspace, error) {
	workspace, ok := f.workspaces[id]
	if !ok {
		return nil, f.err
	}

	role, ok := f.roles[id]
	if !ok {
		return nil, f.err
	}

	copied := *workspace
	copied.Role = role

	return &copied, f.err
}

func (f *fakeAPIWorkspaces) ListByUser(_ db.Id, _, _ int) ([]*db.Workspace, error) {
	var list []*db.Workspace
	for id := range f.workspaces {
		workspace, _ := f.GetForUser(id, "")
		if workspace != nil {
			list = append(list, workspace)
		}
	}

	return list, f.err
}

func (f *fakeAPIWorkspaces) CountByUser(_ db.Id) (int64, error) {
	return int64(len(f.workspaces)), f.err
}

func (f *fakeAPIWorkspaces) Update(workspace *db.Workspace) error {
	f.updated = append(f.updated, workspace)

	return f.err
}

func (f *fakeAPIWorkspaces) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func TestUnitWorkspaceActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1")}
	stored := &db.Workspace{Id: db.Id("ws-1"), Name: "Platform", MembersCount: 1}

	newAPI := func(workspaces *fakeAPIWorkspaces) *API {
		return &API{Workspace: module.NewWorkspace(workspaces)}
	}

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Post("/api/v1/workspaces", api.CreateWorkspaceAction)
		router.Get("/api/v1/workspaces", api.ListWorkspacesAction)
		router.Get("/api/v1/workspaces/{workspaceId}", api.GetWorkspaceAction)
		router.Put("/api/v1/workspaces/{workspaceId}", api.UpdateWorkspaceAction)
		router.Delete("/api/v1/workspaces/{workspaceId}", api.DeleteWorkspaceAction)

		return router
	}

	serve := func(api *API, withUser bool, method, target, payload string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(payload))
		if withUser {
			req = req.WithContext(middleware.WithUserContext(req.Context(), user))
		}
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, req)

		return w
	}

	newStore := func() *fakeAPIWorkspaces {
		return &fakeAPIWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"ws-1": stored},
			roles:      map[db.Id]string{"ws-1": db.WorkspaceRoleOwner},
		}
	}

	t.Run("create returns 201", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodPost, "/api/v1/workspaces", `{"name":"Platform"}`)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Platform"`)
	})

	t.Run("create returns 400 for a short name", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodPost, "/api/v1/workspaces", `{"name":"x"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		store := newStore()
		store.err = errors.New("boom")

		w := serve(newAPI(store), true, http.MethodPost, "/api/v1/workspaces", `{"name":"Platform"}`)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list returns workspaces with paging meta", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodGet, "/api/v1/workspaces", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Platform"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("get returns a workspace for a member", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodGet, "/api/v1/workspaces/ws-1", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"ws-1"`)
	})

	t.Run("get returns a workspace for an access key", func(t *testing.T) {
		w := serve(newAPI(newStore()), false, http.MethodGet, "/api/v1/workspaces/ws-1", "")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get returns 404 for a non member", func(t *testing.T) {
		store := newStore()
		store.roles = map[db.Id]string{}

		w := serve(newAPI(store), true, http.MethodGet, "/api/v1/workspaces/ws-1", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update renames a workspace", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodPut, "/api/v1/workspaces/ws-1", `{"name":"Renamed"}`)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Renamed"`)
	})

	t.Run("update returns 403 for a non manager", func(t *testing.T) {
		store := newStore()
		store.roles["ws-1"] = db.WorkspaceRoleRegular

		w := serve(newAPI(store), true, http.MethodPut, "/api/v1/workspaces/ws-1", `{"name":"Renamed"}`)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("delete returns 200", func(t *testing.T) {
		w := serve(newAPI(newStore()), true, http.MethodDelete, "/api/v1/workspaces/ws-1", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_deleted_successfully")
	})

	t.Run("delete returns 404 for a non member", func(t *testing.T) {
		store := newStore()
		store.roles = map[db.Id]string{}

		w := serve(newAPI(store), true, http.MethodDelete, "/api/v1/workspaces/ws-1", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
