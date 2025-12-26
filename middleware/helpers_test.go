// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitShouldSkipAuth(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{path: "/api/v1/public/_health", want: true},
		{path: "/api/v1/public/action/login", want: true},
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
