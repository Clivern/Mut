<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 512 512"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <rect x="90" y="90" width="92" height="92" rx="22" />
    <rect x="210" y="90" width="92" height="92" rx="22" />
    <rect x="330" y="90" width="92" height="92" rx="22" />
    <rect x="90" y="210" width="92" height="92" rx="22" />
    <rect x="210" y="210" width="92" height="92" rx="22" />
    <rect x="330" y="210" width="92" height="92" rx="22" />
    <rect x="90" y="330" width="92" height="92" rx="22" />
    <rect x="210" y="330" width="92" height="92" rx="22" />
    <rect x="342" y="342" width="92" height="92" rx="22" fill="#F5B700" transform="rotate(14 388 388)" />
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
