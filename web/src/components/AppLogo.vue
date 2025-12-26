<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 512 512"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <template v-if="variant === 'tile'">
      <rect width="512" height="512" rx="116" />
      <g transform="translate(56 56) scale(.78125)" fill="none" stroke="var(--theme-bg)">
        <circle cx="256" cy="256" r="206" stroke-width="34" />
        <circle cx="256" cy="256" r="112" stroke-width="26" />
        <path d="M256 256 L392 120" stroke-width="30" stroke-linecap="round" />
        <circle cx="256" cy="256" r="28" fill="var(--theme-bg)" stroke="none" />
        <circle cx="352" cy="332" r="32" fill="#F5B700" stroke="none" />
      </g>
    </template>
    <template v-else>
      <g fill="none" stroke="currentColor">
        <circle cx="256" cy="256" r="206" stroke-width="34" />
        <circle cx="256" cy="256" r="112" stroke-width="26" />
        <path d="M256 256 L392 120" stroke-width="30" stroke-linecap="round" />
      </g>
      <circle cx="256" cy="256" r="28" />
      <circle cx="352" cy="332" r="32" fill="#F5B700" />
    </template>
  </svg>
</template>

<script setup>
// Radar mark: a sweep watching the cluster, the amber blip is the anomaly it caught
defineProps({
  alt: {
    type: String,
    default: ''
  },
  variant: {
    type: String,
    default: 'mark'
  }
})
</script>
