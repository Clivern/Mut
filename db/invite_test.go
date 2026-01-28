// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestWorkspaceWithOwner(t *testing.T, database *sql.DB, owner *User) *Workspace {
	t.Helper()

	workspace := &Workspace{Name: "ws-" + uuid.NewString()[:8]}
	require.NoError(t, NewWorkspaceRepository(database).CreateWithOwner(workspace, owner.Id))
	t.Cleanup(func() {
		_ = NewWorkspaceRepository(database).Delete(workspace.Id)
	})

	return workspace
}

func TestIntegrationUserInviteRepository(t *testing.T) {
	database := openTestDB(t)
	inviter := createTestUser(t, database)
	workspace := createTestWorkspaceWithOwner(t, database, inviter)
	repo := NewUserInviteRepository(database)
	email := "inv-" + uuid.NewString()[:8] + "@t.test"
	invite := &UserInvite{
		Email:         email,
		Role:          UserRoleRegular,
		Token:         "inv-" + uuid.NewString(),
		Status:        "pending",
		InviterUserId: inviter.Id,
		WorkspaceId:   workspace.Id,
		ExpiresAt:     time.Now().UTC().Add(time.Hour),
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(invite))

		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, email, got.Email)
	})

	t.Run("pending by email", func(t *testing.T) {
		pending, err := repo.ListPendingByEmail(email)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pending), 1)

		count, err := repo.CountPendingByEmailInWorkspace(workspace.Id, email)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("accept", func(t *testing.T) {
		now := time.Now().UTC()
		require.NoError(t, repo.UpdateStatus(invite.Id, "accepted", &now))

		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		assert.Equal(t, "accepted", got.Status)
	})

	t.Run("mark expired", func(t *testing.T) {
		expired := &UserInvite{
			Email:         "exp-" + uuid.NewString()[:8] + "@t.test",
			Role:          UserRoleRegular,
			Token:         "inv-" + uuid.NewString(),
			Status:        "pending",
			InviterUserId: inviter.Id,
			WorkspaceId:   workspace.Id,
			ExpiresAt:     time.Now().UTC().Add(-time.Minute),
		}
		require.NoError(t, repo.Create(expired))

		marked, err := repo.MarkExpiredAsExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, marked, int64(1))
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 20, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		total, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(len(list)))

		require.NoError(t, repo.Delete(invite.Id))

		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
