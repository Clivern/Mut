<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('clusters_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('clusters_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div
          class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4"
        >
          <div>
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('clusters_page.list_title') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('clusters_page.list_desc') }}</p>
          </div>
          <button
            type="button"
            class="btn-primary inline-flex items-center justify-center gap-2"
            @click="openCreateModal"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            {{ $t('clusters_page.new_cluster') }}
          </button>
        </div>

        <div
          v-if="errorMessage"
          class="mx-6 mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ errorMessage }}
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('clusters_page.name') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('clusters_page.description') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('clusters_page.created') }}
                </th>
                <th
                  class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('clusters_page.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="!clusters.length">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('clusters_page.no_clusters') }}
                </td>
              </tr>
              <tr v-for="cluster in clusters" :key="cluster.id" class="hover:bg-theme-hover">
                <td class="px-6 py-4">
                  <router-link
                    :to="`/clusters/${cluster.id}`"
                    class="text-sm font-medium text-theme-text hover:underline"
                  >
                    {{ cluster.name }}
                  </router-link>
                  <p class="mt-0.5 font-mono text-xs text-theme-textLight">{{ cluster.id }}</p>
                </td>
                <td class="px-6 py-4 text-sm text-theme-text">
                  {{ cluster.description || '—' }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                  {{ cluster.createdAt ? formatCreatedAt(cluster.createdAt, locale) : '—' }}
                </td>
                <td class="px-6 py-4 text-right whitespace-nowrap">
                  <router-link
                    v-if="cluster.stateCollectedAt"
                    :to="`/clusters/${cluster.id}`"
                    class="mr-3 text-sm font-medium text-primary-600 hover:text-primary-700 hover:underline"
                  >
                    {{ $t('clusters_page.dashboard') }}
                  </router-link>
                  <button
                    v-else
                    type="button"
                    class="mr-3 text-sm font-medium text-primary-600 hover:text-primary-700 hover:underline"
                    @click="openSetupModal(cluster)"
                  >
                    {{ $t('clusters_page.setup_button') }}
                  </button>
                  <button
                    type="button"
                    class="mr-3 text-sm font-medium text-theme-text hover:underline"
                    @click="openEditModal(cluster)"
                  >
                    {{ $t('clusters_page.edit') }}
                  </button>
                  <button
                    type="button"
                    class="text-sm text-red-600 hover:text-red-700 font-medium"
                    @click="removeModalCluster = cluster"
                  >
                    {{ $t('clusters_page.remove') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>

    <BaseModal
      :open="showCreateModal"
      :title="$t('clusters_page.new_cluster')"
      :description="$t('clusters_page.new_cluster_desc')"
      size="sm"
      @close="closeCreateModal"
    >
      <form @submit.prevent="handleCreate" class="mt-4 space-y-4">
        <div
          v-if="createError"
          class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ createError }}
        </div>
        <div>
          <label for="create-cluster-name" class="form-label">{{ $t('clusters_page.name') }}</label>
          <input
            id="create-cluster-name"
            v-model="createName"
            type="text"
            required
            minlength="2"
            maxlength="100"
            class="input-field"
            :placeholder="$t('clusters_page.name_placeholder')"
            autofocus
          />
        </div>
        <div>
          <label for="create-cluster-description" class="form-label">
            {{ $t('clusters_page.description') }}
          </label>
          <textarea
            id="create-cluster-description"
            v-model="createDescription"
            required
            maxlength="500"
            rows="3"
            class="input-field"
            :placeholder="$t('clusters_page.description_placeholder')"
          />
        </div>
        <div class="flex gap-3 justify-end">
          <button type="button" @click="closeCreateModal" class="btn-secondary">
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" class="btn-primary" :disabled="creating">
            {{ $t('common.create') }}
          </button>
        </div>
      </form>
    </BaseModal>

    <BaseModal
      :open="showEditModal"
      :title="$t('clusters_page.edit_cluster')"
      size="sm"
      @close="closeEditModal"
    >
      <form @submit.prevent="handleUpdate" class="mt-4 space-y-4">
        <div
          v-if="editError"
          class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ editError }}
        </div>
        <div>
          <label for="edit-cluster-name" class="form-label">{{ $t('clusters_page.name') }}</label>
          <input
            id="edit-cluster-name"
            v-model="editName"
            type="text"
            required
            minlength="2"
            maxlength="100"
            class="input-field"
            :placeholder="$t('clusters_page.name_placeholder')"
            autofocus
          />
        </div>
        <div>
          <label for="edit-cluster-description" class="form-label">
            {{ $t('clusters_page.description') }}
          </label>
          <textarea
            id="edit-cluster-description"
            v-model="editDescription"
            required
            maxlength="500"
            rows="3"
            class="input-field"
            :placeholder="$t('clusters_page.description_placeholder')"
          />
        </div>
        <div class="flex gap-3 justify-end">
          <button type="button" @click="closeEditModal" class="btn-secondary">
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" class="btn-primary" :disabled="updating">
            {{ $t('common.save') }}
          </button>
        </div>
      </form>
    </BaseModal>

    <BaseModal
      :open="showSetupModal"
      :title="$t('clusters_page.setup_title')"
      :description="$t('clusters_page.setup_desc', { name: setupCluster?.name })"
      size="xl"
      @close="closeSetupModal"
    >
      <div class="mt-4 space-y-5 text-sm text-theme-text">
        <div>
          <p class="font-medium">{{ $t('clusters_page.setup_step_key') }}</p>
          <p class="mt-1 text-theme-textLight">{{ $t('clusters_page.setup_key_hint') }}</p>
          <div v-if="accessKey" class="mt-2 space-y-2">
            <p class="text-theme-text">{{ $t('clusters_page.setup_key_created') }}</p>
          </div>
          <div v-else class="mt-2">
            <button
              type="button"
              class="btn-primary"
              :disabled="creatingKey"
              @click="createAccessKey"
            >
              {{ creatingKey ? $t('clusters_page.setup_creating_key') : $t('clusters_page.setup_create_key') }}
            </button>
          </div>
          <p v-if="keyError" class="mt-2 text-sm text-red-700">{{ keyError }}</p>
        </div>

        <div>
          <p class="font-medium">{{ $t('clusters_page.setup_step_run') }}</p>
          <div class="mt-2 flex items-start gap-2">
            <pre
              class="flex-1 overflow-x-auto rounded-lg bg-theme-hover p-3 font-mono text-xs text-theme-text whitespace-pre-wrap break-all"
            >{{ installCommand }}</pre>
            <button type="button" class="btn-secondary shrink-0" @click="copyInstallCommand">
              {{ copied ? $t('clusters_page.copied') : $t('clusters_page.copy') }}
            </button>
          </div>
          <p class="mt-2 text-theme-textLight">{{ $t('clusters_page.setup_wait') }}</p>
        </div>

        <div class="flex justify-end">
          <button type="button" class="btn-secondary" @click="closeSetupModal">
            {{ $t('common.close') }}
          </button>
        </div>
      </div>
    </BaseModal>

    <BaseModal
      :open="!!removeModalCluster"
      :title="$t('clusters_page.remove_title')"
      :description="$t('clusters_page.remove_confirm', { name: removeModalCluster?.name })"
      size="xl"
      @close="removeModalCluster = null"
    >
      <div class="mt-4 space-y-5 text-sm text-theme-text">
        <ul class="list-disc space-y-2 pl-5 text-theme-textLight">
          <li>{{ $t('clusters_page.remove_tip_record') }}</li>
          <li>{{ $t('clusters_page.remove_tip_agent') }}</li>
        </ul>

        <div>
          <p class="font-medium">{{ $t('clusters_page.remove_step_uninstall') }}</p>
          <div class="mt-2 flex items-start gap-2">
            <pre
              class="flex-1 overflow-x-auto rounded-lg bg-theme-hover p-3 font-mono text-xs text-theme-text whitespace-pre-wrap break-all"
            >{{ uninstallCommand }}</pre>
            <button type="button" class="btn-secondary shrink-0" @click="copyUninstallCommand">
              {{ copied ? $t('clusters_page.copied') : $t('clusters_page.copy') }}
            </button>
          </div>
        </div>

        <div class="flex gap-3 justify-end">
          <button type="button" class="btn-secondary" :disabled="removing" @click="removeModalCluster = null">
            {{ $t('common.cancel') }}
          </button>
          <button type="button" class="btn-danger" :disabled="removing" @click="confirmRemove">
            {{ removing ? $t('clusters_page.removing') : $t('clusters_page.remove') }}
          </button>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import BaseModal from '@/components/BaseModal.vue'
