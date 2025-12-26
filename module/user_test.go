// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitUser(t *testing.T) {
	newUser := func() (*User, *fakeUsers, *fakeSessions) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Name: "Admin", Email: "admin@example.com", Password: "hash", Role: db.UserRoleAdmin, IsActive: true},
		}}
		sessions := &fakeSessions{}

		return NewUser(users, sessions), users, sessions
	}

	t.Run("create hashes the password", func(t *testing.T) {
		module, users, _ := newUser()

		user, err := module.CreateUser(&CreateUserRequest{
			Name:     "Ops",
			Email:    "ops@example.com",
			Password: "Secret#123",
			Role:     db.UserRoleRegular,
			IsActive: true,
		})
		require.NoError(t, err)

		require.Len(t, users.created, 1)
		assert.Equal(t, "ops@example.com", user.Email)
		assert.True(t, util.ComparePassword(user.Password, "Secret#123"))
	})

	t.Run("create rejects a duplicate email", func(t *testing.T) {
		module, _, _ := newUser()

		_, err := module.CreateUser(&CreateUserRequest{Name: "Dup", Email: "admin@example.com", Password: "Secret#123", Role: db.UserRoleRegular})
		assert.ErrorIs(t, err, ErrUserEmailAlreadyExists)
	})

	t.Run("get returns not found for an unknown user", func(t *testing.T) {
		module, _, _ := newUser()

		_, err := module.GetUser("missing")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("update keeps the password when empty and revokes sessions on deactivate", func(t *testing.T) {
		module, users, sessions := newUser()

		user, err := module.UpdateUser("u-1", &UpdateUserRequest{
			Name:     "Renamed",
			Email:    "admin@example.com",
			Role:     db.UserRoleReadOnly,
			IsActive: false,
		})
		require.NoError(t, err)

		assert.Equal(t, "Renamed", user.Name)
		assert.Equal(t, "hash", user.Password)
		assert.Len(t, users.updated, 1)
		assert.Equal(t, []db.Id{"u-1"}, sessions.deletedByUser)
	})

	t.Run("delete refuses to remove the current user", func(t *testing.T) {
		module, users, _ := newUser()
		current := users.byId["u-1"]

		assert.ErrorIs(t, module.DeleteUser("u-1", current), ErrCannotDeleteSelf)
		assert.ErrorIs(t, module.DeleteUser("missing", current), ErrUserNotFound)
	})

	t.Run("list returns users and total", func(t *testing.T) {
		module, _, _ := newUser()

		result, err := module.ListUsers(10, 0)
		require.NoError(t, err)

		assert.Len(t, result.Users, 1)
		assert.Equal(t, int64(1), result.Total)
	})
}
