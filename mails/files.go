// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mails

import _ "embed"

//go:embed invite.html
var Invite string

//go:embed welcome.html
var Welcome string
