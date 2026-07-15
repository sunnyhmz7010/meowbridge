import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import App from './App.vue'

vi.mock('@/components/ToastHost.vue', () => ({
  default: { template: '<aside data-test="toast-host" />' },
}))

describe('App', () => {
  it('keeps toast host outside routed layouts', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          RouterView: { template: '<main data-test="router-view" />' },
        },
      },
    })

    expect(wrapper.find('[data-test="router-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="toast-host"]').exists()).toBe(true)
  })
})
