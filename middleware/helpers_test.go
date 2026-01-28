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

func TestUnitShouldSkipAuth(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{path: "/api/v1/public/_health", want: true},
		{path: "/assets/app.js", want: true},
		{path: "/login", want: true},
		{path: "/api/v1/users", want: false},
		{path: "/api/v1/me", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.want, ShouldSkipAuth(tc.path))
		})
	}
}

func TestUnitIsAssetPath(t *testing.T) {
	assert.True(t, IsAssetPath("/assets/app.js"))
	assert.False(t, IsAssetPath("/api/v1/me"))
	assert.False(t, IsAssetPath("/login"))
}

func TestUnitRequestContext(t *testing.T) {
	user := &db.User{Id: "u-1"}
	key := &db.AccessKey{Id: "k-1"}
	workspace := &db.Workspace{Id: "w-1"}

	ctx := WithUserContext(context.Background(), user)
	got, ok := GetUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, user, got)

	_, ok = GetUserFromContext(context.Background())
	assert.False(t, ok)

	ctx = WithAccessKeyContext(context.Background(), key)
	gotKey, ok := GetAccessKeyFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, key, gotKey)

	_, ok = GetAccessKeyFromContext(context.Background())
	assert.False(t, ok)

	ctx = context.WithValue(context.Background(), ContextKeyWorkspace, workspace)
	gotWorkspace, ok := GetWorkspaceFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, workspace, gotWorkspace)

	_, ok = GetWorkspaceFromContext(context.Background())
	assert.False(t, ok)
}

func TestUnitGetClientIP(t *testing.T) {
	newRequest := func(remoteAddr string, headers map[string]string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remoteAddr
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		return req
	}

	assert.Equal(
		t,
		"1.2.3.4",
		GetClientIP(newRequest("9.9.9.9:1234", map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8"})),
	)
	assert.Equal(
		t,
		"1.2.3.4",
		GetClientIP(newRequest("9.9.9.9:1234", map[string]string{"X-Real-Ip": "1.2.3.4"})),
	)
	assert.Equal(t, "9.9.9.9", GetClientIP(newRequest("9.9.9.9:1234", nil)))
}

func TestUnitParseFeatures(t *testing.T) {
	assert.Empty(t, ParseFeatures(""))

	features := ParseFeatures("a, b ,a")
	assert.True(t, features["a"])
	assert.True(t, features["b"])
}
