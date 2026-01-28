<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 512 512"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <rect x="80" y="80" width="160" height="160" rx="38" />
    <rect x="272" y="80" width="160" height="160" rx="38" />
    <rect x="80" y="272" width="160" height="160" rx="38" />
    <rect x="272" y="272" width="160" height="160" rx="38" fill="#F5B700" transform="rotate(14 352 352)" />
  </svg>
</template>

<script setup>
// Odd pod mark: a grid of pods, the amber one knocked out of line is the anomaly Mut caught
defineProps({
  alt: {
    type: String,
    default: ''
  }
})
</script>
