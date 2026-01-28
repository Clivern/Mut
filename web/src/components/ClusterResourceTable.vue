<template>
  <section class="section overflow-hidden p-0" :aria-labelledby="headingId">
    <div class="border-b border-theme-border px-6 py-4">
      <h2 :id="headingId" class="text-lg font-semibold text-theme-text">
        {{ title }}
        <span class="ml-2 text-sm font-normal tabular-nums text-theme-textLight">{{ rows.length }}</span>
      </h2>
      <p class="mt-0.5 text-sm text-theme-textLight">{{ description }}</p>
    </div>

    <div v-if="rows.length" class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs font-medium text-theme-textLight">
            <th v-for="column in columns" :key="column.label" class="px-6 py-3 font-medium">{{ column.label }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-theme-border">
          <tr v-for="(row, index) in rows" :key="index" class="text-theme-text">
            <td v-for="column in columns" :key="column.label" class="px-6 py-3" :class="column.tone ? column.tone(row) : ''">
              {{ column.value(row) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <p v-else class="px-6 py-6 text-sm" :class="error ? 'text-red-700' : 'text-theme-textLight'">
      {{ error ? $t('dashboard.resources.not_collected') : emptyText }}
    </p>
  </section>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: { type: String, required: true },
  description: { type: String, required: true },
  columns: { type: Array, required: true },
  rows: { type: Array, required: true },
  error: { type: String, default: null },
  emptyText: { type: String, required: true },
})

const headingId = computed(() => `resource-${props.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`)
</script>
