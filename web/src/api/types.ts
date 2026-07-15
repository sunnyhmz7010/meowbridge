export type MsgType = 'text' | 'html' | 'markdown'

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
  active: boolean
}

export interface EndpointUpdate {
  name: string
  default_title: string
  msg_type: MsgType
  html_height: number
  default_url: string
  default_img_url: string
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
