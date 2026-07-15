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
      active: true,
    })
  })
})
