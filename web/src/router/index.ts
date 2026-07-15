import { createRouter, createWebHistory } from 'vue-router'
import { setUnauthorizedHandler } from '@/api/client'
import { authStore } from '@/stores/auth'
import LoginPage from '@/pages/LoginPage.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/', redirect: '/endpoints' },
    { path: '/login', name: 'login', component: LoginPage, meta: { public: true } },
    {
      path: '/endpoints',
      name: 'endpoints',
      component: () => import('@/pages/EndpointsPage.vue'),
    },
    {
      path: '/endpoints/new',
      name: 'endpoint-new',
      component: () => import('@/pages/EndpointFormPage.vue'),
    },
    {
      path: '/endpoints/:id',
      name: 'endpoint-edit',
      component: () => import('@/pages/EndpointFormPage.vue'),
      props: true,
    },
    { path: '/logs', name: 'logs', component: () => import('@/pages/LogsPage.vue') },
    {
      path: '/logs/:id',
      name: 'log-detail',
      component: () => import('@/pages/LogDetailPage.vue'),
      props: true,
    },
    { path: '/settings', name: 'settings', component: () => import('@/pages/SettingsPage.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/endpoints' },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) {
    return authStore.isAuthenticated.value && to.name === 'login' ? '/endpoints' : true
  }
  return authStore.isAuthenticated.value ? true : '/login'
})

setUnauthorizedHandler(() => {
  authStore.logout()
  if (router.currentRoute.value.name !== 'login') {
    router.push('/login')
  }
})

export default router
