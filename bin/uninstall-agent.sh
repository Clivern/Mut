#!/usr/bin/env bash
# Removes the Mut agent from the cluster of the current kubeconfig context.
set -euo pipefail

CONFIRMED=""

usage() {
  cat <<USAGE
Usage: $0 --yes

Removes the Mut agent: its Deployment, ConfigMap, Secret, ServiceAccount and RBAC.
The mut-system namespace is removed only when nothing else is left in it.

  --yes   Confirm the removal. Nothing is deleted without it.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes) CONFIRMED=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage >&2; exit 1 ;;
  esac
done

if [[ -z "${CONFIRMED}" ]]; then
  usage >&2
  exit 1
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required but was not found in PATH" >&2
  exit 1
fi

echo "Removing the Mut agent from context: $(kubectl config current-context)"

kubectl -n mut-system delete deployment/mut-agent secret/mut-agent-credentials \
  configmap/mut-agent-config serviceaccount/mut-agent --ignore-not-found
kubectl delete clusterrolebinding/mut-agent-reader clusterrole/mut-agent-reader --ignore-not-found

# Pods outlive their Deployment until they stop, so wait for them before looking for leftovers.
kubectl -n mut-system delete pods -l app.kubernetes.io/name=mut-agent --ignore-not-found --wait=true --timeout=120s

# Kubernetes adds these to every namespace, so they don't count as leftovers.
leftovers=$(kubectl -n mut-system get pods,services,deployments,replicasets,secrets,configmaps,serviceaccounts -o name 2>/dev/null \
  | grep -v -E '^(serviceaccount/default|configmap/kube-root-ca.crt)$' || true)

if [[ -z "${leftovers}" ]]; then
  kubectl delete namespace mut-system --ignore-not-found
  echo "Mut agent removed. Namespace mut-system deleted."
else
  echo "Mut agent removed. Namespace mut-system kept because it still has:"
  echo "${leftovers}"
fi
