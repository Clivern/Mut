// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMembers struct {
	db.WorkspaceMemberRepository
	members map[db.Id]*db.WorkspaceMember
	deleted []db.Id
}

func (f *fakeMembers) Create(workspaceId db.Id, userId db.Id, role string) error {
	f.members[userId] = &db.WorkspaceMember{WorkspaceId: workspaceId, UserId: userId, Role: role}

	return nil
}

func (f *fakeMembers) GetByWorkspaceAndUser(_ db.Id, userId db.Id) (*db.WorkspaceMember, error) {
	return f.members[userId], nil
}

func (f *fakeMembers) CountByWorkspaceId(_ db.Id) (int64, error) {
	return int64(len(f.members)), nil
}

func (f *fakeMembers) ListByWorkspaceId(_ db.Id, _, _ int) ([]*db.WorkspaceMember, error) {
	var list []*db.WorkspaceMember
	for _, m := range f.members {
		list = append(list, m)
	}

	return list, nil
}

func (f *fakeMembers) UpdateRole(_ db.Id, userId db.Id, role string) error {
	f.members[userId].Role = role

	return nil
}

func (f *fakeMembers) Delete(_ db.Id, userId db.Id) error {
	f.deleted = append(f.deleted, userId)
	delete(f.members, userId)

	return nil
}

func TestUnitWorkspaceMembers(t *testing.T) {
	owner := &db.User{Id: db.Id("u-owner")}
	target := &db.User{Id: db.Id("u-target"), Email: "target@example.com"}

	newModule := func() (*WorkspaceMembers, *fakeMembers) {
		members := &fakeMembers{members: map[db.Id]*db.WorkspaceMember{
			"u-owner": {UserId: "u-owner", Role: db.WorkspaceRoleOwner},
		}}
		workspaces := &fakeWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"w-1": {Id: "w-1", Name: "Platform"}},
			roles:      map[db.Id]string{"w-1": db.WorkspaceRoleOwner},
		}
		users := &fakeUsers{byId: map[db.Id]*db.User{"u-target": target}}

		return NewWorkspaceMembers(members, workspaces, users), members
	}

	t.Run("update cannot change the owner role", func(t *testing.T) {
		module, _ := newModule()

		_, err := module.UpdateMemberRole("w-1", owner, "u-owner", &UpdateWorkspaceMemberRequest{
			Role: db.WorkspaceRoleAdmin,
		})

		assert.ErrorIs(t, err, ErrWorkspaceOwnerProtected)
	})

	t.Run("remove cannot remove the owner", func(t *testing.T) {
		module, members := newModule()

		err := module.RemoveMember("w-1", owner, "u-owner")

		assert.ErrorIs(t, err, ErrWorkspaceOwnerProtected)
		assert.Contains(t, members.members, db.Id("u-owner"))
	})

	t.Run("remove deletes a regular member", func(t *testing.T) {
		module, members := newModule()
		members.members["u-target"] = &db.WorkspaceMember{UserId: "u-target", Role: db.WorkspaceRoleRegular}

		require.NoError(t, module.RemoveMember("w-1", owner, "u-target"))
		assert.Equal(t, []db.Id{"u-target"}, members.deleted)
	})
}
