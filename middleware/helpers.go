// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"strings"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/module"
)

const (
	// ContextKeyUser is the key for storing user in context
	ContextKeyUser contextKey = "user"
)

// IsAssetPath returns true if path is a static asset
func IsAssetPath(path string) bool {
	return strings.HasPrefix(path, "/assets/")
}

// ShouldSkipAuth determines if authentication should be skipped for a given path
func ShouldSkipAuth(path string) bool {
	return strings.HasPrefix(path, "/api/v1/public/") ||
		!strings.HasPrefix(path, "/api/v1/")
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	return module.UserFrom(ctx)
}

// WithUserContext sets the user in the request context
func WithUserContext(ctx context.Context, user *db.User) context.Context {
	ctx = module.WithPrincipal(ctx, &module.Principal{User: user})

	return context.WithValue(ctx, ContextKeyUser, user)
}
