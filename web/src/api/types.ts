export type MsgType = 'text' | 'html' | 'markdown'

export type ParserMode = 'auto' | 'preset' | 'custom'

export interface ParserConfig {
  mode: ParserMode
  preset: string
  field_mapping: Record<string, string[]>
  default_values: Record<string, string>
}

export interface WebhookPreset {
  id: string
  name: string
  description: string
  field_mapping: Record<string, string[]>
  default_values: Record<string, string>
}

export interface WebhookPreviewRequest {
  parser_config: ParserConfig
  payload: unknown
}

export interface WebhookPreviewResult {
  source_type: string
  title: string
  msg: string
  url: string
  img_url: string
  msg_type: MsgType
}

export interface SetupStatus {
  needs_setup: boolean
}

export interface EndpointView {
  id: number
  name: string
  token: string
  meow_nickname: string
  default_title: string
  msg_type: MsgType
  html_height: number
  default_url: string
  default_img_url: string
  parser_config: ParserConfig
  active: boolean
  created_at: string
  updated_at: string
}

export interface EndpointInput {
  name: string
  meow_nickname: string
  default_title: string
  msg_type: MsgType
  html_height: number
  default_url: string
  default_img_url: string
  parser_config: ParserConfig
  active: boolean
}

export interface EndpointUpdate {
  name: string
  default_title: string
  msg_type: MsgType
  html_height: number
  default_url: string
  default_img_url: string
  parser_config: ParserConfig
  active?: boolean
}

export interface PushLogListItem {
  id: number
  endpoint_id: number
  endpoint_name: string
  source_type: string
  parsed_title: string
  parsed_msg: string
  parsed_msg_type: string
  meow_status_code: number
  success: boolean
  error_message: string
  created_at: string
}

export interface PushLog {
  id?: number
  endpoint_id?: number
  endpoint_name?: string
  token?: string
  source_type?: string
  request_method?: string
  request_headers?: string
  request_query?: string
  request_payload?: string
  parsed_title?: string
  parsed_msg?: string
  parsed_msg_type?: string
  meow_status_code?: number
  meow_response_body?: string
  success?: boolean
  error_message?: string
  created_at?: string
}
