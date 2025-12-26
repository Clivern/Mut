<template>
  <div id="app" class="min-h-screen">
    <FlashNotice :message="flashMessage" :duration="flashDuration" @dismiss="clearFlash" />
    <CookieConsent />
    <router-view />
  </div>
</template>

<script setup>
import CookieConsent from '@/components/CookieConsent.vue'
import FlashNotice from '@/components/FlashNotice.vue'
import { flashMessage, flashDuration, clearFlash } from '@/lib/flash'
</script>

<style scoped>
/* Additional app-level styles can go here */
</style>

