<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <header>
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
        <span class="mt-2 inline-block text-xs text-theme-textLight">{{ $t('dashboard.sample_note') }}</span>
      </header>

      <ClusterBoard />
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/components/AppNav.vue'
import ClusterBoard from '@/components/ClusterBoard.vue'
import { user } from '@/lib/auth'
</script>
