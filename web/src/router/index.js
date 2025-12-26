import { createRouter, createWebHistory } from 'vue-router'
import { loadUserFromStorage } from '@/utils/storage'

const routes = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresGuest: true, title: 'Login', description: 'Login to platform' }
  },
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/Setup.vue'),
    meta: { requiresGuest: true, title: 'Setup', description: 'Setup platform' }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { requiresAuth: true, title: 'Dashboard', description: 'Platform dashboard' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/Profile.vue'),
    meta: { requiresAuth: true, title: 'Profile', description: 'Your account profile' }
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/Users.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, title: 'Users', description: 'Platform users' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, title: 'Settings', description: 'Settings' }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' }
  },
  {
    path: '/500',
    name: 'ServerError',
    component: () => import('@/views/ServerError.vue'),
    meta: { title: 'Server error' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFoundCatchAll',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      return { el: to.hash, behavior: 'smooth' }
    }
    return { top: 0 }
  },
})

function isSafeRedirect(redirect) {
  return typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')
}

// Navigation guard - reads auth state directly from storage
router.beforeEach((to) => {
  const currentUser = loadUserFromStorage()
  const isAuthenticated = !!currentUser

  if (to.meta.requiresAuth && !isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresGuest && isAuthenticated) {
    const redirect = to.query.redirect
    return isSafeRedirect(redirect) ? redirect : '/dashboard'
  }
  if (to.meta.requiresAdmin && currentUser?.role !== 'admin') {
    return '/404'
  }
})

export default router
