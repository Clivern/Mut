<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('settings_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('settings_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <div class="space-y-8">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="border-b border-theme-border px-6 py-4">
              <h2 class="text-lg font-semibold text-theme-text">{{ $t('settings_page.general') }}</h2>
              <p class="mt-1 text-sm text-theme-textLight">{{ $t('settings_page.general_desc') }}</p>
            </div>

            <div v-if="loading" class="p-8 text-center text-sm text-theme-textLight">
              {{ $t('settings_page.loading') }}
            </div>

            <form v-else class="p-6 space-y-5" @submit.prevent="handleSave">
              <div>
                <label for="platform-name" class="form-label">{{ $t('settings_page.platform_name') }}</label>
                <input
                  id="platform-name"
                  v-model="form.platformName"
                  type="text"
                  required
                  maxlength="100"
                  class="input-field"
                  :disabled="saving"
                />
              </div>

              <div>
                <label for="platform-url" class="form-label">{{ $t('settings_page.platform_url') }}</label>
                <input
                  id="platform-url"
                  v-model="form.platformURL"
                  type="url"
                  required
                  maxlength="255"
                  class="input-field"
                  :disabled="saving"
                />
              </div>

              <div>
                <label for="platform-email" class="form-label">{{ $t('settings_page.platform_email') }}</label>
                <input
                  id="platform-email"
                  v-model="form.platformEmail"
                  type="email"
                  required
                  maxlength="255"
                  class="input-field"
                  :disabled="saving"
                />
              </div>

              <div>
                <label for="platform-description" class="form-label">{{
                  $t('settings_page.platform_description')
                }}</label>
                <textarea
                  id="platform-description"
                  v-model="form.platformDescription"
                  rows="3"
                  maxlength="1024"
                  class="input-field"
                  :disabled="saving"
                />
              </div>

              <div class="pt-2">
                <button type="submit" class="btn-primary disabled:opacity-50" :disabled="saving">
                  {{ saving ? $t('settings_page.saving') : $t('settings_page.save') }}
                </button>
              </div>
            </form>
          </section>

          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="border-b border-theme-border px-6 py-4">
              <h2 class="text-lg font-semibold text-theme-text">{{ $t('settings_page.sign_in') }}</h2>
              <p class="mt-1 text-sm text-theme-textLight">{{ $t('settings_page.sign_in_desc') }}</p>
            </div>

            <div class="p-6 space-y-4">
              <div class="rounded-lg border border-theme-border px-4 py-3">
                <span class="text-sm font-medium text-theme-text">{{ $t('settings_page.github') }}</span>
              </div>
              <div class="rounded-lg border border-theme-border px-4 py-3">
                <span class="text-sm font-medium text-theme-text">{{ $t('settings_page.google') }}</span>
              </div>
              <p class="text-xs text-theme-textLight">{{ $t('settings_page.sign_in_hint') }}</p>
            </div>
          </section>
        </div>

        <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('settings_page.maintenance') }}</h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('settings_page.maintenance_desc') }}</p>
          </div>

          <form v-if="!loading" class="p-6 space-y-5" @submit.prevent="handleSave">
            <label class="flex items-start gap-3">
              <input
                v-model="form.maintenanceMode"
                type="checkbox"
                class="mt-1 rounded border-theme-border"
                :disabled="saving"
              />
              <span>
                <span class="block text-sm font-medium text-theme-text">{{ $t('settings_page.maintenance_mode') }}</span>
                <span class="block text-xs text-theme-textLight">{{ $t('settings_page.maintenance_mode_desc') }}</span>
              </span>
            </label>

            <div class="pt-2">
              <button type="submit" class="btn-primary disabled:opacity-50" :disabled="saving">
                {{ saving ? $t('settings_page.saving') : $t('settings_page.save') }}
              </button>
            </div>
          </form>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { settings_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'

const { t } = useI18n()

const form = reactive({
  platformName: '',
  platformURL: '',
  platformEmail: '',
  platformDescription: '',
  maintenanceMode: false,
})

const loading = ref(true)
const saving = ref(false)
const errorMessage = ref(null)

async function loadSettings() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await settings_api.getSettings()
    const settings = res.data?.settings ?? {}
    form.platformName = settings.platform_name ?? ''
    form.platformURL = settings.platform_url ?? ''
    form.platformEmail = settings.platform_email ?? ''
    form.platformDescription = settings.platform_description ?? ''
    form.maintenanceMode = settings.maintenance_mode === '1'
  } catch (err) {
    errorMessage.value = errorText(err, t('settings_page.failed_load'))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  errorMessage.value = null

  try {
    const res = await settings_api.updateSettings({ ...form })
    showFlash(res.data?.successMessage ?? t('common.saved'))
  } catch (err) {
    errorMessage.value = errorText(err, t('settings_page.failed_save'))
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>
