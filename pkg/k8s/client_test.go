// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKubeconfig = `apiVersion: v1
kind: Config
current-context: dev
clusters:
- name: dev
  cluster:
    server: https://dev.example.com:6443
- name: prod
  cluster:
    server: https://prod.example.com:6443
contexts:
- name: dev
  context:
    cluster: dev
    user: admin
- name: prod
  context:
    cluster: prod
    user: admin
users:
- name: admin
  user:
    token: secret
`

func TestUnitLoadFromKubeconfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, os.WriteFile(path, []byte(testKubeconfig), 0600))

	restConfig, err := load(Config{Name: "dev", Kubeconfig: path, QPS: 20, Burst: 40})
	require.NoError(t, err)
	assert.Equal(t, "https://dev.example.com:6443", restConfig.Host)
	assert.Equal(t, float32(20), restConfig.QPS)
	assert.Equal(t, 40, restConfig.Burst)

	restConfig, err = load(Config{Name: "prod", Kubeconfig: path, Context: "prod"})
	require.NoError(t, err)
	assert.Equal(t, "https://prod.example.com:6443", restConfig.Host)
}

func TestUnitLoadInClusterOutsidePod(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	_, err := load(Config{Name: "local", InCluster: true})
	assert.Error(t, err)
}

func TestUnitNewCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"major":"1","minor":"34","gitVersion":"v1.34.0"}`))
	}))
	defer server.Close()

	config := strings.Replace(testKubeconfig, "https://dev.example.com:6443", server.URL, 1)
	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, os.WriteFile(path, []byte(config), 0600))

	cluster, err := NewCluster(Config{Name: "dev", Kubeconfig: path})
	require.NoError(t, err)
	assert.Equal(t, "dev", cluster.Config().Name)

	_, err = NewCluster(Config{Name: "prod", Kubeconfig: path, Context: "prod"})
	assert.Error(t, err)
}
