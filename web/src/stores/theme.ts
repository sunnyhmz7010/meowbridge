import { computed, ref } from 'vue'

export type ThemePreference = 'light' | 'dark' | 'system'
export type ResolvedTheme = 'light' | 'dark'

const storageKey = 'meowbridge_theme'
const mediaQuery = '(prefers-color-scheme: dark)'
const preference = ref<ThemePreference>(readPreference())
const systemTheme = ref<ResolvedTheme>(resolveSystemTheme())
const resolved = computed<ResolvedTheme>(() =>
  preference.value === 'system' ? systemTheme.value : preference.value,
)
let initialized = false

function readPreference(): ThemePreference {
  const value = localStorage.getItem(storageKey)
  return value === 'light' || value === 'dark' || value === 'system' ? value : 'system'
}

function resolveSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined' || !window.matchMedia) {
    return 'dark'
  }
  return window.matchMedia(mediaQuery).matches ? 'dark' : 'light'
}

function applyTheme(): void {
  const root = document.documentElement
  root.dataset.theme = resolved.value
  root.style.colorScheme = resolved.value
}

function handleSystemChange(event: MediaQueryListEvent): void {
  systemTheme.value = event.matches ? 'dark' : 'light'
  applyTheme()
}

function init(): void {
  systemTheme.value = resolveSystemTheme()
  applyTheme()
  if (initialized || typeof window === 'undefined' || !window.matchMedia) {
    return
  }
  window.matchMedia(mediaQuery).addEventListener('change', handleSystemChange)
  initialized = true
}

function setPreference(next: ThemePreference): void {
  preference.value = next
  localStorage.setItem(storageKey, next)
  applyTheme()
}

export const themeStore = {
  preference,
  resolved,
  init,
  setPreference,
}
