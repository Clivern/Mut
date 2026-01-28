// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"
	"time"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserInvites struct {
	db.UserInviteRepository
	created       []*db.UserInvite
	item          *db.UserInvite
	list          []*db.UserInvite
	pending       []*db.UserInvite
	pendingCount  int64
	total         int64
	deleted       []db.Id
	statusUpdates []string
	err           error
}

func (f *fakeUserInvites) Create(invite *db.UserInvite) error {
	if f.err != nil {
		return f.err
	}

	invite.Id = db.Id("inv-new")
	f.created = append(f.created, invite)

	return nil
}

func (f *fakeUserInvites) GetById(_ db.Id) (*db.UserInvite, error) {
	return f.item, f.err
}

func (f *fakeUserInvites) ListByWorkspaceId(_ db.Id, _, _ int) ([]*db.UserInvite, error) {
	return f.list, f.err
}

func (f *fakeUserInvites) ListPendingByEmail(_ string) ([]*db.UserInvite, error) {
	return f.pending, f.err
}

func (f *fakeUserInvites) UpdateStatus(id db.Id, status string, _ *time.Time) error {
	f.statusUpdates = append(f.statusUpdates, status)

	return f.err
}

func (f *fakeUserInvites) MarkExpiredAsExpired() (int64, error) {
	return 0, f.err
}

func (f *fakeUserInvites) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeUserInvites) CountByWorkspaceId(_ db.Id) (int64, error) {
	return f.total, f.err
}

func (f *fakeUserInvites) CountPendingByEmailInWorkspace(_ db.Id, _ string) (int64, error) {
	return f.pendingCount, f.err
}

type fakeInviteMailer struct {
	sent []string
	err  error
}

func (f *fakeInviteMailer) SendInviteEmail(to, _, _ string) error {
	f.sent = append(f.sent, to)

	return f.err
}

