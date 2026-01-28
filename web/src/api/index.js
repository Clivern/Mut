import axios from 'axios'
import { removeUserFromStorage } from '@/utils/storage'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Request interceptor
api.interceptors.request.use((config) => {
  try {
    const locale = typeof localStorage !== 'undefined' && localStorage.getItem('_flocale')
    if (locale === 'en' || locale === 'fr') {
      config.headers['Accept-Language'] = locale
    }
  } catch (_) {}
  return config
})

// Endpoints and pages where a 401 does not redirect to login
const noRedirectUrls = ['public/action/logout']
const publicPages = ['/', '/login', '/setup']

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const { pathname, search } = window.location

    if (error.response?.status === 401 &&
        !noRedirectUrls.includes(error.config?.url) &&
        !publicPages.includes(pathname)) {
      removeUserFromStorage()
      window.location.href = `/login?redirect=${encodeURIComponent(pathname + search)}`
    }
    return Promise.reject(error)
  }
)

// API endpoints
export const health_api = {
  check: () => api.get('public/_health'),
  ready: () => api.get('public/_ready'),
}

// API endpoints for the current user's session and profile
export const auth_api = {
  logout: () => api.post('public/action/logout'),
  me: () => api.get('/me'),
  getProfile: () => api.get('/action/profile'),
  updateProfile: (data) => api.put('/action/profile', data),
}

// API endpoints for the first-run setup
export const setup_api = {
  install: (data) => api.post('public/action/setup', data),
  checkInstalled: () => api.get('public/action/setup/status'),
}

// API endpoints for workspaces
export const workspace_api = {
  list: (params) => api.get('/workspaces', { params }),
  get: (id) => api.get(`/workspaces/${id}`),
  create: (data) => api.post('/workspaces', data),
  update: (id, data) => api.put(`/workspaces/${id}`, data),
  delete: (id) => api.delete(`/workspaces/${id}`),
}

// API endpoints for workspace members and their roles
export const workspace_members_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/members`, { params }),
  updateRole: (workspaceId, userId, data) => api.put(`/workspaces/${workspaceId}/members/${userId}`, data),
  remove: (workspaceId, userId) => api.delete(`/workspaces/${workspaceId}/members/${userId}`),
}

// API endpoints for workspace invites (owners and admins only)
export const workspace_invite_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/invites`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/invites`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/invites/${id}`),
}

// API endpoints for workspace access keys (owners and admins only)
export const workspace_access_keys_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/accessKeys`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/accessKeys`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/accessKeys/${id}`),
}

// API endpoints for workspace clusters
export const cluster_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/clusters`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/clusters`, data),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/clusters/${id}`),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/clusters/${id}`, data),
  remove: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/clusters/${id}`),
  // A full collect can take a while on large clusters, so allow more than the default timeout
  state: (workspaceId, clusterId) =>
    api.get(`/workspaces/${workspaceId}/clusters/${clusterId}/state`, { timeout: 120000 }),
}

// API endpoints for workspace assistant chats (private to the user who holds them)
export const chat_api = {
  list: (workspaceId) => api.get(`/workspaces/${workspaceId}/chats`),
  create: (workspaceId) => api.post(`/workspaces/${workspaceId}/chats`),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/chats/${id}`),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/chats/${id}`),
  send: (workspaceId, id, data, onEvent) =>
    streamEvents(`/workspaces/${workspaceId}/chats/${id}/messages`, data, onEvent),
}

// POSTs data and calls onEvent(name, data) for each server-sent event until the stream ends.
// Axios cannot read a streamed body in the browser and EventSource only does GET, so this uses fetch.
// A refused request throws an axios-shaped error, so errorText reads its message.
async function streamEvents(url, data, onEvent) {
  const res = await fetch(`${api.defaults.baseURL}${url}`, {
    method: 'POST',
    credentials: 'same-origin',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
    },
    body: JSON.stringify(data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw { response: { status: res.status, data: body } }
  }

  const reader = res.body.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ''

  for (;;) {
    const { value, done } = await reader.read()
    if (done) return

    buffer += value
    const events = buffer.split('\n\n')
    buffer = events.pop()

    for (const raw of events) {
      const name = raw.match(/^event: (.*)$/m)?.[1]
      const payload = raw.match(/^data: (.*)$/m)?.[1]
      if (name && payload) onEvent(name, JSON.parse(payload))
    }
  }
}

// API endpoints for settings
export const settings_api = {
  getSettings: () => api.get('/action/settings'),
  updateSettings: (data) => api.put('/action/settings', data),
}

// API endpoints for platform users (admin only)
export const users_api = {
  list: (params) => api.get('/users', { params }),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
}

// API endpoints for current user's API keys
export const api_keys_api = {
  list: (params) => api.get('/apiKeys', { params }),
  create: (data) => api.post('/apiKeys', data),
  delete: (id) => api.delete(`/apiKeys/${id}`),
}

