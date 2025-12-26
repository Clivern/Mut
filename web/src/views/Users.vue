<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <div class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <main>
        <div class="mb-8 flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-2xl font-semibold text-theme-text">{{ $t('users_page.title') }}</h1>
            <p class="text-sm text-theme-textLight mt-1">{{ $t('users_page.subtitle') }}</p>
          </div>
          <button type="button" class="btn-primary" @click="openCreate">
            {{ $t('users_page.add_user') }}
          </button>
        </div>

        <div v-if="errorMessage" class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {{ errorMessage }}
        </div>

        <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
          <div v-if="loading" class="p-8 text-center text-sm text-theme-textLight">
            {{ $t('users_page.loading') }}
          </div>
          <div v-else-if="users.length === 0" class="p-8 text-center text-sm text-theme-textLight">
            {{ $t('users_page.no_users') }}
          </div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-bg">
                <tr>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('users_page.name') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('users_page.role') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('users_page.status') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('users_page.created') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('users_page.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="item in users" :key="item.id" class="hover:bg-theme-hover/50">
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-3">
                      <img :src="item.avatar" :alt="item.email" class="h-8 w-8 rounded-full">
                      <div class="min-w-0">
                        <p class="text-sm font-medium text-theme-text truncate">{{ item.name }}</p>
                        <p class="text-xs text-theme-textLight truncate">{{ item.email }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-sm text-theme-text">{{ $t(`users_page.role_${item.role}`) }}</td>
                  <td class="px-4 py-3">
                    <span
                      class="rounded-full px-2.5 py-0.5 text-xs font-medium"
                      :class="item.isActive ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' : 'bg-theme-hover text-theme-textLight border border-theme-border'"
                    >
                      {{ item.isActive ? $t('users_page.active') : $t('users_page.inactive') }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-sm text-theme-textLight">{{ formatCreatedAt(item.createdAt, locale) }}</td>
                  <td class="px-4 py-3 text-right space-x-3 whitespace-nowrap">
                    <button type="button" class="text-sm font-medium text-primary-800 hover:underline" @click="openEdit(item)">
                      {{ $t('users_page.edit') }}
                    </button>
                    <button
                      v-if="item.id !== user?.id"
                      type="button"
                      class="text-sm text-red-600 hover:text-red-700 font-medium"
                      @click="deleteModalUser = item"
                    >
                      {{ $t('users_page.delete') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="total > limit" class="flex items-center justify-between border-t border-theme-border px-4 py-3 text-sm">
            <button type="button" class="btn-secondary disabled:opacity-50" :disabled="page <= 1" @click="goToPage(page - 1)">
              {{ $t('users_page.previous') }}
            </button>
            <span class="text-theme-textLight">{{ $t('users_page.page_of', { page, pages }) }}</span>
            <button type="button" class="btn-secondary disabled:opacity-50" :disabled="page >= pages" @click="goToPage(page + 1)">
              {{ $t('users_page.next') }}
            </button>
          </div>
        </section>
      </main>
    </div>

    <div v-if="formOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4" @click.self="closeForm">
      <div class="w-full max-w-md rounded-xl border border-theme-border bg-white shadow-lg">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-base font-semibold text-theme-text">
            {{ editing ? $t('users_page.edit_user') : $t('users_page.add_user') }}
          </h2>
        </div>
        <form class="p-6 space-y-4" @submit.prevent="handleSave">
          <div>
            <label for="user-name" class="form-label">{{ $t('users_page.name') }}</label>
            <input id="user-name" v-model="form.name" type="text" required minlength="2" maxlength="60" class="input-field" :disabled="saving">
          </div>
          <div>
            <label for="user-email" class="form-label">{{ $t('users_page.email') }}</label>
            <input id="user-email" v-model="form.email" type="email" required maxlength="60" class="input-field" :disabled="saving">
          </div>
          <div>
            <label for="user-password" class="form-label">{{ $t('users_page.password') }}</label>
            <input
              id="user-password"
              v-model="form.password"
              type="password"
              :required="!editing"
              minlength="8"
              maxlength="60"
              autocomplete="new-password"
              class="input-field"
              :placeholder="editing ? $t('users_page.password_keep') : ''"
              :disabled="saving"
            >
          </div>
          <div>
            <label for="user-role" class="form-label">{{ $t('users_page.role') }}</label>
            <select id="user-role" v-model="form.role" class="input-field" :disabled="saving">
              <option v-for="role in roles" :key="role" :value="role">{{ $t(`users_page.role_${role}`) }}</option>
            </select>
          </div>
          <label class="flex items-center gap-2 text-sm text-theme-text">
            <input v-model="form.isActive" type="checkbox" class="rounded border-theme-border" :disabled="saving">
            {{ $t('users_page.active') }}
          </label>

          <div v-if="formError" class="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">
            {{ formError }}
          </div>

          <div class="flex justify-end gap-3 pt-2">
            <button type="button" class="btn-secondary" :disabled="saving" @click="closeForm">
              {{ $t('users_page.cancel') }}
            </button>
            <button type="submit" class="btn-primary disabled:opacity-50" :disabled="saving">
              {{ saving ? $t('users_page.saving') : $t('users_page.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <ConfirmModal
      :open="!!deleteModalUser"
      :title="$t('users_page.delete_title')"
      :message="$t('users_page.delete_confirm', { email: deleteModalUser?.email })"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('users_page.delete')"
      :busy-label="$t('users_page.deleting')"
      @cancel="deleteModalUser = null"
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { users_api } from '@/api'
import { user } from '@/lib/auth'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'
import { useUrlPagination } from '@/lib/pagination'
import { formatCreatedAt } from '@/utils/date'

const { t, locale } = useI18n()
const { page, offset, limit, goToPage } = useUrlPagination({ limit: 25 })

const roles = ['admin', 'regular', 'readonly']

const users = ref([])
const total = ref(0)
const loading = ref(true)
const errorMessage = ref(null)

const formOpen = ref(false)
const editing = ref(null)
const saving = ref(false)
const formError = ref(null)
const form = reactive({
  name: '',
  email: '',
  password: '',
  role: 'regular',
  isActive: true
})

const deleteModalUser = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

const pages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

async function loadUsers() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await users_api.list({ limit, offset: offset.value })
    users.value = res.data?.users ?? []
    total.value = res.data?._meta?.total ?? 0
  } catch (err) {
    errorMessage.value = errorText(err, t('users_page.failed_load'))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', email: '', password: '', role: 'regular', isActive: true })
  formError.value = null
  formOpen.value = true
}

function openEdit(item) {
  editing.value = item
  Object.assign(form, { name: item.name, email: item.email, password: '', role: item.role, isActive: item.isActive })
  formError.value = null
  formOpen.value = true
}

function closeForm() {
  if (!saving.value) formOpen.value = false
}

async function handleSave() {
  saving.value = true
  formError.value = null

  try {
    const res = editing.value
      ? await users_api.update(editing.value.id, { ...form })
      : await users_api.create({ ...form })
    showFlash(res.data?.successMessage ?? t('common.saved'))
    formOpen.value = false
    await loadUsers()
  } catch (err) {
    formError.value = errorText(err, t('users_page.failed_save'))
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await users_api.delete(deleteModalUser.value.id)
    deleteModalUser.value = null
    await loadUsers()
  } catch (err) {
    deleteError.value = errorText(err, t('users_page.failed_delete'))
  } finally {
    deleting.value = false
  }
}

watch(page, loadUsers, { immediate: true })
</script>
