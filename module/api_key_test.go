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

type fakeAPIKeyStore struct {
	db.APIKeyRepository
	keys    map[db.Id]*db.APIKey
	err     error
	deleted []db.Id
}

func (f *fakeAPIKeyStore) Create(key *db.APIKey) error {
	if f.err != nil {
		return f.err
	}

	key.Id = db.Id("k-new")
	f.keys[key.Id] = key

	return nil
}

func (f *fakeAPIKeyStore) GetById(id db.Id) (*db.APIKey, error) {
	return f.keys[id], f.err
}

func (f *fakeAPIKeyStore) ListByUserId(_ db.Id, _, _ int) ([]*db.APIKey, error) {
	var list []*db.APIKey
	for _, key := range f.keys {
		list = append(list, key)
	}

	return list, f.err
}

func (f *fakeAPIKeyStore) CountByUserId(_ db.Id) (int64, error) {
	return int64(len(f.keys)), f.err
}

func (f *fakeAPIKeyStore) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeAPIKeyStore) DeleteExpired() (int64, error) {
	return 0, f.err
}

func TestUnitAPIKey(t *testing.T) {
	user := &db.User{Id: db.Id("u-1")}
	stored := &db.APIKey{Id: db.Id("k-1"), UserId: db.Id("u-1"), Name: "CI"}

	newStore := func() *fakeAPIKeyStore {
		return &fakeAPIKeyStore{keys: map[db.Id]*db.APIKey{"k-1": stored}}
	}

	t.Run("create returns the raw key once", func(t *testing.T) {
		apiKeys := NewAPIKey(newStore())

		result, err := apiKeys.CreateAPIKey(&CreateAPIKeyRequest{Name: "CI"}, user)
		require.NoError(t, err)

		assert.NotEmpty(t, result.Key)
		assert.Equal(t, "CI", result.Name)
	})

	t.Run("create rejects a bad expiry", func(t *testing.T) {
		apiKeys := NewAPIKey(newStore())

		_, err := apiKeys.CreateAPIKey(&CreateAPIKeyRequest{Name: "CI", ExpiresAt: "nope"}, user)
		assert.ErrorIs(t, err, ErrInvalidExpiresAt)
	})

	t.Run("create reports repository errors", func(t *testing.T) {
		store := newStore()
		store.err = errors.New("boom")

		_, err := NewAPIKey(store).CreateAPIKey(&CreateAPIKeyRequest{Name: "CI"}, user)
		require.Error(t, err)
	})

	t.Run("list returns keys with a total", func(t *testing.T) {
		result, err := NewAPIKey(newStore()).ListAPIKeys(user, 10, 0)
		require.NoError(t, err)

		require.Len(t, result.APIKeys, 1)
		assert.Equal(t, int64(1), result.Total)
	})

	t.Run("get returns one key", func(t *testing.T) {
		result, err := NewAPIKey(newStore()).GetAPIKey("k-1", user)
		require.NoError(t, err)

		assert.Equal(t, db.Id("k-1"), result.Id)
		assert.Empty(t, result.Key)
	})

	t.Run("get rejects a key owned by someone else", func(t *testing.T) {
		_, err := NewAPIKey(newStore()).GetAPIKey("k-1", &db.User{Id: "u-9"})
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("get rejects a missing key", func(t *testing.T) {
		_, err := NewAPIKey(newStore()).GetAPIKey("k-9", user)
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("delete revokes the key", func(t *testing.T) {
		store := newStore()

		require.NoError(t, NewAPIKey(store).DeleteAPIKey("k-1", user))
		assert.Equal(t, []db.Id{"k-1"}, store.deleted)
	})

	t.Run("delete rejects a missing key", func(t *testing.T) {
		err := NewAPIKey(newStore()).DeleteAPIKey("k-9", user)
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("delete reports repository errors", func(t *testing.T) {
		store := newStore()
		store.err = errors.New("boom")

		assert.Error(t, NewAPIKey(store).DeleteAPIKey("k-1", user))
	})
}
