export type MsgType = 'text' | 'html' | 'markdown'

export interface Endpoint {
  ID?: number
  id?: number
  Name?: string
  name?: string
  Token?: string
  token?: string
  MeowNickname?: string
  meow_nickname?: string
  DefaultTitle?: string
  default_title?: string
  MsgType?: MsgType
  msg_type?: MsgType
  HTMLHeight?: number
  html_height?: number
  DefaultURL?: string
  default_url?: string
  DefaultImgURL?: string
  default_img_url?: string
  Active?: boolean
  active?: boolean
  CreatedAt?: string
  created_at?: string
  UpdatedAt?: string
  updated_at?: string
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
  ID?: number
  id?: number
  EndpointID?: number
  endpoint_id?: number
  EndpointName?: string
  endpoint_name?: string
  Token?: string
  token?: string
  SourceType?: string
  source_type?: string
  RequestMethod?: string
  request_method?: string
  RequestHeaders?: string
  request_headers?: string
  RequestQuery?: string
  request_query?: string
  RequestPayload?: string
  request_payload?: string
  ParsedTitle?: string
  parsed_title?: string
  ParsedMsg?: string
  parsed_msg?: string
  ParsedMsgType?: string
  parsed_msg_type?: string
  MeowStatusCode?: number
  meow_status_code?: number
  MeowResponseBody?: string
  meow_response_body?: string
  Success?: boolean
  success?: boolean
  ErrorMessage?: string
  error_message?: string
  CreatedAt?: string
  created_at?: string
}
