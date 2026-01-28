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
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/middleware"
	"github.com/clivern/mut/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type fakeInviteRecords struct {
	db.UserInviteRepository
	item         *db.UserInvite
	list         []*db.UserInvite
	pendingCount int64
	createErr    error
	getErr       error
	deleteErr    error
}

func (f *fakeInviteRecords) Create(invite *db.UserInvite) error {
	return f.createErr
}

func (f *fakeInviteRecords) GetById(_ db.Id) (*db.UserInvite, error) {
	return f.item, f.getErr
}

func (f *fakeInviteRecords) ListByWorkspaceId(_ db.Id, _, _ int) ([]*db.UserInvite, error) {
	return f.list, nil
}

func (f *fakeInviteRecords) CountByWorkspaceId(_ db.Id) (int64, error) {
	return int64(len(f.list)), nil
}

func (f *fakeInviteRecords) ListPendingByEmail(_ string) ([]*db.UserInvite, error) {
	return nil, nil
}

func (f *fakeInviteRecords) MarkExpiredAsExpired() (int64, error) {
	return 0, nil
}

func (f *fakeInviteRecords) CountPendingByEmailInWorkspace(_ db.Id, _ string) (int64, error) {
	return f.pendingCount, nil
}

func (f *fakeInviteRecords) Delete(_ db.Id) error {
	return f.deleteErr
}

type fakeInviteUsers struct {
	db.UserRepository
}

func (f *fakeInviteUsers) GetByEmail(_ string) (*db.User, error) {
	return nil, nil
}

type fakeInviteWorkspaces struct {
	db.WorkspaceRepository
	role string
}

func (f *fakeInviteWorkspaces) GetForUser(id db.Id, _ db.Id) (*db.Workspace, error) {
	if f.role == "" {
		return nil, nil
	}

	return &db.Workspace{Id: id, Name: "Platform", Role: f.role}, nil
}

type fakeInviteMembers struct {
	db.WorkspaceMemberRepository
}

func (f *fakeInviteMembers) GetByWorkspaceAndUser(_ db.Id, _ db.Id) (*db.WorkspaceMember, error) {
	return nil, nil
}

type fakeInviteMailer struct {
	sent int
	err  error
}

func (f *fakeInviteMailer) SendInviteEmail(_, _, _ string) error {
	f.sent++

	return f.err
}

func TestUnitInviteActions(t *testing.T) {
	user := &db.User{Id: db.Id("user-1")}
	stored := &db.UserInvite{
		Id:            db.Id("invite-1"),
		Email:         "bo@example.com",
		Role:          db.UserRoleRegular,
		Status:        module.InviteStatusPending,
		InviterUserId: db.Id("user-1"),
		WorkspaceId:   db.Id("ws-1"),
		ExpiresAt:     time.Now().Add(time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	body := `{"email":"bo@example.com","role":"regular"}`

	newAPI := func(role string, invites *fakeInviteRecords, mailer *fakeInviteMailer) *API {
		return &API{
			Invite: module.NewInvite(
				invites,
				&fakeInviteUsers{},
				&fakeInviteWorkspaces{role: role},
				&fakeInviteMembers{},
				mailer,
			),
		}
	}

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Post("/api/v1/workspaces/{workspaceId}/invites", api.CreateInviteAction)
		router.Get("/api/v1/workspaces/{workspaceId}/invites", api.ListInvitesAction)
		router.Get("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", api.GetInviteAction)
		router.Delete("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", api.DeleteInviteAction)

		return router
	}

	serve := func(api *API, method, target, payload string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(payload))
		req = req.WithContext(middleware.WithUserContext(req.Context(), user))
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, req)

		return w
	}

	t.Run("create returns 201 as pending and sends the invite email", func(t *testing.T) {
		mailer := &fakeInviteMailer{}

		w := serve(newAPI("owner", &fakeInviteRecords{}, mailer), http.MethodPost, "/api/v1/workspaces/ws-1/invites", body)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"pending"`)
		assert.Equal(t, 1, mailer.sent)
	})

	t.Run("create still returns 201 when the email fails to send", func(t *testing.T) {
		mailer := &fakeInviteMailer{err: errors.New("smtp down")}

		w := serve(newAPI("admin", &fakeInviteRecords{}, mailer), http.MethodPost, "/api/v1/workspaces/ws-1/invites", body)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("create returns 400 for an invalid email", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{}, &fakeInviteMailer{}),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/invites",
			`{"email":"nope","role":"regular"}`,
		)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 409 when a pending invite already exists", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{pendingCount: 1}, &fakeInviteMailer{}),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/invites",
			body,
		)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "pending_invite_already_exists")
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{createErr: errors.New("boom")}, &fakeInviteMailer{}),
			http.MethodPost,
			"/api/v1/workspaces/ws-1/invites",
			body,
		)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list returns invites with paging meta", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{list: []*db.UserInvite{stored}}, &fakeInviteMailer{}),
			http.MethodGet,
			"/api/v1/workspaces/ws-1/invites",
			"",
		)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"bo@example.com"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("get returns one invite", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{item: stored}, &fakeInviteMailer{}),
			http.MethodGet,
			"/api/v1/workspaces/ws-1/invites/invite-1",
			"",
		)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"role":"regular"`)
	})

	t.Run("get returns 404 for an unknown invite", func(t *testing.T) {
		w := serve(newAPI("owner", &fakeInviteRecords{}, &fakeInviteMailer{}), http.MethodGet, "/api/v1/workspaces/ws-1/invites/invite-1", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "invite_not_found")
	})

	t.Run("delete returns 204", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{item: stored}, &fakeInviteMailer{}),
			http.MethodDelete,
			"/api/v1/workspaces/ws-1/invites/invite-1",
			"",
		)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for an unknown invite", func(t *testing.T) {
		w := serve(newAPI("owner", &fakeInviteRecords{}, &fakeInviteMailer{}), http.MethodDelete, "/api/v1/workspaces/ws-1/invites/invite-1", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		w := serve(
			newAPI("owner", &fakeInviteRecords{item: stored, deleteErr: errors.New("boom")}, &fakeInviteMailer{}),
			http.MethodDelete,
			"/api/v1/workspaces/ws-1/invites/invite-1",
			"",
		)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
