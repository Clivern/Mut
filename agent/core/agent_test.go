// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/mut/agent/pkg/k8s"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAgentPushSendsStateWithAccessKey(t *testing.T) {
	var gotMethod, gotPath, gotKey string
	var gotState k8s.ClusterState

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-Access-Key")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotState))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	agent := &Agent{
		config: Config{
			Server:      server.URL,
			WorkspaceId: "w-1",
			ClusterId:   "c-1",
			AccessKey:   "mut_secret",
		},
		client: server.Client(),
	}

	state := k8s.NewClusterState("prod", "v1.31.0")
	err := agent.Push(context.Background(), state)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/api/v1/workspaces/w-1/clusters/c-1/state", gotPath)
	assert.Equal(t, "mut_secret", gotKey)
	assert.Equal(t, "prod", gotState.Cluster.Name)
}

func TestUnitAgentPushReturnsErrorOnRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errorMessage":"invalid access key"}`))
	}))
	defer server.Close()

	agent := &Agent{
		config: Config{Server: server.URL, WorkspaceId: "w-1", ClusterId: "c-1", AccessKey: "bad"},
		client: server.Client(),
	}

	err := agent.Push(context.Background(), k8s.NewClusterState("prod", ""))
	assert.ErrorContains(t, err, "status 401")
}
