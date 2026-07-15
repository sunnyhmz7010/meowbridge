import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

function installMatchMedia(matches: boolean): void {
  vi.stubGlobal('matchMedia', vi.fn(() => ({
    matches,
    media: '(prefers-color-scheme: dark)',
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })))
}

describe('themeStore', () => {
  beforeEach(() => {
    vi.resetModules()
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
    installMatchMedia(true)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('defaults to system and resolves dark when the system is dark', async () => {
    const { themeStore } = await import('./theme')

    themeStore.init()

    expect(themeStore.preference.value).toBe('system')
    expect(themeStore.resolved.value).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('persists explicit light theme and applies it to the document', async () => {
    const { themeStore } = await import('./theme')

    themeStore.setPreference('light')

    expect(localStorage.getItem('meowbridge_theme')).toBe('light')
    expect(themeStore.preference.value).toBe('light')
    expect(themeStore.resolved.value).toBe('light')
    expect(document.documentElement.dataset.theme).toBe('light')
  })
})
