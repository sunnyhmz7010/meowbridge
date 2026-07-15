import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import EndpointFormPage from './EndpointFormPage.vue'

const apiMocks = vi.hoisted(() => ({
  getWebhookPresets: vi.fn(),
  previewWebhook: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: {} }),
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/components/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/components/toast', () => ({
  showToast: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  ApiError: class ApiError extends Error {
    status = 500
  },
  defaultParserConfig: () => ({
    mode: 'auto',
    preset: '',
    field_mapping: {},
    default_values: {},
  }),
  apiClient: {
    getWebhookPresets: apiMocks.getWebhookPresets,
    previewWebhook: apiMocks.previewWebhook,
  },
}))

describe('EndpointFormPage', () => {
  it('preserves literal spaces in custom parser fragments', async () => {
    apiMocks.getWebhookPresets.mockResolvedValue([])
    apiMocks.previewWebhook.mockResolvedValue({
      source_type: 'custom',
      title: '',
      msg: 'ok',
      url: '',
      img_url: '',
      msg_type: 'markdown',
    })

    const wrapper = mount(EndpointFormPage)
    await vi.dynamicImportSettled()

    await wrapper.findAll('select').at(1)!.setValue('custom')
    await nextTick()

    const textareas = wrapper.findAll('textarea')
    await textareas.at(1)!.setValue('仓库: \n$.hook.url\n\\n分支: \n$.ref')
    await textareas.at(2)!.setValue('{"hook":{"url":"https://github.com/sunnyhmz7010/meowbridge"},"ref":"refs/heads/main"}')
    await wrapper.find('button.app-button-secondary.px-3').trigger('click')

    expect(apiMocks.previewWebhook).toHaveBeenCalledWith(expect.objectContaining({
      parser_config: expect.objectContaining({
        mode: 'custom',
        field_mapping: expect.objectContaining({
          msg: ['仓库: ', '$.hook.url', '\\n分支: ', '$.ref'],
        }),
      }),
    }))
  })
})
