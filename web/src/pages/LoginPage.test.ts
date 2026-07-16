import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import LoginPage from './LoginPage.vue'

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  login: vi.fn(),
  setup: vi.fn(),
  getSetupStatus: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mocks.push }),
}))

vi.mock('@/stores/auth', () => ({
  authStore: {
    login: mocks.login,
    setup: mocks.setup,
  },
}))

vi.mock('@/api/client', async () => {
  const actual = await vi.importActual<typeof import('@/api/client')>('@/api/client')
  return {
    ...actual,
    apiClient: {
      getSetupStatus: mocks.getSetupStatus,
    },
  }
})

describe('LoginPage', () => {
  beforeEach(() => {
    mocks.push.mockReset()
    mocks.login.mockReset()
    mocks.setup.mockReset()
    mocks.getSetupStatus.mockReset()
  })

  it('shows the localized credential error for failed login', async () => {
    mocks.getSetupStatus.mockResolvedValueOnce({ needs_setup: false })
    mocks.login.mockRejectedValueOnce(new ApiError(401, 'invalid credentials'))
    const wrapper = mount(LoginPage)
    await vi.dynamicImportSettled()

    await wrapper.get('input').setValue('bad-password')
    await wrapper.get('form').trigger('submit')
    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('密码错误或凭证无效')
    expect(wrapper.text()).not.toContain('invalid credentials')
    expect(mocks.push).not.toHaveBeenCalled()
  })

  it('sets the initial admin password when setup is required', async () => {
    mocks.getSetupStatus.mockResolvedValueOnce({ needs_setup: true })
    mocks.setup.mockResolvedValueOnce(undefined)
    const wrapper = mount(LoginPage)
    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('初始化管理员')
    await wrapper.get('input').setValue('first-password')
    await wrapper.get('form').trigger('submit')
    await vi.dynamicImportSettled()

    expect(mocks.setup).toHaveBeenCalledWith('first-password')
    expect(mocks.login).not.toHaveBeenCalled()
    expect(mocks.push).toHaveBeenCalledWith('/endpoints')
  })
})
