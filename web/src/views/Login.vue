<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-sm w-full">
      <div class="text-center mb-10">
        <router-link to="/" class="inline-flex justify-center text-theme-text">
          <AppLogo :alt="$t('common.logo_alt')" class="h-28 w-auto" />
        </router-link>
      </div>

      <div class="text-center mb-8">
        <h1 class="text-2xl font-semibold text-theme-text tracking-tight">
          {{ $t('login.title') }}
        </h1>
        <p class="mt-2 text-sm text-theme-textLight leading-relaxed">
          {{ $t('login.subtitle') }}
        </p>
      </div>

      <div class="bg-white rounded-lg border border-theme-border p-8 shadow-sm">
        <form class="space-y-5" @submit.prevent="handleLogin">
          <div>
            <label for="email" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('login.email') }}
            </label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              class="input-field"
              :placeholder="$t('login.email_placeholder')"
              :disabled="loading"
            >
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('login.password') }}
            </label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              required
              class="input-field"
              :placeholder="$t('login.password_placeholder')"
              :disabled="loading"
            >
          </div>

          <label class="flex items-center gap-2 text-sm text-theme-text">
            <input v-model="form.rememberMe" type="checkbox" class="rounded border-theme-border" :disabled="loading">
            {{ $t('login.remember_me') }}
          </label>

          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
            <p class="text-sm text-red-800 text-center">
              {{ error }}
            </p>
          </div>

          <button
            type="submit"
            class="btn-primary w-full py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="loading"
          >
            <span v-if="!loading">{{ $t('login.submit') }}</span>
            <span v-else class="flex items-center justify-center">
              <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              {{ $t('login.submitting') }}
            </span>
          </button>
        </form>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-10">
        {{ $t('common.copyright') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { auth_api, setup_api } from '@/api'
import { saveUser } from '@/lib/auth'
import { applyUserPreferences } from '@/lib/preferences'
import AppLogo from '@/components/AppLogo.vue'
import { errorText } from '@/lib/error'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const form = reactive({
  email: '',
  password: '',
  rememberMe: false
})

const loading = ref(false)
const error = ref(null)

function getRedirectTarget() {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }
  return '/dashboard'
}

const handleLogin = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await auth_api.login({
      email: form.email,
      password: form.password,
      rememberMe: form.rememberMe
    })
    saveUser(response.data?.user)
    applyUserPreferences(response.data?.user)
    router.replace(getRedirectTarget())
  } catch (err) {
    error.value = errorText(err, t('login.error_default'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const response = await setup_api.checkInstalled()
    if (!response.data.installed) {
      router.push('/setup')
    }
  } catch (err) {
    console.error('Failed to check setup status:', err)
  }
})
</script>
