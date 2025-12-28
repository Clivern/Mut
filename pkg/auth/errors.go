// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package auth

import "errors"

var (
	ErrType        = errors.New("auth: unknown auth type")
	ErrConfig      = errors.New("auth: incomplete scheme config")
	ErrCredentials = errors.New("auth: invalid credentials")
)
