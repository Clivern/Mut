import { reactive } from 'vue'
import { cluster_api } from '@/api'
import { formatRelative } from '@/utils/date'

// Static cluster definitions. This is the temporary source of truth until the
// backend owns clusters; every board below reads from here.
export const seedClusters = reactive([
  { id: 'prod-us-east', region: 'us-east-1', state: 'ok', cpu: 61, mem: 72, open: 0, createdAt: '2024-11-02T09:00:00Z' },
  { id: 'prod-eu-west', region: 'eu-west-1', state: 'warn', cpu: 74, mem: 88, open: 1, createdAt: '2024-11-18T14:30:00Z' },
  { id: 'staging', region: 'us-west-2', state: 'warn', cpu: 52, mem: 60, open: 1, createdAt: '2025-01-06T11:15:00Z' },
  { id: 'dev-sandbox', region: 'us-west-2', state: 'ok', cpu: 18, mem: 31, open: 0, createdAt: '2025-03-21T08:45:00Z' },
  { id: 'edge-apac', region: 'ap-southeast-1', state: 'crit', cpu: 91, mem: 94, open: 1, createdAt: '2025-06-09T16:20:00Z' },
])

export const anomalies = [
  {
    id: 'a1',
    cluster: 'prod-eu-west',
    severity: 'critical',
    title: 'Memory keeps climbing in checkout-api',
    metric: 'Heap up 38% in 6 h',
    when: '12 min ago',
    cause: 'Growth starts at the v2.14.0 rollout at 09:12 and follows request volume.',
    fix: 'Roll back to v2.13.4, or raise the memory limit from 1.5Gi to 2Gi while the leak is traced.',
    actions: [
      { label: 'Roll back', message: 'Rollback to v2.13.4 queued for checkout-api.' },
      { label: 'Open trace', message: 'Opening heap trace for checkout-api.' },
    ],
  },
  {
    id: 'a2',
    cluster: 'edge-apac',
    severity: 'critical',
    title: 'Node pressure on 3 of 4 edge nodes',
    metric: 'Memory 94%, evictions rising',
    when: '3 min ago',
    cause: 'The ingest DaemonSet has no memory limit and grew after the new feed was turned on.',
    fix: 'Set a 512Mi limit and cordon edge-apac-03.',
    actions: [
      { label: 'Apply limit', message: '512Mi limit queued for the ingest DaemonSet.' },
      { label: 'Cordon node', message: 'Cordon queued for edge-apac-03.' },
    ],
  },
  {
    id: 'a3',
    cluster: 'prod-us-east',
    severity: 'warning',
    title: 'CPU throttling in payments-worker',
    metric: 'Throttled 22% of periods',
    when: '41 min ago',
    cause: 'The CPU limit is 1 core, but settlement bursts reach 2.4 cores.',
    fix: 'Raise the limit to 2.5 cores between 02:00 and 04:00.',
    actions: [
      { label: 'Apply window', message: 'Limit change scheduled for 02:00–04:00.' },
      { label: 'Dismiss', dismiss: true },
    ],
  },
  {
    id: 'a4',
    cluster: 'staging',
    severity: 'info',
    title: 'Autoscaler ran without a cap',
    metric: 'Scaled to 40 replicas',
    when: '2 h ago',
    cause: 'The load-test HPA has no maxReplicas, and the test ran for 90 minutes.',
    fix: 'Set maxReplicas to 12.',
    actions: [
      { label: 'Set max 12', message: 'maxReplicas set to 12 for load-test.' },
      { label: 'Dismiss', dismiss: true },
    ],
  },
]

export const recommendations = [
  { id: 'r1', workload: 'search-indexer', cluster: 'prod-us-east', change: 'CPU request 8 → 3', p95: '1.9 CPU', savings: 412 },
  { id: 'r2', workload: 'reporting-cron', cluster: 'dev-sandbox', change: '2 CPU / 4Gi → 0.5 CPU / 1Gi', p95: '0.3 CPU', savings: 230 },
  { id: 'r3', workload: 'batch-ml-train', cluster: 'staging', change: 'On-demand → spot node pool', p95: '—', savings: 640 },
  { id: 'r4', workload: 'checkout-api', cluster: 'prod-eu-west', change: 'Memory limit 1.5Gi → 1.2Gi', p95: '0.9Gi', savings: 96 },
  { id: 'r5', workload: 'image-resizer', cluster: 'edge-apac', change: 'CPU request 4 → 1.5', p95: '0.8 CPU', savings: 178 },
]

