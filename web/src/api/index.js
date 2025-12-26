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
const noRedirectUrls = ['public/action/login', 'public/action/logout']
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
  login: (data) => api.post('public/action/login', data),
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

