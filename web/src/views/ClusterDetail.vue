<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <div v-if="cluster">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ cluster.name }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ cluster.description || '—' }} · {{ clusterStatus(openCount) }}
        </p>
      </div>
      <div v-else-if="!loading">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('clusters_page.not_found') }}
        </h1>
      </div>

      <p v-if="loading" class="text-sm text-theme-textLight">{{ $t('common.loading') }}</p>
      <ClusterBoard v-if="cluster" :cluster-id="cluster.id" :snapshot="snapshot" />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import ClusterBoard from '@/components/ClusterBoard.vue'
import { cluster_api } from '@/api'
import { fetchClusterState, toBoardSnapshot } from '@/lib/clusters'
import { loadWorkspaceFromStorage } from '@/utils/storage'

const { t } = useI18n()
const route = useRoute()

const cluster = ref(null)
const workspaceId = computed(() => loadWorkspaceFromStorage()?.id)
const snapshot = ref(null)
const loading = ref(false)

// Live open-anomaly count when a snapshot arrived, otherwise zero.
const openCount = computed(() => (snapshot.value ? snapshot.value.anomalies.length : 0))

function clusterStatus(open) {
  return open ? t('dashboard.open_count', { count: open }) : t('dashboard.healthy')
}

onMounted(async () => {
  loading.value = true

  try {
    const { data } = await cluster_api.get(workspaceId.value, route.params.clusterId)
    cluster.value = data
  } catch (_) {
    cluster.value = null
  }

  if (cluster.value) {
    try {
      const state = await fetchClusterState(workspaceId.value, cluster.value.id)
      snapshot.value = toBoardSnapshot(state, cluster.value.id)
    } catch (_) {
      // The collect has not run yet or failed: the board falls back to its static data.
      snapshot.value = null
    }
  }

  loading.value = false
})
</script>
