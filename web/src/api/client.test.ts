import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  ApiError,
  apiClient,
  normalizeEndpoint,
  setAuthTokenProvider,
  setUnauthorizedHandler,
} from './client'

describe('apiClient', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    setAuthTokenProvider(() => null)
    setUnauthorizedHandler(() => {})
  })

  it('unwraps successful login response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({ ok: true, data: { token: 'jwt-token' } }, { status: 200 }),
      ),
    )

    await expect(apiClient.login('secret')).resolves.toBe('jwt-token')
  })

  it('loads webhook presets and previews parser results', async () => {
    const fetchMock = vi.fn(async (path: RequestInfo | URL) => {
      if (String(path).endsWith('/webhook/presets')) {
        return Response.json({
          ok: true,
          data: [{ id: 'github_push_minimal', name: 'GitHub 简化 Push', field_mapping: {}, default_values: {} }],
        }, { status: 200 })
      }
      return Response.json({
        ok: true,
        data: {
          source_type: 'github_push_minimal',
          title: 'GitHub Push',
          msg: '分支: main',
          url: '',
          img_url: '',
          msg_type: 'markdown',
        },
      }, { status: 200 })
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiClient.getWebhookPresets()).resolves.toHaveLength(1)
    await expect(apiClient.previewWebhook({
      parser_config: { mode: 'preset', preset: 'github_push_minimal', field_mapping: {}, default_values: {} },
      payload: { ref: 'refs/heads/main' },
    })).resolves.toMatchObject({ title: 'GitHub Push', msg_type: 'markdown' })
  })

  it('sends bearer token for authenticated requests', async () => {
    const fetchMock = vi.fn(async () => Response.json({ ok: true, data: [] }, { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)
    setAuthTokenProvider(() => 'abc')

    await apiClient.listEndpoints()

    const calls = fetchMock.mock.calls as unknown as Array<[RequestInfo | URL, RequestInit]>
    const requestInit = calls[0][1]
    const headers = requestInit.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer abc')
  })

  it('throws ApiError for backend error response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ ok: false, error: 'invalid credentials' }, { status: 401 })),
    )

    await expect(apiClient.login('bad')).rejects.toMatchObject({
      status: 401,
      message: 'invalid credentials',
    })
  })

  it('calls unauthorized handler on 401', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ ok: false, error: 'invalid token' }, { status: 401 })),
    )

    await expect(apiClient.listEndpoints()).rejects.toBeInstanceOf(ApiError)
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('does not treat a wrong current password as an expired login', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ ok: false, error: 'invalid credentials' }, { status: 401 })),
    )

    await expect(apiClient.changePassword('wrong', 'new-secret')).rejects.toMatchObject({
      status: 401,
      message: 'invalid credentials',
    })
    expect(onUnauthorized).not.toHaveBeenCalled()
  })

  it('treats an invalid token during password change as an expired login', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ ok: false, error: 'invalid token' }, { status: 401 })),
    )

    await expect(apiClient.changePassword('old', 'new-secret')).rejects.toMatchObject({
      status: 401,
      message: 'invalid token',
    })
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })
})

describe('normalizeEndpoint', () => {
  it('normalizes stable snake_case JSON field names', () => {
    expect(
      normalizeEndpoint({
        id: 7,
        name: 'Build',
        token: 'token-1',
        meow_nickname: 'sunny',
        default_title: '',
        msg_type: 'markdown',
        html_height: 300,
        default_url: '',
        default_img_url: '',
        parser_config: { mode: 'preset', preset: 'github_push_minimal', field_mapping: {}, default_values: {} },
        active: true,
        created_at: '',
        updated_at: '',
      }),
    ).toMatchObject({
      id: 7,
      name: 'Build',
      token: 'token-1',
      meow_nickname: 'sunny',
      msg_type: 'markdown',
      html_height: 300,
      parser_config: { mode: 'preset', preset: 'github_push_minimal', field_mapping: {}, default_values: {} },
      active: true,
    })
  })

  it('defaults missing parser_config to auto mode', () => {
    expect(
      normalizeEndpoint({
        id: 8,
        name: 'Build',
        token: 'token-2',
        meow_nickname: 'sunny',
        default_title: '',
        msg_type: 'text',
        html_height: 200,
        default_url: '',
        default_img_url: '',
        parser_config: null,
        active: true,
        created_at: '',
        updated_at: '',
      }),
    ).toMatchObject({
      parser_config: { mode: 'auto', preset: '', field_mapping: {}, default_values: {} },
    })
  })
})
