<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 512 512"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <g fill="none" stroke="currentColor">
      <circle cx="256" cy="256" r="206" stroke-width="34" />
      <circle cx="256" cy="256" r="112" stroke-width="26" />
      <path d="M256 256 L392 120" stroke-width="30" stroke-linecap="round" />
    </g>
    <circle cx="256" cy="256" r="28" />
    <circle cx="352" cy="332" r="32" fill="#F5B700" />
  </svg>
</template>

<script setup>
// Radar mark: a sweep watching the cluster, the amber blip is the anomaly it caught
defineProps({
  alt: {
    type: String,
    default: ''
  }
})
</script>
