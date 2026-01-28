// Placeholder resources shaped like the ClusterState JSON the backend returns.
// Swap for the live collect once it feeds the board. Secrets carry key names only.
export const fakeClusterResources = {
  cluster: { name: 'prod-eu-west', serverVersion: 'v1.30.4', collectedAt: '2026-11-07T22:45:00Z' },
  errors: {
    policies: 'forbidden: cannot list resource "networkpolicies" in API group "networking.k8s.io"',
  },
  namespaces: [
    { name: 'payments', phase: 'Active' },
    { name: 'edge', phase: 'Active' },
    { name: 'search', phase: 'Active' },
    { name: 'ops', phase: 'Terminating' },
  ],
  statefulSets: [
    { namespace: 'auth', name: 'session-store', serviceName: 'session-store', replicasDesired: 3, replicasReady: 3 },
  ],
  daemonSets: [
    { namespace: 'edge', name: 'ingest-daemon', desiredScheduled: 4, readyScheduled: 3, updatedScheduled: 4, unavailableScheduled: 1 },
  ],
  replicaSets: [
    { namespace: 'payments', name: 'checkout-api-7d9f8b', replicasDesired: 3, replicasReady: 2 },
    { namespace: 'search', name: 'search-indexer-6c4f5', replicasDesired: 5, replicasReady: 4 },
  ],
  jobs: [
    { namespace: 'ops', name: 'batch-report-28491630', active: 0, succeeded: 0, failed: 1 },
    { namespace: 'ops', name: 'batch-report-28491660', active: 1, succeeded: 0, failed: 0 },
    { namespace: 'web', name: 'thumbnail-backfill', active: 0, succeeded: 12, failed: 0 },
  ],
  cronJobs: [
    { namespace: 'ops', name: 'batch-report', schedule: '30 2 * * *', suspend: false, active: 1, lastSchedule: '2026-11-07T02:30:00Z' },
    { namespace: 'web', name: 'sitemap-refresh', schedule: '0 */6 * * *', suspend: true, active: 0, lastSchedule: '2026-11-06T18:00:00Z' },
  ],
  services: [
    { namespace: 'payments', name: 'checkout-api', type: 'ClusterIP', clusterIP: '10.100.4.12', ports: [{ port: 8080, protocol: 'TCP' }] },
    { namespace: 'web', name: 'storefront', type: 'LoadBalancer', clusterIP: '10.100.8.40', ports: [{ port: 443, protocol: 'TCP' }, { port: 80, protocol: 'TCP' }] },
    { namespace: 'auth', name: 'session-store', type: 'ClusterIP', clusterIP: '10.100.2.77', ports: [{ port: 6379, protocol: 'TCP' }] },
  ],
  endpoints: [
    { namespace: 'payments', name: 'checkout-api', service: 'checkout-api', addresses: ['10.0.1.23', '10.0.1.48', '10.0.2.11'], ready: 2, serving: 2, terminating: 1 },
    { namespace: 'web', name: 'storefront', service: 'storefront', addresses: ['10.0.1.61', '10.0.2.30', '10.0.3.12'], ready: 3, serving: 3, terminating: 0 },
    { namespace: 'search', name: 'search-indexer', service: 'search-indexer', addresses: [], ready: 0, serving: 0, terminating: 0 },
  ],
  ingresses: [
    { namespace: 'web', name: 'storefront', class: 'nginx', rules: [{ host: 'shop.example.com', paths: ['/'] }], tls: ['shop.example.com'] },
    { namespace: 'payments', name: 'checkout', class: 'nginx', rules: [{ host: 'pay.example.com', paths: ['/api', '/health'] }], tls: [] },
  ],
  ingressClasses: [
    { name: 'nginx', controller: 'k8s.io/ingress-nginx', default: true },
    { name: 'alb', controller: 'ingress.k8s.aws/alb', default: false },
  ],
  configMaps: [
    { namespace: 'payments', name: 'checkout-config', keys: ['LOG_LEVEL', 'FEATURE_FLAGS'] },
    { namespace: 'edge', name: 'ingest-feed', keys: ['feed.yaml'] },
  ],
  secrets: [
    { namespace: 'payments', name: 'checkout-db', type: 'Opaque', keys: ['username', 'password'] },
    { namespace: 'web', name: 'storefront-tls', type: 'kubernetes.io/tls', keys: ['tls.crt', 'tls.key'] },
  ],
  volumes: [
    { name: 'pvc-3f2a9c1e', capacity: '20Gi', phase: 'Bound', storageClass: 'gp3', claim: 'auth/session-store-data' },
    { name: 'pvc-9b1d7e44', capacity: '5Gi', phase: 'Released', storageClass: 'gp3', claim: 'ops/reports-cache' },
  ],
  claims: [
    { namespace: 'auth', name: 'session-store-data', phase: 'Bound', storageClass: 'gp3', volume: 'pvc-3f2a9c1e', requested: '20Gi' },
    { namespace: 'search', name: 'index-cache', phase: 'Pending', storageClass: 'gp3', volume: '', requested: '50Gi' },
  ],
  storageClasses: [
    { name: 'gp3', provisioner: 'ebs.csi.aws.com', reclaim: 'Delete', expansion: true },
    { name: 'standard', provisioner: 'kubernetes.io/no-provisioner', reclaim: 'Retain', expansion: false },
  ],
  autoscalers: [
    { namespace: 'payments', name: 'checkout-api', targetKind: 'Deployment', targetName: 'checkout-api', minReplicas: 2, maxReplicas: 6, currentReplicas: 3, desiredReplicas: 4 },
    { namespace: 'search', name: 'search-indexer', targetKind: 'Deployment', targetName: 'search-indexer', minReplicas: 3, maxReplicas: 5, currentReplicas: 5, desiredReplicas: 5 },
  ],
  disruptions: [
    { namespace: 'payments', name: 'checkout-api', allowed: 1, currentHealthy: 2, desiredHealthy: 2, expectedPods: 3 },
    { namespace: 'auth', name: 'session-store', allowed: 0, currentHealthy: 3, desiredHealthy: 3, expectedPods: 3 },
  ],
  quotas: [
    { namespace: 'payments', name: 'compute', hard: { cpu: '16', memory: '32Gi' }, used: { cpu: '11', memory: '24Gi' } },
  ],
  limits: [
    { namespace: 'edge', name: 'default-limits', defaults: { cpu: '500m', memory: '512Mi' } },
  ],
  priorities: [
    { name: 'system-critical', value: 2000000000, globalDefault: false },
    { name: 'standard', value: 1000, globalDefault: true },
  ],
  policies: [],
  accounts: [
    { namespace: 'payments', name: 'checkout-api', secrets: ['checkout-api-token-x7k2q'] },
  ],
  roles: [
    { namespace: 'ops', name: 'report-runner', rules: [{ groups: [''], resources: ['configmaps'], verbs: ['get', 'list'] }] },
  ],
  clusterRoles: [
    { name: 'view', rules: [{ groups: [''], resources: ['pods', 'services'], verbs: ['get', 'list', 'watch'] }] },
  ],
  bindings: [
    { namespace: 'ops', name: 'report-runner', roleKind: 'Role', roleName: 'report-runner', subjects: [{ kind: 'ServiceAccount', name: 'report-runner', namespace: 'ops' }] },
  ],
  clusterBindings: [
    { name: 'ops-view', roleKind: 'ClusterRole', roleName: 'view', subjects: [{ kind: 'Group', name: 'ops-team' }] },
  ],
  runtimeClasses: [
    { name: 'gvisor', handler: 'runsc' },
  ],
}

