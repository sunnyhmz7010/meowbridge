import type {
  Endpoint,
  EndpointInput,
  EndpointUpdate,
  EndpointView,
  PushLog,
  PushLogListItem,
} from './types'

type SuccessResponse<T> = { ok: true; data: T }
type ErrorResponse = { ok: false; error: string }

let tokenProvider: () => string | null = () => localStorage.getItem('meowbridge_token')
let unauthorizedHandler: () => void = () => {}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function setAuthTokenProvider(provider: () => string | null): void {
  tokenProvider = provider
}

export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

export function normalizeEndpoint(input: Endpoint): EndpointView {
  return {
    id: input.id ?? input.ID ?? 0,
    name: input.name ?? input.Name ?? '',
    token: input.token ?? input.Token ?? '',
    meow_nickname: input.meow_nickname ?? input.MeowNickname ?? '',
    default_title: input.default_title ?? input.DefaultTitle ?? '',
    msg_type: input.msg_type ?? input.MsgType ?? 'text',
    html_height: input.html_height ?? input.HTMLHeight ?? 200,
    default_url: input.default_url ?? input.DefaultURL ?? '',
    default_img_url: input.default_img_url ?? input.DefaultImgURL ?? '',
    active: input.active ?? input.Active ?? false,
    created_at: input.created_at ?? input.CreatedAt ?? '',
    updated_at: input.updated_at ?? input.UpdatedAt ?? '',
  }
}

export function normalizePushLog(input: PushLog): Required<PushLogListItem> & {
  token: string
  request_method: string
  request_headers: string
  request_query: string
  request_payload: string
  meow_response_body: string
} {
  return {
    id: input.id ?? input.ID ?? 0,
    endpoint_id: input.endpoint_id ?? input.EndpointID ?? 0,
    endpoint_name: input.endpoint_name ?? input.EndpointName ?? '',
    token: input.token ?? input.Token ?? '',
    source_type: input.source_type ?? input.SourceType ?? '',
    request_method: input.request_method ?? input.RequestMethod ?? '',
    request_headers: input.request_headers ?? input.RequestHeaders ?? '',
    request_query: input.request_query ?? input.RequestQuery ?? '',
    request_payload: input.request_payload ?? input.RequestPayload ?? '',
    parsed_title: input.parsed_title ?? input.ParsedTitle ?? '',
    parsed_msg: input.parsed_msg ?? input.ParsedMsg ?? '',
    parsed_msg_type: input.parsed_msg_type ?? input.ParsedMsgType ?? '',
    meow_status_code: input.meow_status_code ?? input.MeowStatusCode ?? 0,
    meow_response_body: input.meow_response_body ?? input.MeowResponseBody ?? '',
    success: input.success ?? input.Success ?? false,
    error_message: input.error_message ?? input.ErrorMessage ?? '',
    created_at: input.created_at ?? input.CreatedAt ?? '',
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  const token = tokenProvider()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  let response: Response
  try {
    response = await fetch(path, { ...init, headers })
  } catch {
    throw new ApiError(0, '服务不可达，请检查 meowbridge 是否正在运行')
  }

  let payload: SuccessResponse<T> | ErrorResponse
  try {
    payload = (await response.json()) as SuccessResponse<T> | ErrorResponse
  } catch {
    payload = { ok: false, error: '服务返回了无法解析的响应' }
  }

  if (response.status === 401) {
    unauthorizedHandler()
  }

  if (!response.ok || !payload.ok) {
    throw new ApiError(response.status, 'error' in payload ? payload.error : '请求失败')
  }

  return payload.data
}

export const apiClient = {
  async login(password: string): Promise<string> {
    const data = await request<{ token: string }>('/api/admin/login', {
      method: 'POST',
      body: JSON.stringify({ password }),
    })
    return data.token
  },
  async listEndpoints(): Promise<EndpointView[]> {
    const data = await request<Endpoint[]>('/api/admin/endpoints')
    return data.map(normalizeEndpoint)
  },
  async getEndpoint(id: number): Promise<EndpointView> {
    return normalizeEndpoint(await request<Endpoint>(`/api/admin/endpoints/${id}`))
  },
  async createEndpoint(input: EndpointInput): Promise<EndpointView> {
    return normalizeEndpoint(
      await request<Endpoint>('/api/admin/endpoints', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
    )
  },
  async updateEndpoint(id: number, input: EndpointUpdate): Promise<EndpointView> {
    return normalizeEndpoint(
      await request<Endpoint>(`/api/admin/endpoints/${id}`, {
        method: 'PUT',
        body: JSON.stringify(input),
      }),
    )
  },
  async deleteEndpoint(id: number): Promise<{ deleted: boolean }> {
    return request<{ deleted: boolean }>(`/api/admin/endpoints/${id}`, { method: 'DELETE' })
  },
  async resetEndpointToken(id: number): Promise<EndpointView> {
    return normalizeEndpoint(
      await request<Endpoint>(`/api/admin/endpoints/${id}/reset-token`, { method: 'POST' }),
    )
  },
  async setEndpointActive(id: number, active: boolean): Promise<{ active: boolean }> {
    return request<{ active: boolean }>(`/api/admin/endpoints/${id}/active`, {
      method: 'PATCH',
      body: JSON.stringify({ active }),
    })
  },
  async listPushLogs(): Promise<PushLogListItem[]> {
    return request<PushLogListItem[]>('/api/admin/push-logs')
  },
  async getPushLog(id: number): Promise<ReturnType<typeof normalizePushLog>> {
    return normalizePushLog(await request<PushLog>(`/api/admin/push-logs/${id}`))
  },
  async cleanupPushLogs(): Promise<{ deleted: number }> {
    return request<{ deleted: number }>('/api/admin/push-logs', { method: 'DELETE' })
  },
  async getSettings(): Promise<Record<string, string>> {
    return request<Record<string, string>>('/api/admin/settings')
  },
  async updateSettings(values: Record<string, string>): Promise<{ updated: boolean }> {
    return request<{ updated: boolean }>('/api/admin/settings', {
      method: 'PUT',
      body: JSON.stringify(values),
    })
  },
  async changePassword(oldPassword: string, newPassword: string): Promise<{ changed: boolean }> {
    return request<{ changed: boolean }>('/api/admin/change-password', {
      method: 'POST',
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    })
  },
}
