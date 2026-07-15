import { afterEach, describe, expect, it } from 'vitest'
import router from './index'
import { authStore } from '@/stores/auth'

describe('admin router auth guard', () => {
  afterEach(async () => {
    authStore.logout()
    await router.push('/login')
  })

  it('redirects unauthenticated users to login', async () => {
    authStore.logout()

    await router.push('/endpoints')

    expect(router.currentRoute.value.name).toBe('login')
  })

  it('redirects authenticated users away from login to endpoints', async () => {
    authStore.token.value = 'jwt-token'
    localStorage.setItem('meowbridge_token', 'jwt-token')

    await router.push('/logs')
    await router.push('/login')

    expect(router.currentRoute.value.name).toBe('endpoints')
  })
})
