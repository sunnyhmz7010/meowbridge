import { computed, ref } from 'vue'
import { apiClient, setAuthTokenProvider, setUnauthorizedHandler } from '@/api/client'

const storageKey = 'meowbridge_token'
const token = ref(localStorage.getItem(storageKey) ?? '')

function persist(nextToken: string): void {
  token.value = nextToken
  if (nextToken) {
    localStorage.setItem(storageKey, nextToken)
  } else {
    localStorage.removeItem(storageKey)
  }
}

export const authStore = {
  token,
  isAuthenticated: computed(() => token.value.length > 0),
  async login(password: string): Promise<void> {
    persist(await apiClient.login(password))
  },
  logout(): void {
    persist('')
  },
}

setAuthTokenProvider(() => token.value || null)
setUnauthorizedHandler(() => authStore.logout())