// Groups that are not namespaced. The namespace filter leaves them unchanged.
export const clusterScopedSources = new Set([
  'namespaces',
  'volumes',
  'storageClasses',
  'ingressClasses',
  'priorities',
  'runtimeClasses',
  'clusterRoles',
  'clusterBindings',
])

const strong = 'font-medium text-theme-text'
const muted = 'text-theme-textLight'
const warn = 'font-semibold text-amber-700'
const bad = 'font-semibold text-red-700'

const namespaceColumn = { label: 'Namespace', value: (row) => row.namespace || '—', tone: () => muted }
const nameColumn = { label: 'Name', value: (row) => row.name, tone: () => strong }

// Maps a ClusterState group to its table: which JSON list it reads, its
// translated title and description, and the columns to show.
export function buildResourceGroups(t) {
  const yesNo = (value) => (value ? t('dashboard.flag_yes') : t('dashboard.flag_no'))
  const pairs = (map) => Object.entries(map ?? {}).map(([key, value]) => `${key} ${value}`).join(', ')
  const group = (source, columns) => ({
    source,
    title: t(`dashboard.resources.${source}.title`),
    description: t(`dashboard.resources.${source}.desc`),
    columns,
  })

  return [
    group('namespaces', [
      nameColumn,
      { label: 'Phase', value: (row) => row.phase, tone: (row) => (row.phase === 'Active' ? muted : warn) },
    ]),
    group('statefulSets', [
      namespaceColumn,
      nameColumn,
      { label: 'Service', value: (row) => row.serviceName, tone: () => muted },
      { label: 'Ready', value: (row) => row.replicasReady, tone: (row) => (row.replicasReady < row.replicasDesired ? bad : muted) },
      { label: 'Desired', value: (row) => row.replicasDesired, tone: () => muted },
    ]),
    group('daemonSets', [
      namespaceColumn,
      nameColumn,
      { label: 'Desired', value: (row) => row.desiredScheduled, tone: () => muted },
      { label: 'Ready', value: (row) => row.readyScheduled, tone: (row) => (row.readyScheduled < row.desiredScheduled ? bad : muted) },
      { label: 'Updated', value: (row) => row.updatedScheduled, tone: () => muted },
      { label: 'Unavailable', value: (row) => row.unavailableScheduled, tone: (row) => (row.unavailableScheduled ? bad : muted) },
    ]),
    group('replicaSets', [
      namespaceColumn,
      nameColumn,
      { label: 'Ready', value: (row) => row.replicasReady, tone: (row) => (row.replicasReady < row.replicasDesired ? bad : muted) },
      { label: 'Desired', value: (row) => row.replicasDesired, tone: () => muted },
    ]),
    group('jobs', [
      namespaceColumn,
      nameColumn,
      { label: 'Active', value: (row) => row.active, tone: () => muted },
      { label: 'Succeeded', value: (row) => row.succeeded, tone: () => muted },
      { label: 'Failed', value: (row) => row.failed, tone: (row) => (row.failed ? bad : muted) },
    ]),
    group('cronJobs', [
      namespaceColumn,
      nameColumn,
      { label: 'Schedule', value: (row) => row.schedule, tone: () => muted },
      { label: 'Suspended', value: (row) => yesNo(row.suspend), tone: (row) => (row.suspend ? warn : muted) },
      { label: 'Active', value: (row) => row.active, tone: () => muted },
      { label: 'Last run', value: (row) => row.lastSchedule, tone: () => muted },
    ]),
    group('services', [
      namespaceColumn,
      nameColumn,
      { label: 'Type', value: (row) => row.type, tone: () => muted },
      { label: 'Cluster IP', value: (row) => row.clusterIP, tone: () => muted },
      { label: 'Ports', value: (row) => row.ports.map((port) => `${port.port}/${port.protocol}`).join(', '), tone: () => muted },
    ]),
    group('endpoints', [
      namespaceColumn,
      { label: 'Service', value: (row) => row.service, tone: () => strong },
      { label: 'Addresses', value: (row) => row.addresses.length, tone: () => muted },
      { label: 'Ready', value: (row) => row.ready, tone: (row) => (row.ready ? muted : bad) },
      { label: 'Serving', value: (row) => row.serving, tone: () => muted },
      { label: 'Terminating', value: (row) => row.terminating, tone: () => muted },
    ]),
    group('ingresses', [
      namespaceColumn,
      nameColumn,
      { label: 'Class', value: (row) => row.class, tone: () => muted },
      { label: 'Hosts', value: (row) => row.rules.map((rule) => rule.host).join(', '), tone: () => muted },
      { label: 'Paths', value: (row) => row.rules.flatMap((rule) => rule.paths).join(', '), tone: () => muted },
      { label: 'TLS', value: (row) => (row.tls.length ? row.tls.join(', ') : '—'), tone: () => muted },
    ]),
    group('ingressClasses', [
      nameColumn,
      { label: 'Controller', value: (row) => row.controller, tone: () => muted },
      { label: 'Default', value: (row) => yesNo(row.default), tone: () => muted },
    ]),
    group('configMaps', [
      namespaceColumn,
      nameColumn,
      { label: 'Keys', value: (row) => row.keys.join(', '), tone: () => muted },
    ]),
    group('secrets', [
      namespaceColumn,
      nameColumn,
      { label: 'Type', value: (row) => row.type, tone: () => muted },
      { label: 'Keys', value: (row) => row.keys.join(', '), tone: () => muted },
    ]),
    group('volumes', [
      nameColumn,
      { label: 'Capacity', value: (row) => row.capacity, tone: () => muted },
      { label: 'Phase', value: (row) => row.phase, tone: (row) => (row.phase === 'Bound' ? muted : warn) },
      { label: 'Storage class', value: (row) => row.storageClass, tone: () => muted },
      { label: 'Claim', value: (row) => row.claim, tone: () => muted },
    ]),
    group('claims', [
      namespaceColumn,
      nameColumn,
      { label: 'Phase', value: (row) => row.phase, tone: (row) => (row.phase === 'Bound' ? muted : warn) },
      { label: 'Requested', value: (row) => row.requested, tone: () => muted },
      { label: 'Storage class', value: (row) => row.storageClass, tone: () => muted },
      { label: 'Volume', value: (row) => row.volume || '—', tone: () => muted },
    ]),
    group('storageClasses', [
      nameColumn,
      { label: 'Provisioner', value: (row) => row.provisioner, tone: () => muted },
      { label: 'Reclaim', value: (row) => row.reclaim, tone: () => muted },
      { label: 'Expansion', value: (row) => yesNo(row.expansion), tone: () => muted },
    ]),
    group('autoscalers', [
      namespaceColumn,
      nameColumn,
      { label: 'Target', value: (row) => `${row.targetKind}/${row.targetName}`, tone: () => muted },
      { label: 'Min', value: (row) => row.minReplicas, tone: () => muted },
      { label: 'Max', value: (row) => row.maxReplicas, tone: () => muted },
      { label: 'Current', value: (row) => row.currentReplicas, tone: () => muted },
      { label: 'Desired', value: (row) => row.desiredReplicas, tone: (row) => (row.desiredReplicas !== row.currentReplicas ? warn : muted) },
    ]),
    group('disruptions', [
      namespaceColumn,
      nameColumn,
      { label: 'Allowed', value: (row) => row.allowed, tone: (row) => (row.allowed ? muted : warn) },
      { label: 'Healthy', value: (row) => row.currentHealthy, tone: () => muted },
      { label: 'Desired healthy', value: (row) => row.desiredHealthy, tone: () => muted },
      { label: 'Expected pods', value: (row) => row.expectedPods, tone: () => muted },
    ]),
    group('quotas', [
      namespaceColumn,
      nameColumn,
      { label: 'Used', value: (row) => pairs(row.used), tone: () => muted },
      { label: 'Hard', value: (row) => pairs(row.hard), tone: () => muted },
    ]),
    group('limits', [
      namespaceColumn,
      nameColumn,
      { label: 'Defaults', value: (row) => pairs(row.defaults), tone: () => muted },
    ]),
    group('priorities', [
      nameColumn,
      { label: 'Value', value: (row) => row.value, tone: () => muted },
      { label: 'Global default', value: (row) => yesNo(row.globalDefault), tone: () => muted },
    ]),
    group('policies', [
      namespaceColumn,
      nameColumn,
      { label: 'Ingress', value: (row) => yesNo(row.ingress), tone: () => muted },
      { label: 'Egress', value: (row) => yesNo(row.egress), tone: () => muted },
    ]),
    group('accounts', [
      namespaceColumn,
      nameColumn,
      { label: 'Secrets', value: (row) => row.secrets.join(', ') || '—', tone: () => muted },
    ]),
    group('roles', [
      namespaceColumn,
      nameColumn,
      { label: 'Rules', value: (row) => row.rules.length, tone: () => muted },
      { label: 'Verbs', value: (row) => [...new Set(row.rules.flatMap((rule) => rule.verbs))].join(', '), tone: () => muted },
    ]),
    group('clusterRoles', [
      nameColumn,
      { label: 'Rules', value: (row) => row.rules.length, tone: () => muted },
      { label: 'Verbs', value: (row) => [...new Set(row.rules.flatMap((rule) => rule.verbs))].join(', '), tone: () => muted },
    ]),
    group('bindings', [
      namespaceColumn,
      nameColumn,
      { label: 'Role', value: (row) => `${row.roleKind}/${row.roleName}`, tone: () => muted },
      { label: 'Subjects', value: (row) => row.subjects.map((subject) => `${subject.kind}:${subject.name}`).join(', '), tone: () => muted },
    ]),
    group('clusterBindings', [
      nameColumn,
      { label: 'Role', value: (row) => `${row.roleKind}/${row.roleName}`, tone: () => muted },
      { label: 'Subjects', value: (row) => row.subjects.map((subject) => `${subject.kind}:${subject.name}`).join(', '), tone: () => muted },
    ]),
    group('runtimeClasses', [
      nameColumn,
      { label: 'Handler', value: (row) => row.handler, tone: () => muted },
    ]),
  ]
}
