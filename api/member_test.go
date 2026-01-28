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

type fakeAPIMembers struct {
	db.WorkspaceMemberRepository
	members []*db.WorkspaceMember
	err     error
	updated []string
	deleted []string
}

func (f *fakeAPIMembers) GetByWorkspaceAndUser(_ db.Id, userId db.Id) (*db.WorkspaceMember, error) {
	for _, member := range f.members {
		if member.UserId == userId {
			return member, f.err
		}
	}

	return nil, f.err
}

func (f *fakeAPIMembers) ListByWorkspaceId(_ db.Id, _, _ int) ([]*db.WorkspaceMember, error) {
	return f.members, f.err
}

func (f *fakeAPIMembers) CountByWorkspaceId(_ db.Id) (int64, error) {
	return int64(len(f.members)), f.err
}

func (f *fakeAPIMembers) UpdateRole(_ db.Id, userId db.Id, role string) error {
	f.updated = append(f.updated, role)

	for _, member := range f.members {
		if member.UserId == userId {
			member.Role = role
		}
	}

	return f.err
}

func (f *fakeAPIMembers) Delete(_ db.Id, userId db.Id) error {
	f.deleted = append(f.deleted, userId.String())

	return f.err
}

type fakeAPIMemberUsers struct {
	db.UserRepository
}

func TestUnitMemberActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1")}
	stored := &db.WorkspaceMember{
		WorkspaceId: db.Id("ws-1"),
		UserId:      db.Id("user-2"),
		Name:        "Bo",
		Email:       "bo@example.com",
		Role:        db.WorkspaceRoleRegular,
	}

	newAPI := func(role string, members *fakeAPIMembers) *API {
		roles := map[db.Id]string{}
		if role != "" {
			roles["ws-1"] = role
		}
		workspaces := &fakeAPIWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"ws-1": {Id: "ws-1", Name: "Platform"}},
			roles:      roles,
		}

		return &API{WorkspaceMembers: module.NewWorkspaceMembers(members, workspaces, &fakeAPIMemberUsers{})}
	}

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/workspaces/{workspaceId}/members", api.ListWorkspaceMembersAction)
		router.Put("/api/v1/workspaces/{workspaceId}/members/{memberUserId}", api.UpdateWorkspaceMemberAction)
		router.Delete("/api/v1/workspaces/{workspaceId}/members/{memberUserId}", api.RemoveWorkspaceMemberAction)

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

	newMembers := func() *fakeAPIMembers {
		return &fakeAPIMembers{members: []*db.WorkspaceMember{stored}}
	}

	t.Run("list returns members with paging meta", func(t *testing.T) {
		w := serve(newAPI("owner", newMembers()), true, http.MethodGet, "/api/v1/workspaces/ws-1/members", "")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"bo@example.com"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list works for an access key", func(t *testing.T) {
		w := serve(newAPI("owner", newMembers()), false, http.MethodGet, "/api/v1/workspaces/ws-1/members", "")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list returns 404 for a non member", func(t *testing.T) {
		w := serve(newAPI("", newMembers()), true, http.MethodGet, "/api/v1/workspaces/ws-1/members", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update changes a member role", func(t *testing.T) {
		w := serve(
			newAPI("owner", newMembers()),
			true,
			http.MethodPut,
			"/api/v1/workspaces/ws-1/members/user-2",
			`{"role":"admin"}`,
		)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"role":"admin"`)
	})

	t.Run("update returns 400 for a bad role", func(t *testing.T) {
		w := serve(
			newAPI("owner", newMembers()),
			true,
			http.MethodPut,
			"/api/v1/workspaces/ws-1/members/user-2",
			`{"role":"ghost"}`,
		)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 404 for an unknown member", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeAPIMembers{}),
			true,
			http.MethodPut,
			"/api/v1/workspaces/ws-1/members/user-9",
			`{"role":"admin"}`,
		)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("remove returns 204", func(t *testing.T) {
		w := serve(newAPI("owner", newMembers()), true, http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", "")

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("remove returns 403 for the workspace owner", func(t *testing.T) {
		members := &fakeAPIMembers{members: []*db.WorkspaceMember{{
			WorkspaceId: db.Id("ws-1"),
			UserId:      db.Id("user-3"),
			Role:        db.WorkspaceRoleOwner,
		}}}

		w := serve(newAPI("owner", members), true, http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-3", "")

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("remove returns 404 for an unknown member", func(t *testing.T) {
		w := serve(newAPI("owner", &fakeAPIMembers{}), true, http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-9", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("remove returns 500 when the repository fails", func(t *testing.T) {
		members := newMembers()
		members.err = errors.New("boom")

		w := serve(newAPI("owner", members), true, http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
