// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package llm

import (
	"errors"
)

var (
	ErrEmbedNotConfigured = errors.New("ai embeddings are not configured")
	ErrLLMNotConfigured   = errors.New("ai llm is not configured")
	ErrNoEmbeddings       = errors.New("ai returned no embeddings")
	ErrNoCompletion       = errors.New("ai returned no completion")
)
