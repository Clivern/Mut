// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitClusterStateShape(t *testing.T) {
	state := NewClusterState("prod", "v1.33.0")
	state.Pods = append(state.Pods, PodState{
		ObjectMeta: ObjectMeta{Name: "web-0", Namespace: "default"},
		Phase:      "Running",
		OwnerKind:  "ReplicaSet",
		OwnerName:  "web-abc",
		Restarts:   2,
		Requests:   map[string]string{"cpu": "500m"},
		Usage:      &ResourceUsage{CPU: "120m", Memory: "256Mi"},
	})
	state.Nodes = append(state.Nodes, NodeState{
		ObjectMeta: ObjectMeta{Name: "node-1"},
		Taints:     []TaintState{{Key: "gpu", Value: "true", Effect: "NoSchedule"}},
	})
	state.Endpoints = append(state.Endpoints, EndpointState{
		ObjectMeta: ObjectMeta{Name: "web-abc", Namespace: "default"},
		Service:    "web",
		Ready:      2,
	})
	state.Secrets = append(state.Secrets, SecretState{
		ObjectMeta: ObjectMeta{Name: "db-credentials", Namespace: "default"},
		Type:       "Opaque",
		Keys:       []string{"password"},
	})
	state.Events = append(state.Events, EventState{
		Namespace: "default",
		Kind:      "Pod",
		Name:      "web-0",
		Type:      "Warning",
		Reason:    "Unhealthy",
		Count:     3,
	})

	raw, err := json.Marshal(state)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, "prod", decoded["cluster"].(map[string]any)["name"])
	assert.Len(t, decoded["pods"], 1)
	assert.Len(t, decoded["events"], 1)

	pod := decoded["pods"].([]any)[0].(map[string]any)
	assert.Equal(t, "ReplicaSet", pod["ownerKind"])
	assert.Equal(t, "web-abc", pod["ownerName"])

	node := decoded["nodes"].([]any)[0].(map[string]any)
	assert.Len(t, node["taints"], 1)

	endpoint := decoded["endpoints"].([]any)[0].(map[string]any)
	assert.Equal(t, "web", endpoint["service"])
	assert.Equal(t, float64(2), endpoint["ready"])
	assert.Equal(t, map[string]any{"cpu": "120m", "memory": "256Mi"}, pod["usage"])
	assert.NotContains(t, pod, "ObjectMeta")

	secret := decoded["secrets"].([]any)[0].(map[string]any)
	assert.Equal(t, []any{"password"}, secret["keys"])
	assert.NotContains(t, secret, "data")
	assert.NotContains(t, secret, "stringData")
}
