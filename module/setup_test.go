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

func TestUnitSetup(t *testing.T) {
	req := &SetupRequest{
		PlatformEmail: "ops@example.com",
	}

	t.Run("is installed is false before setup and true after", func(t *testing.T) {
		configs := &fakeConfigs{values: map[string]string{}}
		setup := NewSetup(configs, &fakeUsers{})

		assert.False(t, setup.IsInstalled())

		configs.values["is_installed"] = "1"
		assert.True(t, setup.IsInstalled())
	})

	t.Run("is installed is false when the config read fails", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{err: errors.New("boom")}, &fakeUsers{})

		assert.False(t, setup.IsInstalled())
	})

	t.Run("install rejects an already installed platform", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{values: map[string]string{"is_installed": "1"}}, &fakeUsers{})

		assert.ErrorIs(t, setup.Install(req), ErrPlatformAlreadyInstalled)
	})

	t.Run("install creates the admin user and default config", func(t *testing.T) {
		configs := &fakeConfigs{values: map[string]string{}}
		users := &fakeUsers{}
		setup := NewSetup(configs, users)

		require.NoError(t, setup.Install(req))

		require.Len(t, users.created, 1)
		admin := users.created[0]
		assert.Equal(t, db.UserRoleAdmin, admin.Role)
		assert.Equal(t, "ops@example.com", admin.Email)
		assert.Empty(t, admin.Password)

		assert.Equal(t, "1", configs.values["is_installed"])
		assert.Equal(t, "ops@example.com", configs.values["platform_email"])
		assert.Equal(t, "0", configs.values["maintenance_mode"])
	})

	t.Run("install fails when the admin user cannot be created", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{values: map[string]string{}}, &fakeUsers{err: errors.New("boom")})

		assert.ErrorIs(t, setup.Install(req), ErrFailedCompleteSetup)
	})
}