// Placeholder cluster state until the live collect feeds the board.
// Mirrors the example data in the Kubernetes State Board artifact.
export const fakeClusterDetail = {
  summary: {
    nodesReady: 5,
    nodesTotal: 6,
    podsRunning: 142,
    podsPending: 2,
    podsFailed: 1,
    replicasReady: 20,
    replicasDesired: 23,
    warnings: 4,
  },
  nodes: [
    { name: 'ip-10-0-1-12', role: 'worker', ready: true, memoryPressure: false, diskPressure: false, pidPressure: false, cpu: 82, memory: 91, pods: 30, kubelet: 'v1.30.4' },
    { name: 'ip-10-0-1-37', role: 'worker', ready: true, memoryPressure: false, diskPressure: false, pidPressure: false, cpu: 64, memory: 72, pods: 27, kubelet: 'v1.30.4' },
    { name: 'ip-10-0-2-08', role: 'worker', ready: true, memoryPressure: false, diskPressure: true, pidPressure: false, cpu: 77, memory: 85, pods: 33, kubelet: 'v1.30.4' },
    { name: 'ip-10-0-2-51', role: 'worker', ready: true, memoryPressure: false, diskPressure: false, pidPressure: false, cpu: 58, memory: 66, pods: 24, kubelet: 'v1.30.4' },
    { name: 'ip-10-0-3-19', role: 'worker', ready: true, memoryPressure: false, diskPressure: false, pidPressure: false, cpu: 69, memory: 80, pods: 28, kubelet: 'v1.29.7' },
    { name: 'ip-10-0-3-44', role: 'worker', ready: false, memoryPressure: false, diskPressure: false, pidPressure: false, cpu: 0, memory: 0, pods: 0, kubelet: 'v1.29.7' },
  ],
  pods: [
    { namespace: 'payments', name: 'checkout-api-7d9f8b-x2k4p', node: 'ip-10-0-1-12', phase: 'Running', ready: '1/2', restarts: 14, reason: 'CrashLoopBackOff' },
    { namespace: 'payments', name: 'checkout-api-7d9f8b-q8m2r', node: 'ip-10-0-1-37', phase: 'Running', ready: '2/2', restarts: 0, reason: '' },
    { namespace: 'payments', name: 'checkout-api-7d9f8b-z4n7t', node: 'ip-10-0-2-08', phase: 'Running', ready: '2/2', restarts: 1, reason: '' },
    { namespace: 'edge', name: 'ingest-daemon-k5w2v', node: 'ip-10-0-2-08', phase: 'Pending', ready: '0/1', restarts: 0, reason: 'ImagePullBackOff' },
    { namespace: 'edge', name: 'ingest-daemon-m3d9c', node: 'ip-10-0-3-19', phase: 'Running', ready: '1/1', restarts: 0, reason: '' },
    { namespace: 'search', name: 'search-indexer-1', node: 'ip-10-0-2-51', phase: 'Running', ready: '0/1', restarts: 0, reason: 'Unready' },
    { namespace: 'search', name: 'search-indexer-2', node: 'ip-10-0-3-19', phase: 'Pending', ready: '0/1', restarts: 0, reason: 'Unschedulable' },
    { namespace: 'ops', name: 'batch-report-28491630-abc12', node: 'ip-10-0-1-37', phase: 'Failed', ready: '0/1', restarts: 0, reason: 'Error' },
    { namespace: 'ops', name: 'batch-report-28491660-xyz98', node: 'ip-10-0-3-44', phase: 'Unknown', ready: '0/1', restarts: 0, reason: 'NodeLost' },
  ],
  events: [
    { type: 'Warning', reason: 'BackOff', object: 'pod/checkout-api-7d9f8b-x2k4p', namespace: 'payments', message: 'Back-off restarting failed container', count: 14, lastSeen: '2026-11-07T22:41:00Z' },
    { type: 'Warning', reason: 'Failed', object: 'pod/ingest-daemon-k5w2v', namespace: 'edge', message: 'Failed to pull image not found', count: 6, lastSeen: '2026-11-07T22:38:00Z' },
    { type: 'Warning', reason: 'FailedScheduling', object: 'pod/search-indexer-2', namespace: 'search', message: '0/6 nodes are available insufficient memory', count: 9, lastSeen: '2026-11-07T22:36:00Z' },
    { type: 'Warning', reason: 'NodeNotReady', object: 'node/ip-10-0-3-44', namespace: '', message: 'Node status is NotReady', count: 1, lastSeen: '2026-11-07T22:30:00Z' },
    { type: 'Warning', reason: 'Unhealthy', object: 'pod/search-indexer-1', namespace: 'search', message: 'Readiness probe failed', count: 4, lastSeen: '2026-11-07T21:58:00Z' },
    { type: 'Normal', reason: 'Pulled', object: 'pod/storefront-6b77c-2dp9l', namespace: 'web', message: 'Container image already present', count: 1, lastSeen: '2026-11-07T22:20:00Z' },
    { type: 'Normal', reason: 'ScalingReplicaSet', object: 'deployment/search-indexer', namespace: 'search', message: 'Scaled up replica set to 5', count: 1, lastSeen: '2026-11-07T22:15:00Z' },
  ],
  workloads: [
    { namespace: 'payments', name: 'checkout-api', kind: 'Deployment', ready: 2, desired: 3, image: 'checkout-api:2.14.0' },
    { namespace: 'edge', name: 'ingest-daemon', kind: 'DaemonSet', ready: 3, desired: 4, image: 'ingest:1.8.2' },
    { namespace: 'search', name: 'search-indexer', kind: 'Deployment', ready: 4, desired: 5, image: 'search-indexer:3.0.1' },
    { namespace: 'kube-system', name: 'coredns', kind: 'Deployment', ready: 2, desired: 2, image: 'coredns:1.11.1' },
    { namespace: 'auth', name: 'session-store', kind: 'StatefulSet', ready: 3, desired: 3, image: 'redis:7.2' },
    { namespace: 'web', name: 'storefront', kind: 'Deployment', ready: 6, desired: 6, image: 'storefront:5.1.0' },
  ],
}

