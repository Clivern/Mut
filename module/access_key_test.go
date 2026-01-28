// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"strings"
	"testing"

	"github.com/clivern/mut/db"
	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAccessKeys struct {
	db.AccessKeyRepository
	keys map[db.Id]*db.AccessKey
}

func (f *fakeAccessKeys) Create(accessKey *db.AccessKey) error {
	accessKey.Id = db.Id("k-new")
	f.keys[accessKey.Id] = accessKey

	return nil
}

func (f *fakeAccessKeys) GetById(id db.Id) (*db.AccessKey, error) {
	return f.keys[id], nil
}

func (f *fakeAccessKeys) GetByKeyHash(keyHash string) (*db.AccessKey, error) {
	for _, k := range f.keys {
		if k.KeyHash == keyHash {
			return k, nil
		}
	}

	return nil, nil
}

func (f *fakeAccessKeys) ListByWorkspaceId(_ db.Id, _, _ int) ([]*db.AccessKey, error) {
	var list []*db.AccessKey
	for _, k := range f.keys {
		list = append(list, k)
	}

	return list, nil
}

func (f *fakeAccessKeys) CountByWorkspaceId(_ db.Id) (int64, error) {
	return int64(len(f.keys)), nil
}

func (f *fakeAccessKeys) Delete(id db.Id) error {
	delete(f.keys, id)

	return nil
}

func TestUnitAccessKey(t *testing.T) {
	owner := &db.User{Id: db.Id("u-owner")}

	newModule := func() (*AccessKey, *fakeAccessKeys) {
		keys := &fakeAccessKeys{keys: map[db.Id]*db.AccessKey{}}
		workspaces := &fakeWorkspaces{
			workspaces: map[db.Id]*db.Workspace{"w-1": {Id: "w-1", Name: "Platform"}},
			roles: map[db.Id]string{
				"w-1": db.WorkspaceRoleOwner,
			},
		}

		return NewAccessKey(keys, workspaces), keys
	}

	t.Run("create returns the secret once and stores only its hash", func(t *testing.T) {
		module, keys := newModule()

		created, err := module.CreateAccessKey("w-1", owner, &CreateAccessKeyRequest{
			Name:        "ci",
			Permissions: []string{AccessKeyPermissionWorkspaceRead},
		})

		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(created.Key, AccessKeyPrefix))
		assert.Equal(t, AccessKeyPrefix+"****"+created.Key[len(created.Key)-4:], created.MaskedKey)

		stored := keys.keys["k-new"]
		assert.NotEqual(t, created.Key, stored.KeyHash)
		assert.Equal(t, hashAccessKey(created.Key), stored.KeyHash)
	})

	t.Run("authenticate finds the key by its secret", func(t *testing.T) {
		module, _ := newModule()

		created, err := module.CreateAccessKey("w-1", owner, &CreateAccessKeyRequest{
			Name:        "ci",
			Permissions: []string{AccessKeyPermissionWorkspaceRead},
		})
		require.NoError(t, err)

		key, err := module.AuthenticateAccessKey(created.Key)

		require.NoError(t, err)
		assert.True(t, lo.Contains(strings.Split(key.Permissions, ","), AccessKeyPermissionWorkspaceRead))
		assert.False(t, lo.Contains(strings.Split(key.Permissions, ","), AccessKeyPermissionMembersRead))
	})

	t.Run("authenticate rejects an unknown secret", func(t *testing.T) {
		module, _ := newModule()

		_, err := module.AuthenticateAccessKey("mut_unknown")

		assert.ErrorIs(t, err, ErrAccessKeyInvalid)
	})

	t.Run("delete revokes the key", func(t *testing.T) {
		module, keys := newModule()

		created, err := module.CreateAccessKey("w-1", owner, &CreateAccessKeyRequest{
			Name:        "ci",
			Permissions: []string{AccessKeyPermissionWorkspaceRead},
		})
		require.NoError(t, err)

		require.NoError(t, module.DeleteAccessKey("w-1", owner, created.Id))
		assert.Empty(t, keys.keys)
	})

	t.Run("get rejects a key from another workspace", func(t *testing.T) {
		module, _ := newModule()

		created, err := module.CreateAccessKey("w-1", owner, &CreateAccessKeyRequest{
			Name:        "ci",
			Permissions: []string{AccessKeyPermissionWorkspaceRead},
		})
		require.NoError(t, err)

		_, err = module.GetAccessKey("w-other", owner, created.Id)

		assert.ErrorIs(t, err, ErrAccessKeyNotFound)
	})
}
