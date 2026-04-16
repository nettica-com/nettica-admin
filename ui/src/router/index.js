import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/join/:pathMatch(.*)*',
    name: 'join',
    component: () => import('../views/Join.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/consent/:pathMatch(.*)*',
    name: 'consent',
    component: () => import('../views/Consent.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/agent',
    name: 'agent',
    component: () => import('../views/Agent.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/devices',
    name: 'devices',
    component: () => import('../views/Devices.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/networks',
    name: 'networks',
    component: () => import('../views/Network.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/accounts',
    name: 'accounts',
    component: () => import('../views/Accounts.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/services',
    name: 'services',
    component: () => import('../views/Services.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/login/:pathMatch(.*)*',
    name: 'login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    name: 'root',
    meta: { requiresAuth: false },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  console.log('router =', to.path)
  if (to.matched.some((record) => record.meta.requiresAuth)) {
    authStore.requiresAuth = true
    if (!authStore.isAuthenticated) {
      authStore.intendedRoute = to.fullPath
      return { path: '/' }
    }
  } else {
    authStore.requiresAuth = false
  }
})

export default router
