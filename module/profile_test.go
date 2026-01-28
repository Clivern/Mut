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

func TestUnitProfile(t *testing.T) {
	newProfile := func() (*Profile, *fakeUsers) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Name: "Ops", Email: "ops@example.com", Language: "en", Theme: "default"},
		}}

		return NewProfile(users), users
	}

	t.Run("get returns the profile", func(t *testing.T) {
		profile, _ := newProfile()

		user, err := profile.GetProfile("u-1")
		require.NoError(t, err)
		assert.Equal(t, "ops@example.com", user.Email)
	})

	t.Run("get rejects an unknown user", func(t *testing.T) {
		profile, _ := newProfile()

		_, err := profile.GetProfile("u-9")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("get reports repository errors", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{err: errors.New("boom")})

		_, err := profile.GetProfile("u-1")
		assert.ErrorIs(t, err, ErrFailedGetProfile)
	})

	t.Run("update changes the profile", func(t *testing.T) {
		profile, users := newProfile()

		updated, err := profile.UpdateProfile("u-1", &UpdateProfileRequest{
			Name:     "Bo",
			Language: "fr",
			Theme:    "dark",
		})
		require.NoError(t, err)

		assert.Equal(t, "Bo", updated.Name)
		assert.Equal(t, "fr", users.byId["u-1"].Language)
		assert.Equal(t, "dark", users.byId["u-1"].Theme)
	})

	t.Run("update keeps the name when empty", func(t *testing.T) {
		profile, _ := newProfile()

		updated, err := profile.UpdateProfile("u-1", &UpdateProfileRequest{
			Language: "en",
			Theme:    "default",
		})
		require.NoError(t, err)

		assert.Equal(t, "Ops", updated.Name)
	})

	t.Run("update rejects an unknown user", func(t *testing.T) {
		profile, _ := newProfile()

		_, err := profile.UpdateProfile("u-9", &UpdateProfileRequest{Language: "en", Theme: "default"})
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("update reports repository errors", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{
			byId: map[db.Id]*db.User{"u-1": {Id: "u-1"}},
			err:  errors.New("boom"),
		})

		_, err := profile.UpdateProfile("u-1", &UpdateProfileRequest{Language: "en", Theme: "default"})
		assert.ErrorIs(t, err, ErrFailedUpdateProfile)
	})
}
