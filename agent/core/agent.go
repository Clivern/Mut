// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/clivern/mut/agent/pkg/k8s"

	"github.com/rs/zerolog/log"
)

// Config describes one in-cluster agent and the Mut workspace cluster it reports for.
type Config struct {
	Server      string
	WorkspaceId string
	ClusterId   string
	AccessKey   string
	Interval    time.Duration
}

// Agent collects the local cluster state and pushes it to the Mut server.
type Agent struct {
	config  Config
	cluster *k8s.Cluster
	client  *http.Client
}

// NewAgent connects to the API server the pod runs in, using its service account.
func NewAgent(config Config) (*Agent, error) {
	cluster, err := k8s.NewCluster(k8s.Config{
		Name:      config.ClusterId,
		InCluster: os.Getenv("KUBERNETES_SERVICE_HOST") != "",
		QPS:       50,
		Burst:     100,
	})
	if err != nil {
		return nil, err
	}

	return &Agent{
		config:  config,
		cluster: cluster,
		client:  &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Run pushes a state immediately and then on every interval until ctx is done.
func (a *Agent) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	a.PushOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.PushOnce(ctx)
		}
	}
}

// Push uploads one cluster state to the Mut server.
func (a *Agent) Push(ctx context.Context, state *k8s.ClusterState) error {
	body, err := json.Marshal(state)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf(
		"%s/api/v1/workspaces/%s/clusters/%s/state",
		a.config.Server,
		a.config.WorkspaceId,
		a.config.ClusterId,
	), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Key", a.config.AccessKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

		return fmt.Errorf("push cluster state: status %d: %s", resp.StatusCode, msg)
	}

	return nil
}

// PushOnce collects and pushes one state, logging failures so the loop keeps running.
func (a *Agent) PushOnce(ctx context.Context) {
	state := a.cluster.Collect(ctx)

	err := a.Push(ctx, state)
	if err != nil {
		log.Error().
			Err(err).
			Str("cluster", a.config.ClusterId).
			Msg("Failed to push cluster state")

		return
	}

	log.Info().
		Str("cluster", a.config.ClusterId).
		Msg("Pushed cluster state")
}
