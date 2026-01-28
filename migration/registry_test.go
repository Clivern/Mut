// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitMigrationsChained(t *testing.T) {
	migrations := GetAll()

	require.Len(t, migrations, 20)

	for i, migration := range migrations {
		assert.NotEmpty(t, migration.Version)
		assert.NotEmpty(t, migration.Description)
		assert.NotNil(t, migration.Up)
		assert.NotNil(t, migration.Down)

		if i > 0 {
			assert.Less(t, migrations[i-1].Version, migration.Version)
		}
	}
}
