# Mut Agent

`mut-agent` is an in-cluster agent that collects Kubernetes cluster state and pushes it to the Mut server. It runs as a Deployment in the `mut-system` namespace.

## Setup

1. **Create an access key.** In Mut, open the workspace, go to **Clusters**, and click **Setup** on the cluster row. Click **Create access key**. The key is created with the `clusters:write` permission and is shown only once, so copy it before you close the dialog.

2. **Run the install command against the cluster.** Select the cluster's kubeconfig context, then run the install script with the values from the setup dialog:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/Clivern/Mut/main/bin/install-agent.sh | bash -s -- \
     --server https://<MUT_SERVER> \
     --workspace <WORKSPACE_ID> \
     --cluster <CLUSTER_ID> \
     --access-key <ACCESS_KEY>
   ```

   The script requires `kubectl` in your `PATH`. It applies the manifests to the current kubeconfig context. Optional flags:

   - `--version` sets the `mut-agent` image tag (default: `0.1.0`).
   - `--interval` sets how often the agent pushes state (default: `300s`).

3. **Wait for the first push.** The agent pushes cluster state within a minute. The cluster row in Mut then shows **Dashboard** instead of **Setup**.

## Remove

Removing a cluster in Mut deletes the cluster record and every state it has stored. It does not stop the agent, which keeps running in the cluster until you uninstall it.

1. **Remove the cluster in Mut.** Open **Clusters**, click **Remove** on the cluster row, and confirm.

2. **Uninstall the agent from the cluster.** Select the cluster's kubeconfig context, then run:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/Clivern/Mut/main/bin/uninstall-agent.sh | bash -s -- --yes
   ```

   The script deletes the agent's Deployment, Secret, ConfigMap, ServiceAccount, and RBAC. It deletes the `mut-system` namespace only when nothing else is left in it. Nothing is deleted without `--yes`.
