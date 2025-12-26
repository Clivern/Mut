<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('dashboard.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('dashboard.subtitle') }}
        </p>
        <p v-if="user?.email" class="mt-2 text-sm text-theme-text">
          {{ $t('dashboard.welcome_back') }}
          <span class="font-medium">{{ user.name || user.email }}</span>
        </p>
      </header>

      <section v-if="showGettingStarted" class="mb-8 overflow-hidden rounded-xl border border-theme-border bg-white shadow-sm">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-theme-border px-6 py-4">
          <div class="min-w-0">
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('dashboard.getting_started') }}</h2>
            <p class="mt-0.5 text-sm text-theme-textLight">{{ $t('dashboard.getting_started_desc') }}</p>
          </div>
          <button
            type="button"
            class="rounded-md px-2 py-1 text-sm text-theme-textLight transition hover:bg-theme-hover hover:text-theme-text"
            @click="dismissGettingStarted"
          >
            {{ $t('dashboard.getting_started_dismiss') }}
          </button>
        </div>

        <ol class="grid grid-cols-1 gap-px bg-theme-border sm:grid-cols-2 lg:grid-cols-3">
          <li v-for="(step, index) in gettingStartedSteps" :key="step.id" class="flex gap-4 bg-white px-6 py-5">
            <span
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-theme-hover text-sm font-semibold text-theme-text"
              aria-hidden="true"
            >
              {{ index + 1 }}
            </span>
            <span class="sr-only">{{ $t('dashboard.step_label', { number: index + 1 }) }}</span>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-theme-text">{{ step.title }}</p>
              <p class="mt-0.5 text-sm leading-relaxed text-theme-textLight">{{ step.description }}</p>
              <router-link
                :to="step.to"
                class="mt-2 inline-flex text-sm font-medium text-primary-800 hover:underline"
              >
                {{ step.action }}
              </router-link>
            </div>
          </li>
        </ol>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import { user } from '@/lib/auth'

const { t } = useI18n()

const GETTING_STARTED_KEY = 'dashboard_getting_started_v1'

const showGettingStarted = ref(localStorage.getItem(GETTING_STARTED_KEY) !== '1')

function dismissGettingStarted() {
  showGettingStarted.value = false
  localStorage.setItem(GETTING_STARTED_KEY, '1')
}

const gettingStartedSteps = computed(() => [
  {
    id: 'profile',
    title: t('dashboard.step_profile'),
    description: t('dashboard.step_profile_desc'),
    action: t('dashboard.step_profile_link'),
    to: '/profile',
  },
  {
    id: 'users',
    title: t('dashboard.step_users'),
    description: t('dashboard.step_users_desc'),
    action: t('dashboard.step_users_link'),
    to: '/users',
    requiresAdmin: true,
  },
  {
    id: 'settings',
    title: t('dashboard.step_settings'),
    description: t('dashboard.step_settings_desc'),
    action: t('dashboard.step_settings_link'),
    to: '/settings',
    requiresAdmin: true,
  },
].filter((step) => !step.requiresAdmin || user.value?.role === 'admin'))
</script>
