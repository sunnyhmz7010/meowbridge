import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import EndpointsPage from './EndpointsPage.vue'

const endpoint = {
  id: 1,
  name: 'Build',
  token: 'old-token',
  meow_nickname: 'sunny',
  default_title: 'title',
  msg_type: 'text',
  html_height: 200,
  default_url: '',
  default_img_url: '',
  active: true,
  created_at: '',
  updated_at: '',
}

const apiMocks = vi.hoisted(() => ({
  listEndpoints: vi.fn(),
  resetEndpointToken: vi.fn(),
  deleteEndpoint: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/components/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/components/EmptyState.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/components/toast', () => ({
  showToast: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  ApiError: class ApiError extends Error {
    status = 500
  },
  apiClient: {
    listEndpoints: apiMocks.listEndpoints,
    resetEndpointToken: apiMocks.resetEndpointToken,
    deleteEndpoint: apiMocks.deleteEndpoint,
  },
}))

describe('EndpointsPage', () => {
  it('clears stale manual copy URL after resetting token', async () => {
    apiMocks.listEndpoints.mockResolvedValue([endpoint])
    apiMocks.resetEndpointToken.mockResolvedValue({ ...endpoint, token: 'new-token' })
    Object.assign(navigator, { clipboard: undefined })

    const wrapper = mount(EndpointsPage, {
      attachTo: document.body,
    })
    await vi.dynamicImportSettled()

    await wrapper.findAll('button').find((button) => button.text() === '复制 URL')!.trigger('click')
    expect(wrapper.text()).toContain('/webhook/old-token')

    await wrapper.findAll('button').find((button) => button.text() === '重置 token')!.trigger('click')
    await wrapper.findAll('button').find((button) => button.text() === '重置')!.trigger('click')
    await vi.dynamicImportSettled()

    expect(wrapper.text()).not.toContain('/webhook/old-token')
  })

  it('clears stale manual copy URL after deleting endpoint', async () => {
    apiMocks.listEndpoints.mockResolvedValue([endpoint])
    apiMocks.deleteEndpoint.mockResolvedValue({ deleted: true })
    Object.assign(navigator, { clipboard: undefined })

    const wrapper = mount(EndpointsPage, {
      attachTo: document.body,
    })
    await vi.dynamicImportSettled()

    await wrapper.findAll('button').find((button) => button.text() === '复制 URL')!.trigger('click')
    expect(wrapper.text()).toContain('/webhook/old-token')

    await wrapper.findAll('button').find((button) => button.text() === '删除')!.trigger('click')
    await wrapper.findAll('button').filter((button) => button.text() === '删除').at(-1)!.trigger('click')
    await vi.dynamicImportSettled()

    expect(wrapper.text()).not.toContain('/webhook/old-token')
  })
})
