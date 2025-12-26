// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"testing"

	"github.com/clivern/mut/conf"
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAuth(t *testing.T) {
	hash, err := util.HashPassword("Secret#123")
	require.NoError(t, err)

	newAuth := func(role string, active bool, maintenance string) (*Auth, *fakeUsers, *fakeSessions) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Email: "ops@example.com", Password: hash, Role: role, IsActive: active},
		}}
		sessions := &fakeSessions{}
		configs := &fakeConfigs{values: map[string]string{"maintenance_mode": maintenance}}

		return NewAuth(users, sessions, configs), users, sessions
	}

	t.Run("login creates a session for valid credentials", func(t *testing.T) {
		auth, users, sessions := newAuth(db.UserRoleRegular, true, "0")
		ctx := WithAppCtx(context.Background(), NewAppCtx("req-1", "curl", "10.0.0.1", nil))

		result, err := auth.Login(ctx, &LoginRequest{Email: "ops@example.com", Password: "Secret#123"})
		require.NoError(t, err)

		assert.Equal(t, db.Id("u-1"), result.User.Id)
		require.Len(t, sessions.created, 1)
		assert.Equal(t, "10.0.0.1", *sessions.created[0].IPAddress)
		assert.Equal(t, "curl", *sessions.created[0].UserAgent)
		assert.Equal(t, []db.Id{"u-1"}, users.lastLogin)
		assert.Equal(t, 0, result.CookieOptions.MaxAge)
	})

	t.Run("login with remember me keeps the cookie", func(t *testing.T) {
		auth, _, _ := newAuth(db.UserRoleRegular, true, "0")

		result, err := auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "Secret#123", RememberMe: true})
		require.NoError(t, err)

		assert.Equal(t, int(conf.RememberMeSessionDuration.Seconds()), result.CookieOptions.MaxAge)
	})

	t.Run("login rejects a wrong password or unknown email", func(t *testing.T) {
		auth, _, sessions := newAuth(db.UserRoleRegular, true, "0")

		_, err := auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "wrong"})
		assert.ErrorIs(t, err, ErrInvalidCredentials)

		_, err = auth.Login(context.Background(), &LoginRequest{Email: "nobody@example.com", Password: "Secret#123"})
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, sessions.created)
	})

	t.Run("login rejects an inactive user", func(t *testing.T) {
		auth, _, _ := newAuth(db.UserRoleRegular, false, "0")

		_, err := auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "Secret#123"})
		assert.ErrorIs(t, err, ErrAccountNotActive)
	})

	t.Run("maintenance mode blocks non-admins only", func(t *testing.T) {
		auth, _, _ := newAuth(db.UserRoleRegular, true, "1")

		_, err := auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "Secret#123"})
		assert.ErrorIs(t, err, ErrMaintenanceModeEnabled)

		auth, _, _ = newAuth(db.UserRoleAdmin, true, "1")

		_, err = auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "Secret#123"})
		assert.NoError(t, err)
	})

	t.Run("logout by token revokes the session owner and ignores unknown tokens", func(t *testing.T) {
		auth, _, sessions := newAuth(db.UserRoleRegular, true, "0")

		result, err := auth.Login(context.Background(), &LoginRequest{Email: "ops@example.com", Password: "Secret#123"})
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
