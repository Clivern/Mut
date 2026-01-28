<template>
  <div id="app" class="min-h-screen">
    <FlashNotice :message="flashMessage" :duration="flashDuration" @dismiss="clearFlash" />
    <CookieConsent />
    <router-view />
    <AssistantPanel v-if="route.meta.requiresWorkspace" />
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
import AssistantPanel from '@/components/AssistantPanel.vue'
import CookieConsent from '@/components/CookieConsent.vue'
import FlashNotice from '@/components/FlashNotice.vue'
import { flashMessage, flashDuration, clearFlash } from '@/lib/flash'

const route = useRoute()
</script>

<style scoped>
/* Additional app-level styles can go here */
</style>

