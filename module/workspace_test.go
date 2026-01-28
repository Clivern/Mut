// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWorkspaces struct {
	db.WorkspaceRepository
	workspaces map[db.Id]*db.Workspace
	roles      map[db.Id]string
	deleted    []db.Id
	updated    []*db.Workspace
	err        error
}

func (f *fakeWorkspaces) CreateWithOwner(workspace *db.Workspace, ownerId db.Id) error {
	if f.err != nil {
		return f.err
	}

	workspace.Id = db.Id("w-new")
	f.workspaces[workspace.Id] = workspace
	f.roles[workspace.Id] = db.WorkspaceRoleOwner

	return nil
}

func (f *fakeWorkspaces) GetForUser(id db.Id, _ db.Id) (*db.Workspace, error) {
	workspace, ok := f.workspaces[id]
	if !ok {
		return nil, f.err
	}

	copied := *workspace
	copied.Role = f.roles[id]
	copied.MembersCount = 1

	return &copied, f.err
}

func (f *fakeWorkspaces) ListByUser(_ db.Id, _, _ int) ([]*db.Workspace, error) {
	var list []*db.Workspace
	for id := range f.workspaces {
		workspace, _ := f.GetForUser(id, "")
		list = append(list, workspace)
	}

	return list, f.err
}

func (f *fakeWorkspaces) CountByUser(_ db.Id) (int64, error) {
	return int64(len(f.workspaces)), f.err
}

func (f *fakeWorkspaces) Update(workspace *db.Workspace) error {
	f.updated = append(f.updated, workspace)

	return f.err
}

func (f *fakeWorkspaces) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func TestUnitWorkspace(t *testing.T) {
	userId := db.Id("u-1")
	req := &WorkspaceRequest{Name: "Platform"}

	newFake := func() *fakeWorkspaces {
		return &fakeWorkspaces{
			workspaces: map[db.Id]*db.Workspace{
				"w-1": {Id: "w-1", Name: "Existing"},
			},
			roles: map[db.Id]string{"w-1": db.WorkspaceRoleRegular},
		}
	}

	t.Run("create makes the caller the owner", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{}, roles: map[db.Id]string{}}
		workspace := NewWorkspace(repo)

		created, err := workspace.CreateWorkspace(userId, req)

		require.NoError(t, err)
		assert.Equal(t, "Platform", created.Name)
		assert.Equal(t, db.WorkspaceRoleOwner, created.Role)
		assert.Equal(t, int64(1), created.MembersCount)
	})

	t.Run("create fails when the repository fails", func(t *testing.T) {
		repo := &fakeWorkspaces{workspaces: map[db.Id]*db.Workspace{}, roles: map[db.Id]string{}, err: errors.New("boom")}

		_, err := NewWorkspace(repo).CreateWorkspace(userId, req)

		assert.Error(t, err)
	})

	t.Run("list returns the page and total", func(t *testing.T) {
		list, err := NewWorkspace(newFake()).ListWorkspaces(userId, 50, 0)

		require.NoError(t, err)
		assert.Len(t, list.Workspaces, 1)
		assert.Equal(t, int64(1), list.Total)
	})

	t.Run("get rejects a workspace the user is not a member of", func(t *testing.T) {
		_, err := NewWorkspace(newFake()).GetWorkspace(userId, "w-missing")

		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("update forbids regular and readonly members", func(t *testing.T) {
		repo := newFake()

		_, err := NewWorkspace(repo).UpdateWorkspace(userId, "w-1", req)

		assert.ErrorIs(t, err, ErrWorkspaceForbidden)
		assert.Empty(t, repo.updated)
	})

	t.Run("update renames the workspace for an admin", func(t *testing.T) {
		repo := newFake()
		repo.roles["w-1"] = db.WorkspaceRoleAdmin

		updated, err := NewWorkspace(repo).UpdateWorkspace(userId, "w-1", req)

		require.NoError(t, err)
		assert.Equal(t, "Platform", updated.Name)
		require.Len(t, repo.updated, 1)
	})

	t.Run("delete forbids regular members", func(t *testing.T) {
		repo := newFake()

		assert.ErrorIs(t, NewWorkspace(repo).DeleteWorkspace(userId, "w-1"), ErrWorkspaceForbidden)
		assert.Empty(t, repo.deleted)
	})

	t.Run("delete removes the workspace for its owner", func(t *testing.T) {
		repo := newFake()
		repo.roles["w-1"] = db.WorkspaceRoleOwner

		require.NoError(t, NewWorkspace(repo).DeleteWorkspace(userId, "w-1"))
		assert.Equal(t, []db.Id{"w-1"}, repo.deleted)
	})
}