import { cluster_api, workspace_access_keys_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'
import { formatCreatedAt } from '@/utils/date'
import { loadWorkspaceFromStorage } from '@/utils/storage'

const { t, locale } = useI18n()

const workspaceId = computed(() => loadWorkspaceFromStorage()?.id)
const clusters = ref([])
const loading = ref(false)
const errorMessage = ref(null)
const showCreateModal = ref(false)
const createName = ref('')
const createDescription = ref('')
const createError = ref(null)
const creating = ref(false)
const showEditModal = ref(false)
const editingCluster = ref(null)
const editName = ref('')
const editDescription = ref('')
const editError = ref(null)
const updating = ref(false)
const removeModalCluster = ref(null)
const removing = ref(false)
const showSetupModal = ref(false)
const setupCluster = ref(null)
const copied = ref(false)
const accessKey = ref(null)
const creatingKey = ref(false)
const keyError = ref(null)

const INSTALL_SCRIPT_URL = 'https://raw.githubusercontent.com/Clivern/Mut/main/bin/install-agent.sh'
const UNINSTALL_SCRIPT_URL = 'https://raw.githubusercontent.com/Clivern/Mut/main/bin/uninstall-agent.sh'

const uninstallCommand = `curl -fsSL ${UNINSTALL_SCRIPT_URL} | bash -s -- --yes`

const installCommand = computed(() =>
  [
    `curl -fsSL ${INSTALL_SCRIPT_URL} | bash -s --`,
    `  --server ${window.location.origin}`,
    `  --workspace ${workspaceId.value}`,
    `  --cluster ${setupCluster.value?.id}`,
    `  --access-key ${accessKey.value ?? '<ACCESS_KEY>'}`,
  ].join(' \\\n'),
)

function openSetupModal(cluster) {
  setupCluster.value = cluster
  accessKey.value = null
  keyError.value = null
  copied.value = false
  showSetupModal.value = true
}

// Reloads the list so a cluster that has received its first push switches to Dashboard.
function closeSetupModal() {
  showSetupModal.value = false
  setupCluster.value = null
  accessKey.value = null
  loadClusters()
}

// Creates a clusters:write access key for this install. The raw key is returned once.
async function createAccessKey() {
  creatingKey.value = true
  keyError.value = null

  try {
    const { data } = await workspace_access_keys_api.create(workspaceId.value, {
      name: `agent-${setupCluster.value.id}`,
      permissions: ['clusters:write'],
    })
    accessKey.value = data.key
  } catch (err) {
    keyError.value = errorText(err, t('clusters_page.failed_create_key'))
  } finally {
    creatingKey.value = false
  }
}

async function copyUninstallCommand() {
  try {
    await navigator.clipboard.writeText(uninstallCommand)
    copied.value = true
  } catch (_) {
    copied.value = false
  }
}

async function copyInstallCommand() {
  try {
    await navigator.clipboard.writeText(installCommand.value)
    copied.value = true
  } catch (_) {
    copied.value = false
  }
}

async function loadClusters() {
  loading.value = true
  errorMessage.value = null

  try {
    const { data } = await cluster_api.list(workspaceId.value, { limit: 100, offset: 0 })
    clusters.value = data.clusters
  } catch (err) {
    errorMessage.value = errorText(err, t('clusters_page.failed_load_clusters'))
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  createName.value = ''
  createDescription.value = ''
  createError.value = null
  showCreateModal.value = true
}

function closeCreateModal() {
  createError.value = null
  showCreateModal.value = false
}

async function handleCreate() {
  creating.value = true
  createError.value = null

  try {
    const { data } = await cluster_api.create(workspaceId.value, {
      name: createName.value.trim(),
      description: createDescription.value.trim(),
    })
    clusters.value = [data, ...clusters.value]
    showFlash(t('clusters_page.cluster_created'))
    closeCreateModal()
  } catch (err) {
    createError.value = errorText(err, t('clusters_page.failed_save_cluster'))
  } finally {
    creating.value = false
  }
}

function openEditModal(cluster) {
  editingCluster.value = cluster
  editName.value = cluster.name
  editDescription.value = cluster.description || ''
  editError.value = null
  showEditModal.value = true
}

function closeEditModal() {
  editingCluster.value = null
  editError.value = null
  showEditModal.value = false
}

async function handleUpdate() {
  updating.value = true
  editError.value = null

  try {
    const { data } = await cluster_api.update(workspaceId.value, editingCluster.value.id, {
      name: editName.value.trim(),
      description: editDescription.value.trim(),
    })
    clusters.value = clusters.value.map((cluster) => (cluster.id === data.id ? data : cluster))
    showFlash(t('clusters_page.cluster_saved'))
    closeEditModal()
  } catch (err) {
    editError.value = errorText(err, t('clusters_page.failed_save_cluster'))
  } finally {
    updating.value = false
  }
}

async function confirmRemove() {
  const cluster = removeModalCluster.value

  removing.value = true
  copied.value = false

  try {
    await cluster_api.remove(workspaceId.value, cluster.id)
    clusters.value = clusters.value.filter((item) => item.id !== cluster.id)
    showFlash(t('clusters_page.cluster_removed'))
  } catch (err) {
    errorMessage.value = errorText(err, t('clusters_page.failed_remove_cluster'))
  } finally {
    removeModalCluster.value = null
    removing.value = false
  }
}

onMounted(loadClusters)
</script>