func TestUnitInvite(t *testing.T) {
	inviter := &db.User{Id: db.Id("u-owner"), Email: "owner@example.com"}

	// newInviteModule wires the module with workspace ws-1 where the inviter is an owner.
	newInviteModule := func(invites *fakeUserInvites, users *fakeUsers, members *fakeMembers, mailer *fakeInviteMailer) *Invite {
		workspaces := &fakeWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"ws-1": {Id: "ws-1", Name: "Platform"}},
			roles:      map[db.Id]string{"ws-1": db.WorkspaceRoleOwner},
		}

		return NewInvite(invites, users, workspaces, members, mailer)
	}

	t.Run("create rejects an email that is already a member", func(t *testing.T) {
		users := &fakeUsers{byId: map[db.Id]*db.User{"u-2": {Id: "u-2", Email: "ada@example.com"}}}
		members := &fakeMembers{members: map[db.Id]*db.WorkspaceMember{"u-2": {UserId: "u-2"}}}

		_, err := newInviteModule(&fakeUserInvites{}, users, members, &fakeInviteMailer{}).
			CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "ada@example.com", Role: "regular"})
		assert.ErrorIs(t, err, ErrUserAlreadyInWorkspace)
	})

	t.Run("create rejects a second pending invite for the same email", func(t *testing.T) {
		invites := &fakeUserInvites{pendingCount: 1}

		_, err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "ada@example.com", Role: "regular"})
		assert.ErrorIs(t, err, ErrPendingInviteExists)
	})

	t.Run("create stores a pending invite, lowercases the email and sends the mail", func(t *testing.T) {
		invites := &fakeUserInvites{}
		mailer := &fakeInviteMailer{}

		response, err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, mailer).
			CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "  Ada@Example.com ", Role: "regular"})
		require.NoError(t, err)
		require.Len(t, invites.created, 1)
		assert.Equal(t, "ada@example.com", invites.created[0].Email)
		assert.Equal(t, InviteStatusPending, response.Status)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, []string{"ada@example.com"}, mailer.sent)
		assert.WithinDuration(t, time.Now().UTC().Add(7*24*time.Hour), invites.created[0].ExpiresAt, 5*time.Minute)
	})

	t.Run("create still succeeds when the mail cannot be sent", func(t *testing.T) {
		invites := &fakeUserInvites{}

		_, err := newInviteModule(
			invites,
			&fakeUsers{},
			&fakeMembers{members: map[db.Id]*db.WorkspaceMember{}},
			&fakeInviteMailer{err: errors.New("smtp down")},
		).CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "ada@example.com", Role: "regular"})
		require.NoError(t, err)
		assert.Len(t, invites.created, 1)
	})

	t.Run("create accepts the invite for an existing user who is not a member yet", func(t *testing.T) {
		users := &fakeUsers{byId: map[db.Id]*db.User{"u-2": {Id: "u-2", Email: "ada@example.com"}}}

		response, err := newInviteModule(
			&fakeUserInvites{},
			users,
			&fakeMembers{members: map[db.Id]*db.WorkspaceMember{}},
			&fakeInviteMailer{},
		).CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "ada@example.com", Role: "regular"})
		require.NoError(t, err)
		assert.Equal(t, InviteStatusAccepted, response.Status)
	})

	t.Run("create wraps the invite repository error", func(t *testing.T) {
		_, err := newInviteModule(
			&fakeUserInvites{err: errors.New("boom")},
			&fakeUsers{},
			&fakeMembers{members: map[db.Id]*db.WorkspaceMember{}},
			&fakeInviteMailer{},
		).CreateInvite("ws-1", inviter, &CreateInviteRequest{Email: "ada@example.com", Role: "regular"})
		assert.ErrorIs(t, err, ErrFailedCreateInvite)
	})

	t.Run("list returns the invites and the total", func(t *testing.T) {
		invites := &fakeUserInvites{
			list:  []*db.UserInvite{{Id: "inv-1", Email: "ada@example.com", Status: InviteStatusPending}},
			total: 5,
		}

		result, err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			ListInvites("ws-1", inviter, 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		require.Len(t, result.Invites, 1)
		assert.Nil(t, result.Invites[0].AcceptedAt)
	})

	t.Run("get hides invites from another workspace", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-2"}}

		_, err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			GetInvite("ws-1", inviter, "inv-1")
		assert.ErrorIs(t, err, ErrInviteNotFound)
	})

	t.Run("get returns the accepted time when set", func(t *testing.T) {
		acceptedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-1", AcceptedAt: &acceptedAt}}

		response, err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			GetInvite("ws-1", inviter, "inv-1")
		require.NoError(t, err)
		require.NotNil(t, response.AcceptedAt)
		assert.Equal(t, "2026-02-03T04:05:06Z", *response.AcceptedAt)
	})

	t.Run("delete returns not found for an invite from another workspace", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-2"}}

		err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			DeleteInvite("ws-1", inviter, "inv-1")
		assert.ErrorIs(t, err, ErrInviteNotFound)
	})

	t.Run("delete removes the invite", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-1"}}

		require.NoError(
			t,
			newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
				DeleteInvite("ws-1", inviter, "inv-1"),
		)
		assert.Equal(t, []db.Id{"inv-1"}, invites.deleted)
	})

	t.Run("attach pending adds the user to each workspace and accepts the invites", func(t *testing.T) {
		invites := &fakeUserInvites{pending: []*db.UserInvite{
			{Id: "inv-1", WorkspaceId: "ws-1", Role: "regular"},
			{Id: "inv-2", WorkspaceId: "ws-2", Role: "admin"},
		}}
		members := &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}

		require.NoError(t, newInviteModule(invites, &fakeUsers{}, members, &fakeInviteMailer{}).
			AttachPending(&db.User{Id: "u-2", Email: "ada@example.com"}))
		assert.Equal(t, db.WorkspaceRoleRegular, members.members["u-2"].Role)
		assert.Equal(t, []string{InviteStatusAccepted, InviteStatusAccepted}, invites.statusUpdates)
	})

	t.Run("attach pending skips workspaces the user already belongs to", func(t *testing.T) {
		invites := &fakeUserInvites{pending: []*db.UserInvite{{Id: "inv-1", WorkspaceId: "ws-1", Role: "regular"}}}
		members := &fakeMembers{members: map[db.Id]*db.WorkspaceMember{"u-2": {UserId: "u-2", Role: db.WorkspaceRoleAdmin}}}

		require.NoError(t, newInviteModule(invites, &fakeUsers{}, members, &fakeInviteMailer{}).
			AttachPending(&db.User{Id: "u-2", Email: "ada@example.com"}))
		assert.Equal(t, db.WorkspaceRoleAdmin, members.members["u-2"].Role)
		assert.Equal(t, []string{InviteStatusAccepted}, invites.statusUpdates)
	})

	t.Run("attach pending wraps the repository error", func(t *testing.T) {
		invites := &fakeUserInvites{err: errors.New("boom")}

		err := newInviteModule(invites, &fakeUsers{}, &fakeMembers{members: map[db.Id]*db.WorkspaceMember{}}, &fakeInviteMailer{}).
			AttachPending(&db.User{Id: "u-2"})
		assert.ErrorIs(t, err, ErrFailedAttachInvites)
	})
}
