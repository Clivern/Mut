<template>
  <div class="space-y-8">
    <section v-if="!isCluster" class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4" :aria-label="$t('dashboard.summary')">
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.healthy_clusters') }}</p>
        <p class="stat-value tabular-nums">{{ healthyCount }} / {{ scopeClusters.length }}</p>
        <p class="mt-2 text-sm text-theme-textLight">{{ $t('dashboard.healthy_clusters_note') }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.open_anomalies') }}</p>
        <p class="stat-value tabular-nums">{{ openAnomalies.length }}</p>
        <p class="mt-2 text-sm text-red-700">{{ $t('dashboard.critical_count', { count: criticalCount }) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.monthly_savings') }}</p>
        <p class="stat-value tabular-nums">${{ pendingSavings.toLocaleString('en-US') }}</p>
        <p class="mt-2 text-sm text-theme-textLight">{{ $t('dashboard.recommendations_open', { count: pendingRecs.length }) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.idle_cpu') }}</p>
        <p class="stat-value tabular-nums">62%</p>
        <p class="mt-2 text-sm text-theme-textLight">{{ $t('dashboard.idle_cpu_note') }}</p>
      </div>
    </section>

    <section v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4" :aria-label="$t('dashboard.summary')">
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.cluster_nodes_ready') }}</p>
        <p class="stat-value tabular-nums">{{ detail.summary.nodesReady }} / {{ detail.summary.nodesTotal }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.cluster_pods_running') }}</p>
        <p class="stat-value tabular-nums">{{ detail.summary.podsRunning }}</p>
        <p class="mt-2 text-sm text-theme-textLight">
          {{ $t('dashboard.cluster_pods_detail', { pending: detail.summary.podsPending, failed: detail.summary.podsFailed }) }}
        </p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.cluster_replicas_ready') }}</p>
        <p class="stat-value tabular-nums">{{ detail.summary.replicasReady }} / {{ detail.summary.replicasDesired }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">{{ $t('dashboard.cluster_warnings') }}</p>
        <p class="stat-value tabular-nums">{{ detail.summary.warnings }}</p>
      </div>
    </section>

    <div class="grid grid-cols-1 gap-8 xl:grid-cols-5">
      <section class="section overflow-hidden p-0" :class="isCluster ? 'xl:col-span-5' : 'xl:col-span-3'" aria-labelledby="anomalies-title">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-theme-border px-6 py-4">
          <div class="min-w-0">
            <h2 id="anomalies-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.anomalies') }}</h2>
            <p class="mt-0.5 text-sm text-theme-textLight">
              {{ activeScope ? $t('dashboard.showing_cluster', { cluster: activeScope }) : $t('dashboard.anomalies_desc') }}
            </p>
          </div>
          <button
            v-if="activeScope && !clusterId"
            type="button"
            class="rounded-md px-2 py-1 text-sm text-theme-textLight transition hover:bg-theme-hover hover:text-theme-text"
            @click="selectedId = null"
          >
            {{ $t('dashboard.all_clusters') }}
          </button>
        </div>

        <ul v-if="visibleAnomalies.length" class="divide-y divide-theme-border">
          <li v-for="anomaly in visibleAnomalies" :key="anomaly.id" class="flex gap-3 px-6 py-5">
            <span
              class="mt-2 h-2 w-2 shrink-0 rounded-full"
              :class="severityClass[anomaly.severity]"
              aria-hidden="true"
            />
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="font-medium text-theme-text">{{ anomaly.title }}</p>
                <span class="rounded-full border border-theme-border px-2 py-0.5 text-xs text-theme-textLight">{{ anomaly.cluster }}</span>
              </div>
              <p class="mt-0.5 text-sm text-theme-textLight">{{ anomaly.metric }} · {{ anomaly.when }}</p>
              <p class="mt-2 text-sm leading-relaxed text-theme-text">
                <span class="mr-1 text-xs font-semibold uppercase tracking-wide text-theme-textLight">{{ $t('dashboard.cause') }}</span>{{ anomaly.cause }}
              </p>
              <p v-if="anomaly.fix" class="mt-1 text-sm leading-relaxed text-theme-text">
                <span class="mr-1 text-xs font-semibold uppercase tracking-wide text-theme-textLight">{{ $t('dashboard.fix') }}</span>{{ anomaly.fix }}
              </p>
              <div v-if="anomaly.actions && anomaly.actions.length" class="mt-3 flex flex-wrap gap-2">
                <button
                  v-for="action in anomaly.actions"
                  :key="action.label"
                  type="button"
                  class="btn-secondary"
                  @click="runAction(anomaly, action)"
                >
                  {{ action.label }}
                </button>
              </div>
            </div>
          </li>
        </ul>
        <p v-else class="px-6 py-10 text-center text-sm text-theme-textLight">{{ $t('dashboard.no_anomalies') }}</p>
      </section>

      <section v-if="!isCluster" class="section xl:col-span-2" aria-labelledby="trend-title">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 id="trend-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.cpu_memory') }}</h2>
            <p class="mt-0.5 text-sm text-theme-textLight">{{ trendCaption }}</p>
          </div>
          <div class="inline-flex rounded-lg border border-theme-border bg-theme-hover p-0.5" role="group" :aria-label="$t('dashboard.time_range')">
            <button
              v-for="range in ranges"
              :key="range"
              type="button"
              class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
              :class="activeRange === range ? 'bg-theme-surface text-theme-text shadow-sm' : 'text-theme-textLight hover:text-theme-text'"
              :aria-pressed="activeRange === range"
              @click="activeRange = range"
            >
              {{ range }}
            </button>
          </div>
        </div>
        <svg class="mt-4 h-auto w-full" viewBox="0 0 560 240" role="img" :aria-label="$t('dashboard.cpu_memory')">
          <g class="text-[11px] fill-theme-textLight">
            <line v-for="tick in trend.ticks" :key="tick" x1="36" x2="548" :y1="yFor(tick)" :y2="yFor(tick)" class="stroke-theme-border" />
            <text v-for="tick in trend.ticks" :key="`t-${tick}`" x="28" :y="yFor(tick) + 4" text-anchor="end">{{ tick }}%</text>
            <text v-for="label in axisLabels" :key="label.key" :x="label.x" y="226" text-anchor="middle">{{ label.text }}</text>
          </g>
          <line :x1="deployX" :x2="deployX" y1="14" y2="206" class="stroke-theme-textLight" stroke-dasharray="3 4" />
          <text :x="deployX + 6" y="24" class="fill-theme-text text-[11px]">{{ $t('dashboard.deploy_marker') }}</text>
          <path :d="memoryPath" fill="none" class="stroke-amber-500" stroke-width="2" stroke-linejoin="round" />
          <path :d="cpuPath" fill="none" class="stroke-theme-text" stroke-width="2" stroke-linejoin="round" />
        </svg>
        <div class="mt-3 flex flex-wrap gap-4 text-xs text-theme-textLight">
          <span class="inline-flex items-center gap-1.5"><i class="h-2 w-2 rounded-full bg-theme-text" />{{ $t('dashboard.legend_cpu') }}</span>
          <span class="inline-flex items-center gap-1.5"><i class="h-2 w-2 rounded-full bg-amber-500" />{{ $t('dashboard.legend_memory') }}</span>
        </div>
      </section>
    </div>

    <section v-if="isCluster" class="section overflow-hidden p-0" aria-labelledby="nodes-title">
      <div class="border-b border-theme-border px-6 py-4">
        <h2 id="nodes-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.cluster_nodes') }}</h2>
        <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.cluster_nodes_desc') }}</p>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-xs font-medium text-theme-textLight">
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_node') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_role') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_ready') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.memory_pressure') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.disk_pressure') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.pid_pressure') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.cpu') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.memory') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_pods') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_kubelet') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-theme-border">
            <tr v-for="node in detail.nodes" :key="node.name" class="text-theme-text">
              <td class="px-6 py-3 font-medium">{{ node.name }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ node.role }}</td>
              <td class="px-6 py-3" :class="flagClass(!node.ready)">{{ flagText(node.ready) }}</td>
              <td class="px-6 py-3" :class="flagClass(node.memoryPressure)">{{ flagText(node.memoryPressure) }}</td>
              <td class="px-6 py-3" :class="flagClass(node.diskPressure)">{{ flagText(node.diskPressure) }}</td>
              <td class="px-6 py-3" :class="flagClass(node.pidPressure)">{{ flagText(node.pidPressure) }}</td>
              <td class="px-6 py-3 text-right tabular-nums">{{ node.cpu }}%</td>
              <td class="px-6 py-3 text-right tabular-nums">{{ node.memory }}%</td>
              <td class="px-6 py-3 text-right tabular-nums">{{ node.pods }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ node.kubelet }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div v-if="isCluster" class="grid grid-cols-1 gap-8 xl:grid-cols-2">
      <section class="section overflow-hidden p-0" aria-labelledby="pods-title">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 id="pods-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.cluster_pods_unhealthy') }}</h2>
          <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.cluster_pods_unhealthy_desc') }}</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-xs font-medium text-theme-textLight">
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_namespace') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_pod') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_phase') }}</th>
                <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_ready') }}</th>
                <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_restarts') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_reason') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-theme-border">
              <tr v-for="pod in visiblePods" :key="pod.name" class="text-theme-text">
                <td class="px-6 py-3 text-theme-textLight">{{ pod.namespace }}</td>
                <td class="px-6 py-3 font-medium">{{ pod.name }}</td>
                <td class="px-6 py-3" :class="phaseClass(pod.phase)">{{ pod.phase }}</td>
                <td class="px-6 py-3 text-right tabular-nums">{{ pod.ready }}</td>
                <td class="px-6 py-3 text-right tabular-nums">{{ pod.restarts }}</td>
                <td class="px-6 py-3" :class="pod.reason ? 'text-red-700' : 'text-theme-textLight'">{{ pod.reason || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="section overflow-hidden p-0" aria-labelledby="events-title">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 id="events-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.cluster_events') }}</h2>
          <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.cluster_events_desc') }}</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-xs font-medium text-theme-textLight">
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_type') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_reason') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_object') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_message') }}</th>
                <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_count') }}</th>
                <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_time') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-theme-border">
              <tr v-for="event in visibleEvents" :key="event.object" class="text-theme-text">
                <td class="px-6 py-3" :class="event.type === 'Warning' ? 'font-semibold text-red-700' : 'text-theme-textLight'">{{ event.type }}</td>
                <td class="px-6 py-3 font-medium">{{ event.reason }}</td>
                <td class="px-6 py-3 text-theme-textLight">{{ event.object }}</td>
                <td class="px-6 py-3 text-theme-textLight">{{ event.message }}</td>
                <td class="px-6 py-3 text-right tabular-nums">{{ event.count }}</td>
                <td class="px-6 py-3 whitespace-nowrap text-theme-textLight">{{ eventTime(event.lastSeen) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <section v-if="isCluster" class="section overflow-hidden p-0" aria-labelledby="replicas-title">
      <div class="border-b border-theme-border px-6 py-4">
        <h2 id="replicas-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.cluster_replicas') }}</h2>
        <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.cluster_replicas_desc') }}</p>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-xs font-medium text-theme-textLight">
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_namespace') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_workload') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_kind') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_ready') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_desired') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_image') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-theme-border">
            <tr v-for="workload in visibleWorkloads" :key="workload.name" class="text-theme-text">
              <td class="px-6 py-3 text-theme-textLight">{{ workload.namespace }}</td>
              <td class="px-6 py-3 font-medium">{{ workload.name }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ workload.kind }}</td>
              <td class="px-6 py-3 text-right tabular-nums" :class="workload.ready < workload.desired ? 'font-semibold text-red-700' : ''">{{ workload.ready }}</td>
              <td class="px-6 py-3 text-right tabular-nums">{{ workload.desired }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ workload.image }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <template v-if="isCluster">
      <p class="text-sm text-theme-textLight">
        {{ $t('dashboard.cluster_version', { version: fakeClusterResources.cluster.serverVersion }) }} ·
        {{ $t('dashboard.cluster_collected', { time: eventTime(fakeClusterResources.cluster.collectedAt) }) }}
      </p>

      <div class="flex flex-wrap items-center gap-3">
        <label for="namespace-filter" class="text-sm font-medium text-theme-text">{{ $t('dashboard.namespace_filter') }}</label>
        <select
          id="namespace-filter"
          v-model="selectedNamespace"
          class="input-field w-auto min-w-[12rem]"
        >
          <option value="all">{{ $t('dashboard.all_namespaces') }}</option>
          <option v-for="name in namespaceOptions" :key="name" :value="name">{{ name }}</option>
        </select>
        <span class="text-xs text-theme-textLight">{{ $t('dashboard.namespace_scope_note') }}</span>
      </div>

      <section class="section overflow-hidden p-0" aria-labelledby="errors-title">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 id="errors-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.resources.errors.title') }}</h2>
        </div>
        <ul v-if="collectionErrors.length" class="divide-y divide-theme-border">
          <li v-for="item in collectionErrors" :key="item.source" class="px-6 py-3 text-sm">
            <span class="font-medium text-theme-text">{{ item.source }}</span>
            <span class="ml-2 text-red-700">{{ item.message }}</span>
          </li>
        </ul>
        <p v-else class="px-6 py-6 text-sm text-theme-textLight">{{ $t('dashboard.resources.errors.none') }}</p>
      </section>

      <ClusterResourceTable
        v-for="section in resourceSections"
        :key="section.source"
        :title="section.title"
        :description="section.description"
        :columns="section.columns"
        :rows="section.rows"
        :error="section.error"
      />
    </template>

    <section v-if="!clusterId" class="section overflow-hidden p-0" aria-labelledby="clusters-title">
      <div class="border-b border-theme-border px-6 py-4">
        <h2 id="clusters-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.clusters') }}</h2>
        <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.clusters_desc') }}</p>
      </div>
      <div class="grid grid-cols-1 gap-4 p-6 sm:grid-cols-2 xl:grid-cols-5">
        <div
          v-for="cluster in clusters"
          :key="cluster.id"
          class="flex flex-col gap-3 rounded-xl border bg-theme-surface p-4 text-left transition-colors hover:border-primary-300"
          :class="selectedId === cluster.id ? 'border-theme-text ring-1 ring-theme-text' : 'border-theme-border'"
        >
          <div class="flex items-center justify-between gap-2">
            <button
              type="button"
              class="truncate font-semibold text-theme-text hover:underline"
              :aria-pressed="selectedId === cluster.id"
              @click="toggleCluster(cluster.id)"
              :title="cluster.id"
            >
              {{ cluster.id }}
            </button>
            <span class="inline-flex items-center gap-1.5 whitespace-nowrap text-xs text-theme-textLight">
              <i class="h-1.5 w-1.5 rounded-full" :class="statusDot[cluster.state]" />
              {{ clusterStatus(cluster) }}
            </span>
          </div>
          <span class="text-xs text-theme-textLight">{{ cluster.region }}</span>
          <div>
            <div class="mb-1 flex justify-between text-xs text-theme-textLight">
              <span>{{ $t('dashboard.cpu') }}</span><span class="tabular-nums">{{ cluster.cpu }}%</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-theme-hover">
              <div class="h-full rounded-full" :class="meterClass(cluster.cpu)" :style="{ width: cluster.cpu + '%' }" />
            </div>
          </div>
          <div>
            <div class="mb-1 flex justify-between text-xs text-theme-textLight">
              <span>{{ $t('dashboard.memory') }}</span><span class="tabular-nums">{{ cluster.mem }}%</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-theme-hover">
              <div class="h-full rounded-full" :class="meterClass(cluster.mem)" :style="{ width: cluster.mem + '%' }" />
            </div>
          </div>
          <router-link :to="`/clusters/${cluster.id}`" class="text-sm font-medium text-primary-800 hover:underline">
            {{ $t('clusters_page.dashboard') }}
          </router-link>
        </div>
      </div>
    </section>

    <section class="section overflow-hidden p-0" aria-labelledby="optimization-title">
      <div class="border-b border-theme-border px-6 py-4">
        <h2 id="optimization-title" class="text-lg font-semibold text-theme-text">{{ $t('dashboard.optimization') }}</h2>
        <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.optimization_desc') }}</p>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-xs font-medium text-theme-textLight">
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_workload') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_cluster') }}</th>
              <th class="px-6 py-3 font-medium">{{ $t('dashboard.col_change') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_p95') }}</th>
              <th class="px-6 py-3 text-right font-medium">{{ $t('dashboard.col_savings') }}</th>
              <th class="px-6 py-3"><span class="sr-only">{{ $t('dashboard.col_action') }}</span></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-theme-border">
            <tr v-for="rec in visibleRecs" :key="rec.id" :class="appliedIds.includes(rec.id) ? 'text-theme-textLight' : 'text-theme-text'">
              <td class="px-6 py-3 font-medium">{{ rec.workload }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ rec.cluster }}</td>
              <td class="px-6 py-3 text-theme-textLight">{{ rec.change }}</td>
              <td class="px-6 py-3 text-right tabular-nums text-theme-textLight">{{ rec.p95 }}</td>
              <td class="px-6 py-3 text-right tabular-nums font-semibold text-green-700">
                <span v-if="appliedIds.includes(rec.id)" class="text-xs font-medium">{{ $t('dashboard.applied') }}</span>
                <span v-else>${{ rec.savings }}</span>
              </td>
              <td class="px-6 py-3 text-right">
                <button
                  v-if="!appliedIds.includes(rec.id)"
                  type="button"
                  class="btn-secondary"
                  @click="applyRec(rec)"
                >
                  {{ $t('dashboard.apply') }}
                </button>
              </td>
            </tr>
            <tr v-if="!visibleRecs.length">
              <td colspan="6" class="px-6 py-6 text-center text-theme-textLight">{{ $t('dashboard.no_recommendations') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div
      class="fixed bottom-4 right-4 max-w-sm rounded-md bg-theme-text px-4 py-2.5 text-sm text-theme-bg shadow-lg transition-all duration-200"
      :class="toastMessage ? 'opacity-100' : 'pointer-events-none opacity-0'"
      role="status"
      aria-live="polite"
    >
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  anomalies,
  recommendations,
  cpuSeries,
  memSeries,
  deployIndex,
  allClusters,
  fakeClusterDetail,
} from '@/lib/clusters'
import { buildResourceGroups, clusterScopedSources, fakeClusterResources } from '@/lib/clusterResources'
import ClusterResourceTable from '@/components/ClusterResourceTable.vue'

const props = defineProps({
  clusterId: { type: String, default: null },
  snapshot: { type: Object, default: null },
})

const { t } = useI18n()

const ranges = ['1h', '24h', '7d']
const activeRange = ref('24h')
const selectedId = ref(props.clusterId)
const appliedIds = ref([])
const dismissedIds = ref([])
const toastMessage = ref('')

let toastTimer = null

// Placeholder cluster state until the live collect feeds the board
const detail = fakeClusterDetail

// Namespace filter for the namespaced tables. Cluster-wide cards and nodes ignore it.
const selectedNamespace = ref('all')
const namespaceOptions = fakeClusterResources.namespaces.map((namespace) => namespace.name)

function inNamespace(row) {
  return selectedNamespace.value === 'all' || row.namespace === selectedNamespace.value
}

const visiblePods = computed(() => detail.pods.filter(inNamespace))
const visibleEvents = computed(() => detail.events.filter(inNamespace))
const visibleWorkloads = computed(() => detail.workloads.filter(inNamespace))

const resourceSections = computed(() =>
  buildResourceGroups(t).map((group) => {
    const scoped = !clusterScopedSources.has(group.source)
    const rows = fakeClusterResources[group.source]

    return {
      ...group,
      rows: scoped ? rows.filter(inNamespace) : rows,
      error: fakeClusterResources.errors[group.source] ?? null,
      emptyText: t(selectedNamespace.value === 'all' ? 'dashboard.resources.empty' : 'dashboard.resources.empty_namespace'),
    }
  }),
)
const collectionErrors = computed(() =>
  Object.entries(fakeClusterResources.errors).map(([source, message]) => ({ source, message })),
)

const isCluster = computed(() => Boolean(props.clusterId))
const clusters = computed(() => props.snapshot?.clusters ?? allClusters())
const boardAnomalies = computed(() => props.snapshot?.anomalies ?? anomalies)
const boardRecs = computed(() => props.snapshot?.recommendations ?? recommendations)
const activeScope = computed(() => props.clusterId ?? selectedId.value)
const scopeClusters = computed(() =>
  activeScope.value ? clusters.value.filter((c) => c.id === activeScope.value) : clusters.value,
)

const severityClass = { critical: 'bg-red-600', warning: 'bg-amber-500', info: 'bg-theme-textLight' }
const statusDot = { ok: 'bg-green-600', warn: 'bg-amber-500', crit: 'bg-red-600' }
const plotLeft = 36
const plotRight = 548
const plotTop = 14
const plotBottom = 206

// Fleet demo series for the global board's trend chart
const staticLabels = [0, 6, 12, 18, 23].map((hour) => ({ index: hour, text: String(hour).padStart(2, '0') + ':00' }))

const trend = computed(() => ({
  cpu: cpuSeries,
  memory: memSeries,
  lo: 30,
  hi: 100,
  ticks: [40, 60, 80, 100],
  labels: staticLabels,
}))

const healthyCount = computed(() => scopeClusters.value.filter((c) => c.state === 'ok').length)
const openAnomalies = computed(() =>
  boardAnomalies.value.filter(
    (a) => !dismissedIds.value.includes(a.id) && (!activeScope.value || a.cluster === activeScope.value),
  ),
)
const criticalCount = computed(() => openAnomalies.value.filter((a) => a.severity === 'critical').length)
const visibleAnomalies = computed(() => openAnomalies.value)
const pendingRecs = computed(() =>
  boardRecs.value.filter(
    (r) => !appliedIds.value.includes(r.id) && (!activeScope.value || r.cluster === activeScope.value),
  ),
)
const pendingSavings = computed(() => pendingRecs.value.reduce((sum, r) => sum + r.savings, 0))
const visibleRecs = computed(() => pendingRecs.value)

const deployX = xFor(deployIndex, cpuSeries.length)
const axisLabels = computed(() =>
  trend.value.labels.map((label) => ({
    key: label.index,
    x: xFor(label.index, trend.value.cpu.length),
    text: label.text,
  })),
)

const trendCaption = computed(() => t(`dashboard.trend_${activeRange.value}`))

function xFor(index, count) {
  return plotLeft + (index * (plotRight - plotLeft)) / Math.max(count - 1, 1)
}

function yFor(value) {
  const { lo, hi } = trend.value

  return plotBottom - ((value - lo) / (hi - lo)) * (plotBottom - plotTop)
}

function linePath(values) {
  return values
    .map((value, i) => `${i ? 'L' : 'M'}${xFor(i, values.length).toFixed(1)} ${yFor(value).toFixed(1)}`)
    .join(' ')
}

const cpuPath = computed(() => linePath(trend.value.cpu))
const memoryPath = computed(() => linePath(trend.value.memory))

function flagText(value) {
  return value ? t('dashboard.flag_yes') : t('dashboard.flag_no')
}

// Highlights a condition that is bad when it is true
function flagClass(bad) {
  return bad ? 'font-semibold text-red-700' : 'text-theme-textLight'
}

function phaseClass(phase) {
  if (phase === 'Running') return 'text-theme-text'
  if (phase === 'Failed') return 'font-semibold text-red-700'

  return 'font-semibold text-amber-700'
}

function eventTime(iso) {
  return new Date(iso).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }) + ' UTC'
}

function clusterStatus(cluster) {
  return cluster.open ? t('dashboard.open_count', { count: cluster.open }) : t('dashboard.healthy')
}

function meterClass(value) {
  if (value > 85) return 'bg-red-600'
  if (value > 70) return 'bg-amber-500'
  return 'bg-theme-text'
}

function toast(message) {
  toastMessage.value = message
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 2600)
}

function toggleCluster(id) {
  if (props.clusterId) return
  selectedId.value = selectedId.value === id ? null : id
}

function runAction(anomaly, action) {
  if (action.dismiss) {
    dismissedIds.value = [...dismissedIds.value, anomaly.id]
    toast(t('dashboard.toast_dismissed', { title: anomaly.title }))
    return
  }
  toast(action.message)
}

function applyRec(rec) {
  appliedIds.value = [...appliedIds.value, rec.id]
  toast(t('dashboard.toast_applied', { workload: rec.workload, amount: rec.savings }))
}
</script>
