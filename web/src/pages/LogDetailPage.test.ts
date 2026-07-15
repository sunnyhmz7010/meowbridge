import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import LogDetailPage from './LogDetailPage.vue'

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: '1' } }),
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/components/AppLayout.vue', () => ({
  default: { template: '<div><slot /></div>' },
}))

vi.mock('@/api/client', () => ({
  ApiError: class ApiError extends Error {
    status = 500
  },
  apiClient: {
    getPushLog: vi.fn(async () => ({
      id: 1,
      endpoint_id: 1,
      endpoint_name: 'Build',
      token: 'token',
      source_type: 'standard',
      request_method: 'POST',
      request_headers: '{}',
      request_query: '{}',
      request_payload: '{}',
      parsed_title: 'title',
      parsed_msg: 'message',
      parsed_msg_type: 'text',
      meow_status_code: 200,
      meow_response_body: '{}',
      success: true,
      error_message: '',
      created_at: '2026-07-15T00:00:00Z',
    })),
  },
}))

describe('LogDetailPage', () => {
  it('shows request method in the summary', async () => {
    const wrapper = mount(LogDetailPage)

    await vi.dynamicImportSettled()

    expect(wrapper.text()).toContain('请求方法')
    expect(wrapper.text()).toContain('POST')
  })
})
