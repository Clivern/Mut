<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8 space-y-8">
      <div>
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('workspace_settings_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('workspace_settings_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-lg font-semibold text-theme-text">
            {{ $t('workspace_settings_page.general') }}
          </h2>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ $t('workspace_settings_page.general_desc') }}
          </p>
        </div>

        <form class="flex flex-wrap items-end gap-3 p-6" @submit.prevent="handleSave">
          <div class="min-w-0 flex-1" style="min-width: 200px;">
            <label for="workspace-name" class="block text-sm font-medium text-theme-text mb-1.5">{{ $t('workspace_settings_page.name') }}</label>
            <input
              id="workspace-name"
              v-model="workspaceName"
              type="text"
              required
              minlength="2"
              maxlength="60"
              class="input-field"
              :disabled="saving"
            >
          </div>
          <button type="submit" class="btn-primary shrink-0" :disabled="saving">
            {{ saving ? $t('workspace_settings_page.saving') : $t('workspace_settings_page.save') }}
          </button>
        </form>
      </section>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-lg font-semibold text-theme-text">
            {{ $t('workspace_settings_page.identifiers') }}
          </h2>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ $t('workspace_settings_page.identifiers_desc') }}
          </p>
        </div>

        <div class="p-6">
          <label class="form-label">{{ $t('workspace_settings_page.id') }}</label>
          <div class="flex items-center gap-2">
            <input
              :value="workspace?.id ?? ''"
              type="text"
              readonly
              class="input-field bg-theme-bg text-theme-textLight font-mono text-sm"
            >
            <button
              type="button"
              class="btn-secondary shrink-0"
              @click="copyText(workspace?.id, 'id')"
            >
              {{
                copiedField === 'id'
                  ? $t('workspace_settings_page.copied')
                  : $t('workspace_settings_page.copy')
              }}
            </button>
          </div>
          <p v-if="copyError" class="mt-2 text-sm text-red-600">{{ copyError }}</p>
        </div>
      </section>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-theme-border">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('workspace_settings_page.access_keys') }}</h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('workspace_settings_page.access_keys_desc') }}</p>
        </div>

        <div class="p-6 space-y-6">
          <form @submit.prevent="handleCreateKey" class="space-y-4">
            <div class="flex flex-wrap items-end gap-3">
              <div class="min-w-0 flex-1" style="min-width: 200px;">
                <label for="key-name" class="block text-sm font-medium text-theme-text mb-1.5">{{ $t('workspace_settings_page.key_name') }}</label>
                <input
                  id="key-name"
                  v-model="keyName"
                  type="text"
                  required
                  minlength="2"
                  maxlength="60"
                  class="input-field"
                  :disabled="creatingKey"
                >
              </div>
              <div class="min-w-0 flex-1" style="min-width: 200px;">
                <label for="key-expires" class="block text-sm font-medium text-theme-text mb-1.5">{{ $t('workspace_settings_page.key_expires') }}</label>
                <input
                  id="key-expires"
                  v-model="keyExpiresAt"
                  type="datetime-local"
                  class="input-field"
                  :disabled="creatingKey"
                >
              </div>
              <button type="submit" class="btn-primary shrink-0" :disabled="creatingKey">
                {{ creatingKey ? $t('workspace_settings_page.creating') : $t('workspace_settings_page.create_key') }}
              </button>
            </div>

            <fieldset class="space-y-2">
              <legend class="block text-sm font-medium text-theme-text mb-1.5">{{ $t('workspace_settings_page.key_permissions') }}</legend>
              <label v-for="permission in KEY_PERMISSIONS" :key="permission" class="flex items-center gap-2 text-sm text-theme-text">
                <input v-model="keyPermissions" type="checkbox" :value="permission" :disabled="creatingKey">
                {{ $t(`workspace_settings_page.permission_${permission.replace(':', '_')}`) }}
              </label>
            </fieldset>

            <div
              v-if="keyError"
              class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
            >
              {{ keyError }}
            </div>
          </form>

          <div v-if="newlyCreatedKey" class="rounded-lg border border-amber-200 bg-amber-50 p-4 space-y-2">
            <p class="text-sm font-medium text-amber-900">{{ $t('workspace_settings_page.copy_key_now') }}</p>
            <code class="block px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono">{{ newlyCreatedKey }}</code>
            <button type="button" class="text-sm text-amber-800 hover:text-amber-900 font-medium" @click="newlyCreatedKey = null">
              {{ $t('workspace_settings_page.dismiss') }}
            </button>
          </div>

          <div v-if="keysLoading" class="text-sm text-theme-textLight py-4">
            {{ $t('workspace_settings_page.loading_keys') }}
          </div>
          <div v-else-if="accessKeys.length === 0" class="text-sm text-theme-textLight py-4">
            {{ $t('workspace_settings_page.no_keys') }}
          </div>
          <div v-else class="border border-theme-border rounded-lg overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-bg">
                <tr>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('workspace_settings_page.key_name') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('workspace_settings_page.masked_key') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('workspace_settings_page.key_permissions') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('workspace_settings_page.expires') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('workspace_settings_page.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="key in accessKeys" :key="key.id" class="hover:bg-theme-hover/50">
                  <td class="px-4 py-3 text-sm text-theme-text">{{ key.name }}</td>
                  <td class="px-4 py-3 text-sm font-mono text-theme-textLight">{{ key.maskedKey }}</td>
                  <td class="px-4 py-3 text-sm text-theme-textLight">{{ key.permissions.join(', ') }}</td>
                  <td class="px-4 py-3 text-sm text-theme-textLight">
                    {{ key.expiresAt ? formatCreatedAt(key.expiresAt, locale) : $t('workspace_settings_page.never') }}
                  </td>
                  <td class="px-4 py-3 text-right">
                    <button type="button" class="text-sm text-red-600 hover:text-red-700 font-medium" @click="revokeModalKey = key">
                      {{ $t('workspace_settings_page.revoke') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </main>

    <ConfirmModal
      :open="!!revokeModalKey"
      :title="$t('workspace_settings_page.revoke_title')"
      :message="$t('workspace_settings_page.revoke_confirm', { name: revokeModalKey?.name })"
      :busy="revokingKey"
      :confirm-label="$t('workspace_settings_page.revoke')"
      :busy-label="$t('workspace_settings_page.revoking')"
      @cancel="revokeModalKey = null"
      @confirm="confirmRevokeKey"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'
import { workspace_api, workspace_access_keys_api } from '@/api'
import { formatCreatedAt } from '@/utils/date'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage } from '@/utils/storage'

const KEY_PERMISSIONS = ['workspace:read', 'members:read', 'clusters:read', 'clusters:write']

const { t, locale } = useI18n()

const workspaceId = computed(() => loadWorkspaceFromStorage()?.id)
const workspace = ref(null)
const workspaceName = ref('')
const saving = ref(false)
const errorMessage = ref(null)

const copiedField = ref(null)
const copyError = ref(null)

const accessKeys = ref([])
const keysLoading = ref(false)
const keyName = ref('')
const keyExpiresAt = ref('')
const keyPermissions = ref([])
const keyError = ref(null)
const creatingKey = ref(false)
const newlyCreatedKey = ref(null)
const revokeModalKey = ref(null)
const revokingKey = ref(false)

async function loadWorkspace() {
  const res = await workspace_api.get(workspaceId.value)
  workspace.value = res.data
  workspaceName.value = res.data.name
}

async function loadAccessKeys() {
  keysLoading.value = true
  try {
    const res = await workspace_access_keys_api.list(workspaceId.value, { limit: 100, offset: 0 })
    accessKeys.value = res.data.accessKeys ?? []
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_load_keys'))
  } finally {
    keysLoading.value = false
  }
}

async function loadAll() {
  errorMessage.value = null

  try {
    await loadWorkspace()
    await loadAccessKeys()
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_load_workspace'))
  }
}

async function handleSave() {
  saving.value = true
  errorMessage.value = null

  try {
    const res = await workspace_api.update(workspaceId.value, { name: workspaceName.value.trim() })
    workspace.value = res.data
    workspaceName.value = res.data.name
    saveWorkspaceToStorage({ ...loadWorkspaceFromStorage(), ...res.data })
    showFlash(t('workspace_settings_page.workspace_saved'))
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_save_workspace'))
  } finally {
    saving.value = false
  }
}

async function copyText(value, field) {
  copyError.value = null

  try {
    await navigator.clipboard.writeText(value ?? '')
    copiedField.value = field
    setTimeout(() => {
      if (copiedField.value === field) copiedField.value = null
    }, 2000)
  } catch (err) {
    copyError.value = t('workspace_settings_page.copy_failed')
  }
}

async function handleCreateKey() {
  keyError.value = null

  if (keyPermissions.value.length === 0) {
    keyError.value = t('workspace_settings_page.permission_required')
    return
  }

  creatingKey.value = true
  newlyCreatedKey.value = null

  try {
    const payload = {
      name: keyName.value.trim(),
      permissions: keyPermissions.value,
    }
    if (keyExpiresAt.value) {
      payload.expiresAt = new Date(keyExpiresAt.value).toISOString()
    }

    const res = await workspace_access_keys_api.create(workspaceId.value, payload)
    newlyCreatedKey.value = res.data.key
    keyName.value = ''
    keyExpiresAt.value = ''
    keyPermissions.value = []
    await loadAccessKeys()
  } catch (err) {
    keyError.value = errorText(err, t('workspace_settings_page.failed_create_key'))
  } finally {
    creatingKey.value = false
  }
}

async function confirmRevokeKey() {
  const key = revokeModalKey.value
  if (!key) return

  revokingKey.value = true
  keyError.value = null

  try {
    await workspace_access_keys_api.delete(workspaceId.value, key.id)
    revokeModalKey.value = null
    await loadAccessKeys()
    showFlash(t('common.deleted'))
  } catch (err) {
    keyError.value = errorText(err, t('workspace_settings_page.failed_revoke_key'))
    revokeModalKey.value = null
  } finally {
    revokingKey.value = false
  }
}

onMounted(loadAll)
</script>
