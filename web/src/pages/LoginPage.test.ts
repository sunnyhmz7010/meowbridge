import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import LoginPage from './LoginPage.vue'

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  login: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mocks.push }),
}))

vi.mock('@/stores/auth', () => ({
  authStore: {
    login: mocks.login,
  },
}))

describe('LoginPage', () => {
  it('shows the localized credential error for failed login', async () => {
    mocks.login.mockRejectedValueOnce(new ApiError(401, 'invalid credentials'))
    const wrapper = mount(LoginPage)

    await wrapper.get('input').setValue('bad-password')
    await wrapper.get('form').trigger('submit')
    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('密码错误或凭证无效')
    expect(wrapper.text()).not.toContain('invalid credentials')
    expect(mocks.push).not.toHaveBeenCalled()
  })
})
