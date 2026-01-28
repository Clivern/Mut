// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (f *fakeSessions) Delete(id db.Id) error {
	for i, session := range f.created {
		if session.Id == id {
			f.created = append(f.created[:i], f.created[i+1:]...)
		}
	}

	return nil
}

func (f *fakeSessions) GetByUserId(userId db.Id) ([]*db.Session, error) {
	var list []*db.Session
	for _, session := range f.created {
		if session.UserId == userId {
			list = append(list, session)
		}
	}

	return list, nil
}

func TestUnitSessionManager(t *testing.T) {
	newManager := func() (*SessionManager, *fakeSessions, *fakeUsers) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Email: "ops@example.com", IsActive: true},
		}}
		sessions := &fakeSessions{}

		return NewSessionManager(sessions, users), sessions, users
	}

	t.Run("create stores a session for a user", func(t *testing.T) {
		manager, sessions, _ := newManager()
		ctx := WithAppCtx(context.Background(), NewAppCtx("req-1", "curl", "10.0.0.1", nil))

		session, err := manager.CreateSession(ctx, "u-1", time.Hour)
		require.NoError(t, err)

		assert.Equal(t, db.Id("u-1"), session.UserId)
		assert.NotEmpty(t, session.Token)
		assert.Equal(t, "10.0.0.1", *session.IPAddress)
		require.Len(t, sessions.created, 1)
	})

	t.Run("create rejects an unknown user", func(t *testing.T) {
		manager, sessions, _ := newManager()

		_, err := manager.CreateSession(context.Background(), "u-9", time.Hour)
		require.Error(t, err)
		assert.Empty(t, sessions.created)
	})

	t.Run("validate returns the user for a live session", func(t *testing.T) {
		manager, sessions, _ := newManager()
		sessions.created = []*db.Session{{
			Id:        "s-1",
			Token:     "token-1",
			UserId:    "u-1",
			ExpiresAt: time.Now().Add(time.Hour),
		}}

		user, session, err := manager.ValidateSession("token-1")
		require.NoError(t, err)

		assert.Equal(t, db.Id("u-1"), user.Id)
		assert.Equal(t, db.Id("s-1"), session.Id)
	})

	t.Run("validate rejects an unknown token", func(t *testing.T) {
		manager, _, _ := newManager()

		_, _, err := manager.ValidateSession("nope")
		require.Error(t, err)
	})

	t.Run("validate drops expired sessions", func(t *testing.T) {
		manager, sessions, _ := newManager()
		sessions.created = []*db.Session{{
			Id:        "s-1",
			Token:     "token-1",
			UserId:    "u-1",
			ExpiresAt: time.Now().Add(-time.Hour),
		}}

		_, _, err := manager.ValidateSession("token-1")
		require.Error(t, err)
		assert.Empty(t, sessions.created)
	})

	t.Run("validate rejects an inactive user", func(t *testing.T) {
		manager, sessions, users := newManager()
		users.byId["u-1"].IsActive = false
		sessions.created = []*db.Session{{
			Id:        "s-1",
			Token:     "token-1",
			UserId:    "u-1",
			ExpiresAt: time.Now().Add(time.Hour),
		}}

		_, _, err := manager.ValidateSession("token-1")
		require.Error(t, err)
	})

	t.Run("revoke removes the user sessions", func(t *testing.T) {
		manager, sessions, _ := newManager()
		sessions.created = []*db.Session{{Id: "s-1", UserId: "u-1"}}

		require.NoError(t, manager.RevokeUserSessions("u-1"))
		assert.Equal(t, []db.Id{"u-1"}, sessions.deletedByUser)
	})

	t.Run("list returns only live sessions", func(t *testing.T) {
		manager, sessions, _ := newManager()
		sessions.created = []*db.Session{
			{Id: "s-1", UserId: "u-1", ExpiresAt: time.Now().Add(time.Hour)},
			{Id: "s-2", UserId: "u-1", ExpiresAt: time.Now().Add(-time.Hour)},
		}

		list, err := manager.GetUserSessions("u-1")
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, db.Id("s-1"), list[0].Id)
	})

	t.Run("cleanup reports expired sessions", func(t *testing.T) {
		manager, _, _ := newManager()

		count, err := manager.CleanupExpiredSessions()
		require.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("validate reports user lookup errors", func(t *testing.T) {
		users := &fakeUsers{byId: map[db.Id]*db.User{
			"u-1": {Id: "u-1", Email: "ops@example.com", IsActive: true},
		}, err: errors.New("boom")}
		sessions := &fakeSessions{created: []*db.Session{{
			Id:        "s-1",
			Token:     "token-1",
			UserId:    "u-1",
			ExpiresAt: time.Now().Add(time.Hour),
		}}}
		manager := NewSessionManager(sessions, users)

		_, _, err := manager.ValidateSession("token-1")
		require.Error(t, err)
	})
}
