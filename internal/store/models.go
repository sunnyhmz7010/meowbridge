package store

import "time"

type Endpoint struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Token         string    `json:"token"`
	MeowNickname  string    `json:"meow_nickname"`
	DefaultTitle  string    `json:"default_title"`
	MsgType       string    `json:"msg_type"`
	HTMLHeight    int       `json:"html_height"`
	DefaultURL    string    `json:"default_url"`
	DefaultImgURL string    `json:"default_img_url"`
	ParserConfig  string    `json:"parser_config"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PushLog struct {
	ID               int64     `json:"id"`
	EndpointID       int64     `json:"endpoint_id"`
	EndpointName     string    `json:"endpoint_name"`
	Token            string    `json:"token"`
	SourceType       string    `json:"source_type"`
	RequestMethod    string    `json:"request_method"`
	RequestHeaders   string    `json:"request_headers"`
	RequestQuery     string    `json:"request_query"`
	RequestPayload   string    `json:"request_payload"`
	ParsedTitle      string    `json:"parsed_title"`
	ParsedMsg        string    `json:"parsed_msg"`
	ParsedMsgType    string    `json:"parsed_msg_type"`
	MeowStatusCode   int       `json:"meow_status_code"`
	MeowResponseBody string    `json:"meow_response_body"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"`
}
