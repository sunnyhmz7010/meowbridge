import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsPage from './SettingsPage.vue'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/components/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/components/ConfirmDialog.vue', () => ({
  default: { template: '<div />', props: ['open'] },
}))

vi.mock('@/components/toast', () => ({
  showToast: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  authStore: { logout: vi.fn() },
}))

vi.mock('@/api/client', () => ({
  ApiError: class ApiError extends Error {
    status = 500
  },
  apiClient: {
    getSettings: vi.fn(async () => ({ log_retention_days: '14' })),
    updateSettings: vi.fn(),
    changePassword: vi.fn(),
  },
}))

describe('SettingsPage', () => {
  it('only exposes log retention and password settings', async () => {
    const wrapper = mount(SettingsPage)

    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('日志保留天数')
    expect(wrapper.text()).toContain('修改密码')
    expect(wrapper.text()).not.toContain('MeoW API Base URL')
  })
})
