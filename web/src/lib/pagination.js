import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

function parsePage(value) {
  const page = Number.parseInt(String(value ?? ''), 10)
  return Number.isFinite(page) && page >= 1 ? page : 1
}

export function useUrlPagination({ limit = 50, queryKey = 'page' } = {}) {
  const route = useRoute()
  const router = useRouter()
  const pageSize = limit

  const page = computed(() => parsePage(route.query[queryKey]))
  const offset = computed(() => (page.value - 1) * pageSize)

  function goToPage(nextPage) {
    const normalized = Math.max(1, Math.floor(nextPage))
    const query = { ...route.query }
    if (normalized === 1) {
      delete query[queryKey]
    } else {
      query[queryKey] = String(normalized)
    }
    return router.replace({ query })
  }

  function goToOffset(newOffset) {
    goToPage(Math.floor(newOffset / pageSize) + 1)
  }

  return {
    page,
    offset,
    limit: pageSize,
    goToPage,
    goToOffset,
  }
}
