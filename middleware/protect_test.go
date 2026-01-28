// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/mut/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitProtectRoles(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	call := func(cfg Config, ctx context.Context) int {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/w-1", nil)
		if ctx != nil {
			req = req.WithContext(ctx)
		}

		rec := httptest.NewRecorder()
		Protect(cfg)(next).ServeHTTP(rec, req)

		return rec.Code
	}

	roles := []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly}

	t.Run("rejects request without principal", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, call(Config{Roles: roles}, nil))
	})

	t.Run("rejects inactive user", func(t *testing.T) {
		ctx := WithUserContext(context.Background(), &db.User{Id: "u-1", Role: db.UserRoleAdmin})

		assert.Equal(t, http.StatusForbidden, call(Config{Roles: roles}, ctx))
	})

	t.Run("rejects user with unlisted role", func(t *testing.T) {
		ctx := WithUserContext(context.Background(), &db.User{Id: "u-1", Role: "ghost", IsActive: true})

		assert.Equal(t, http.StatusForbidden, call(Config{Roles: roles}, ctx))
	})

	t.Run("lets user with listed role through", func(t *testing.T) {
		ctx := WithUserContext(context.Background(), &db.User{Id: "u-1", Role: db.UserRoleRegular, IsActive: true})

		assert.Equal(t, http.StatusNoContent, call(Config{Roles: roles}, ctx))
	})

	t.Run("rejects access key when user required", func(t *testing.T) {
		ctx := WithAccessKeyContext(context.Background(), &db.AccessKey{Id: "k-1", WorkspaceId: db.Id("w-1")})

		assert.Equal(t, http.StatusUnauthorized, call(Config{Roles: roles}, ctx))
	})
}
