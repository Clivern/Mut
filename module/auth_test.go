// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"testing"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAuth(t *testing.T) {
	newAuth := func(role string, active bool, maintenance string) (*Auth, *fakeUsers, *fakeSessions) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Email: "ops@example.com", Role: role, IsActive: active},
		}}
		sessions := &fakeSessions{}
		configs := &fakeConfigs{values: map[string]string{"maintenance_mode": maintenance}}

		return NewAuth(users, sessions, configs), users, sessions
	}

	t.Run("oauth login links a local user on first login", func(t *testing.T) {
		auth, users, sessions := newAuth(db.UserRoleRegular, true, "0")
		ctx := WithAppCtx(context.Background(), NewAppCtx("req-1", "curl", "10.0.0.1", nil))

		result, err := auth.LoginWithOAuth(ctx, &OAuthIdentity{
			Provider:       db.UserProviderGithub,
			ProviderUserID: "github-1",
			Email:          "ops@example.com",
			Name:           "Ops",
		})
		require.NoError(t, err)

		assert.Equal(t, db.Id("u-1"), result.User.Id)
		assert.Equal(t, "Ops", users.byId["u-1"].Name)
		assert.Equal(t, db.UserProviderGithub, users.byId["u-1"].Provider)
		require.Len(t, sessions.created, 1)
		assert.Equal(t, "10.0.0.1", *sessions.created[0].IPAddress)
		assert.Equal(t, "curl", *sessions.created[0].UserAgent)
		assert.Equal(t, []db.Id{"u-1"}, users.lastLogin)
		assert.Equal(t, 0, result.CookieOptions.MaxAge)
	})

	t.Run("oauth login finds returning users by provider", func(t *testing.T) {
		auth, users, sessions := newAuth(db.UserRoleRegular, true, "0")
		providerUserID := "github-1"
		users.byId["u-1"].Provider = db.UserProviderGithub
		users.byId["u-1"].ProviderUserId = &providerUserID

		result, err := auth.LoginWithOAuth(context.Background(), &OAuthIdentity{
			Provider:       db.UserProviderGithub,
			ProviderUserID: "github-1",
			Email:          "changed@example.com",
			Name:           "Ops",
		})
		require.NoError(t, err)

		assert.Equal(t, db.Id("u-1"), result.User.Id)
		require.Len(t, sessions.created, 1)
	})

	t.Run("oauth login creates a user for a new email", func(t *testing.T) {
		auth, users, sessions := newAuth(db.UserRoleRegular, true, "0")

		result, err := auth.LoginWithOAuth(context.Background(), &OAuthIdentity{
			Provider:       db.UserProviderGoogle,
			ProviderUserID: "google-1",
			Email:          "new@example.com",
			Name:           "",
		})
		require.NoError(t, err)

		assert.Equal(t, "new", result.User.Name)
		assert.Equal(t, "new@example.com", result.User.Email)
		assert.Equal(t, db.UserProviderGoogle, result.User.Provider)
		assert.Equal(t, db.UserRoleRegular, result.User.Role)
		assert.True(t, result.User.IsActive)
		assert.True(t, result.User.IsEmailVerified)
		require.Len(t, users.created, 1)
		require.Len(t, sessions.created, 1)
		assert.Equal(t, users.created[0].Id, sessions.created[0].UserId)
	})

	t.Run("oauth login rejects an inactive user", func(t *testing.T) {
		auth, _, _ := newAuth(db.UserRoleRegular, false, "0")

		_, err := auth.LoginWithOAuth(context.Background(), &OAuthIdentity{
			Provider:       db.UserProviderGithub,
			ProviderUserID: "github-1",
			Email:          "ops@example.com",
		})
		assert.ErrorIs(t, err, ErrAccountNotActive)
	})

	t.Run("maintenance mode blocks non-admins only", func(t *testing.T) {
		identity := &OAuthIdentity{
			Provider:       db.UserProviderGithub,
			ProviderUserID: "github-1",
			Email:          "ops@example.com",
		}

		auth, _, _ := newAuth(db.UserRoleRegular, true, "1")

		_, err := auth.LoginWithOAuth(context.Background(), identity)
		assert.ErrorIs(t, err, ErrMaintenanceModeEnabled)

		auth, _, _ = newAuth(db.UserRoleAdmin, true, "1")

		_, err = auth.LoginWithOAuth(context.Background(), identity)
		assert.NoError(t, err)
	})

	t.Run("logout by token revokes the session owner and ignores unknown tokens", func(t *testing.T) {
		auth, _, sessions := newAuth(db.UserRoleRegular, true, "0")

		result, err := auth.LoginWithOAuth(context.Background(), &OAuthIdentity{
			Provider:       db.UserProviderGithub,
			ProviderUserID: "github-1",
			Email:          "ops@example.com",
		})
		require.NoError(t, err)

		require.NoError(t, auth.LogoutByToken("unknown"))
		require.NoError(t, auth.LogoutByToken(""))
		assert.Empty(t, sessions.deletedByUser)

		require.NoError(t, auth.LogoutByToken(result.Session.Token))
		assert.Equal(t, []db.Id{"u-1"}, sessions.deletedByUser)
	})

	t.Run("logout revokes the user sessions", func(t *testing.T) {
		auth, _, sessions := newAuth(db.UserRoleRegular, true, "0")

		require.NoError(t, auth.Logout("u-1"))
		assert.Equal(t, []db.Id{"u-1"}, sessions.deletedByUser)
	})
}
