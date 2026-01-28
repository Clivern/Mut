// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/clivern/mut/db"
)

// ErrNoCluster is returned by a tool when no cluster is open on the page.
var ErrNoCluster = errors.New("no cluster is open")

// ChatTool is one read-only action the assistant can run to look up cluster data.
type ChatTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Run         ChatToolRun `json:"-"`
}

// ChatToolRun looks up data for a cluster and returns it as plain JSON-friendly values.
type ChatToolRun func(ctx context.Context, workspaceId, clusterId db.Id, args map[string]string) (any, error)

// chatResourceGroups lists the ClusterState groups a listResources call can read.
const chatResourceGroups = "nodes, " +
	"namespaces, " +
	"pods, " +
	"deployments, " +
	"statefulSets, " +
	"daemonSets, " +
	"replicaSets, " +
	"jobs, " +
	"cronJobs, " +
	"services, " +
	"endpoints, " +
	"ingresses, " +
	"ingressClasses, " +
	"configMaps, " +
	"secrets, " +
	"volumes, " +
	"claims, " +
	"storageClasses, " +
	"autoscalers, " +
	"disruptions, " +
	"quotas, " +
	"limits, " +
	"priorities, " +
	"policies, " +
	"accounts, " +
	"roles, " +
	"clusterRoles, " +
	"bindings, " +
	"clusterBindings, " +
	"runtimeClasses, " +
	"events, " +
	"errors"

// NewChatTools returns the cluster tools the assistant can run.
func NewChatTools(cluster *Cluster) []*ChatTool {
	return []*ChatTool{
		{
			Name: "clusterSummary",
			Description: "Counts for the open cluster: nodes ready, pods by phase, replicas ready, " +
				"warning events and collection errors. No arguments.",
			Run: SummaryTool(cluster),
		},
		{
			Name: "listResources",
			Description: fmt.Sprintf(
				"Lists one kind of resource in the open cluster. Arguments: group (one of: %s) "+
					"and namespace (optional, limits namespaced resources to one namespace).",
				chatResourceGroups,
			),
			Run: ListResourcesTool(cluster),
		},
	}
}

func SummaryTool(cluster *Cluster) ChatToolRun {
	return func(ctx context.Context, workspaceId, clusterId db.Id, _ map[string]string) (any, error) {
		if clusterId == "" {
			return nil, ErrNoCluster
		}

		state, err := cluster.GetState(ctx, workspaceId, clusterId)
		if err != nil {
			return nil, err
		}

		pods := map[string]int{}
		for _, pod := range state.Pods {
			pods[pod.Phase]++
		}

		nodesReady := 0
		for _, node := range state.Nodes {
			ready := false
			for _, condition := range node.Conditions {
				if condition.Type == "Ready" {
					ready = condition.Status == "True"
					break
				}
			}

			if ready {
				nodesReady++
			}
		}

		replicasReady, replicasDesired := 0, 0
		for _, deployment := range state.Deployments {
			replicasReady += int(deployment.ReplicasReady)
			replicasDesired += int(deployment.ReplicasDesired)
		}

		warnings := 0
		for _, event := range state.Events {
			if event.Type == "Warning" {
				warnings++
			}
		}

		collectionErrors := []string{}
		for key := range state.Errors {
			collectionErrors = append(collectionErrors, key)
		}

		return map[string]any{
			"nodesReady":       nodesReady,
			"nodesTotal":       len(state.Nodes),
			"pods":             pods,
			"replicasReady":    replicasReady,
			"replicasDesired":  replicasDesired,
			"warningEvents":    warnings,
			"collectionErrors": collectionErrors,
		}, nil
	}
}

func ListResourcesTool(cluster *Cluster) ChatToolRun {
	return func(ctx context.Context, workspaceId, clusterId db.Id, args map[string]string) (any, error) {
		if clusterId == "" {
			return nil, ErrNoCluster
		}

		state, err := cluster.GetState(ctx, workspaceId, clusterId)
		if err != nil {
			return nil, err
		}

		raw, err := json.Marshal(state)
		if err != nil {
			return nil, err
		}

		var groups map[string]json.RawMessage
		err = json.Unmarshal(raw, &groups)
		if err != nil {
			return nil, err
		}

		group := args["group"]
		data, ok := groups[group]
		if !ok {
			return nil, fmt.Errorf("unknown group %q", group)
		}

		namespace := args["namespace"]
		if namespace == "" || group == "errors" || group == "cluster" {
			return data, nil
		}

		var items []map[string]any
		err = json.Unmarshal(data, &items)
		if err != nil {
			return nil, err
		}

		filtered := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if item["namespace"] == namespace {
				filtered = append(filtered, item)
			}
		}

		return filtered, nil
	}
}
