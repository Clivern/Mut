#!/usr/bin/env bash
# Installs the Mut agent into the cluster of the current kubeconfig context.
set -euo pipefail

SERVER=""
WORKSPACE=""
CLUSTER=""
ACCESS_KEY=""
VERSION="0.1.0"
INTERVAL="300s"

usage() {
  cat <<USAGE
Usage: $0 --server URL --workspace ID --cluster ID --access-key KEY [--version TAG] [--interval DURATION]

  --server      Base URL of the Mut server, e.g. https://mut.example.com
  --workspace   Workspace id that owns the cluster
  --cluster     Cluster id registered in the workspace
  --access-key  Workspace access key with the clusters:write permission
  --version     mut-agent image tag to install (default: ${VERSION})
  --interval    How often the agent pushes cluster state (default: ${INTERVAL})
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server) SERVER="$2"; shift 2 ;;
    --workspace) WORKSPACE="$2"; shift 2 ;;
    --cluster) CLUSTER="$2"; shift 2 ;;
    --access-key) ACCESS_KEY="$2"; shift 2 ;;
    --version) VERSION="$2"; shift 2 ;;
    --interval) INTERVAL="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) usage >&2; exit 1 ;;
  esac
done

: "${SERVER:?--server is required}"
: "${WORKSPACE:?--workspace is required}"
: "${CLUSTER:?--cluster is required}"
: "${ACCESS_KEY:?--access-key is required}"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required but was not found in PATH" >&2
  exit 1
fi

echo "Installing the Mut agent into context: $(kubectl config current-context)"

kubectl apply -f - <<EOF_MANIFEST
apiVersion: v1
kind: Namespace
metadata:
  name: mut-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mut-agent
  namespace: mut-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mut-agent-reader
rules:
  - apiGroups: [""]
    resources:
      - namespaces
      - nodes
      - pods
      - services
      - configmaps
      - secrets
      - persistentvolumes
      - persistentvolumeclaims
      - serviceaccounts
      - resourcequotas
      - limitranges
      - endpoints
      - events
    verbs: ["get", "list"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
    verbs: ["get", "list"]
  - apiGroups: ["batch"]
    resources: ["jobs", "cronjobs"]
    verbs: ["get", "list"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses", "ingressclasses", "networkpolicies"]
    verbs: ["get", "list"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list"]
  - apiGroups: ["autoscaling"]
    resources: ["horizontalpodautoscalers"]
    verbs: ["get", "list"]
  - apiGroups: ["policy"]
    resources: ["poddisruptionbudgets"]
    verbs: ["get", "list"]
  - apiGroups: ["discovery.k8s.io"]
    resources: ["endpointslices"]
    verbs: ["get", "list"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings", "clusterroles", "clusterrolebindings"]
    verbs: ["get", "list"]
  - apiGroups: ["scheduling.k8s.io"]
    resources: ["priorityclasses"]
    verbs: ["get", "list"]
  - apiGroups: ["node.k8s.io"]
    resources: ["runtimeclasses"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mut-agent-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mut-agent-reader
subjects:
  - kind: ServiceAccount
    name: mut-agent
    namespace: mut-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mut-agent-config
  namespace: mut-system
data:
  SERVER: "${SERVER}"
  WORKSPACE: "${WORKSPACE}"
  CLUSTER: "${CLUSTER}"
  INTERVAL: "${INTERVAL}"
---
apiVersion: v1
kind: Secret
metadata:
  name: mut-agent-credentials
  namespace: mut-system
type: Opaque
stringData:
  ACCESS_KEY: "${ACCESS_KEY}"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mut-agent
  namespace: mut-system
  labels:
    app.kubernetes.io/name: mut-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: mut-agent
  template:
    metadata:
      labels:
        app.kubernetes.io/name: mut-agent
    spec:
      serviceAccountName: mut-agent
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: agent
          image: clivern/mut-agent:${VERSION}
          envFrom:
            - configMapRef:
                name: mut-agent-config
            - secretRef:
                name: mut-agent-credentials
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
EOF_MANIFEST

# Restart so pods pick up a rotated access key, since envFrom reads Secrets only at start.
kubectl -n mut-system rollout restart deployment/mut-agent
kubectl -n mut-system rollout status deployment/mut-agent --timeout=180s

echo "Mut agent installed. Cluster state will appear in Mut for cluster ${CLUSTER}."
