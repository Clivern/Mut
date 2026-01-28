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

func (f *fakeWorkspaces) GetById(id db.Id) (*db.Workspace, error) {
	workspace, ok := f.workspaces[id]
	if !ok {
		return nil, f.err
	}

	copied := *workspace

	return &copied, f.err
}

func TestUnitPerm(t *testing.T) {
	workspace := &db.Workspace{Id: db.Id("w-1"), Name: "Platform"}
	user := &db.User{Id: db.Id("u-1"), IsActive: true}
	key := &db.AccessKey{WorkspaceId: db.Id("w-1"), Permissions: AccessKeyPermissionWorkspaceRead}
	otherKey := &db.AccessKey{WorkspaceId: db.Id("w-2"), Permissions: AccessKeyPermissionWorkspaceRead}

	t.Run("member can read workspace", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{"w-1": workspace}}

		allowed, err := NewPerm(repo).WithUser(user).WithWorkspaceId("w-1").Can(AccessKeyPermissionWorkspaceRead)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("member can list members", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{"w-1": workspace}}

		allowed, err := NewPerm(repo).WithUser(user).WithWorkspace(workspace).Can(AccessKeyPermissionMembersRead)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("non member cannot read", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{}}

		allowed, err := NewPerm(repo).WithUser(user).WithWorkspace(workspace).Can(AccessKeyPermissionWorkspaceRead)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("access key with permission can read", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{"w-1": workspace}}

		allowed, err := NewPerm(repo).WithAccessKey(key).WithWorkspace(workspace).Can(AccessKeyPermissionWorkspaceRead)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("access key without permission cannot list members", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{"w-1": workspace}}

		allowed, err := NewPerm(repo).WithAccessKey(key).WithWorkspace(workspace).Can(AccessKeyPermissionMembersRead)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("access key from another workspace cannot read", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{"w-1": workspace}}

		allowed, err := NewPerm(repo).WithAccessKey(otherKey).WithWorkspace(workspace).Can(AccessKeyPermissionWorkspaceRead)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("missing workspace", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{}}

		_, err := NewPerm(repo).WithUser(user).Can(AccessKeyPermissionWorkspaceRead)
		assert.ErrorIs(t, err, ErrPermissionWorkspaceMissing)
	})
}

func TestUnitPermManagementRequiresOwnerOrAdmin(t *testing.T) {
	workspace := &db.Workspace{Id: db.Id("w-1"), Name: "Platform"}
	user := &db.User{Id: db.Id("u-1"), IsActive: true}
	manage := []string{PermissionMembersManage, PermissionAccessKeysManage, PermissionClustersManage}

	cases := map[string]bool{
		db.WorkspaceRoleOwner:   true,
		db.WorkspaceRoleAdmin:   true,
		db.WorkspaceRoleRegular: false,
	}
	for role, want := range cases {
		repo := &fakeWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"w-1": workspace},
			roles:      map[db.Id]string{"w-1": role},
		}

		for _, permission := range manage {
			allowed, err := NewPerm(repo).WithUser(user).WithWorkspace(workspace).Can(permission)
			require.NoError(t, err)
			assert.Equal(t, want, allowed, "role %s, permission %s", role, permission)
		}
	}
}

func TestUnitPermAccessKeysNeverManage(t *testing.T) {
	workspace := &db.Workspace{Id: db.Id("w-1"), Name: "Platform"}
	key := &db.AccessKey{WorkspaceId: db.Id("w-1"), Permissions: AccessKeyPermissionClustersWrite}

	for _, permission := range []string{PermissionMembersManage, PermissionAccessKeysManage, PermissionClustersManage} {
		allowed, err := NewPerm(&fakeWorkspaces{}).WithAccessKey(key).WithWorkspace(workspace).Can(permission)
		require.NoError(t, err)
		assert.False(t, allowed, permission)
	}
}