export const cpuSeries =[40, 42, 45, 47, 50, 52, 55, 58, 61, 64, 66, 70, 72, 74, 73, 75, 78, 76, 74, 71, 73, 74, 75, 74]
export const memSeries = [52, 53, 55, 57, 58, 60, 62, 64, 66, 70, 74, 78, 80, 82, 84, 85, 87, 88, 88, 89, 88, 88, 89, 88]
export const deployIndex = 9

export function allClusters() {
  return seedClusters
}

// Fetches a live read-only snapshot of a workspace cluster from the backend.
// Throws when the cluster is missing or the collect fails; callers fall back
// to the static seeds below.
export async function fetchClusterState(workspaceId, clusterId) {
  const { data } = await cluster_api.state(workspaceId, clusterId)

  return data
}

// Adapts a backend ClusterState payload to the board shape (clusters,
// anomalies, recommendations) so ClusterBoard renders live data unchanged.
// Events become anomalies and running pods without limits become guardrail
// recommendations; usage meters stay at zero because the snapshot carries
// requests, not live utilization.
export function toBoardSnapshot(state, clusterId) {
  const events = state.events || []
  const warnings = events.filter((event) => event.type === 'Warning')

  const seen = new Set()
  const recommendations = (state.pods || [])
    .filter((pod) => pod.phase === 'Running' && !Object.keys(pod.limits || {}).length)
    .map((pod) => pod.ownerName || pod.name)
    .filter((name) => {
      if (seen.has(name)) return false
      seen.add(name)

      return true
    })
    .slice(0, 8)
    .map((name, index) => ({
      id: `live-rec-${index}`,
      workload: name,
      cluster: clusterId,
      change: 'Set CPU and memory limits',
      p95: '—',
      savings: 0,
    }))

  return {
    clusters: [
      {
        id: clusterId,
        name: state.cluster?.name || clusterId,
        region: '',
        state: warnings.length ? 'warn' : 'ok',
        cpu: 0,
        mem: 0,
        open: warnings.length,
        createdAt: state.cluster?.collectedAt,
      },
    ],
    anomalies: events.slice(0, 20).map((event, index) => ({
      id: `live-${index}`,
      cluster: clusterId,
      severity: event.type === 'Warning' ? 'warning' : 'info',
      title: event.reason || event.kind || 'Event',
      metric: `${event.count || 1}× ${event.kind || ''} ${event.name || ''}`.trim(),
      when: formatRelative(event.lastSeen, event.lastSeen),
      cause: (event.message || '').slice(0, 240),
      fix: '',
      actions: [],
    })),
    recommendations,
  }
}
