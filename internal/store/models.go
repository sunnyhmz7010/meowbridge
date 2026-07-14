package store

import "time"

type Endpoint struct {
	ID            int64
	Name          string
	Token         string
	MeowNickname  string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Setting struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

type PushLog struct {
	ID               int64
	EndpointID       int64
	EndpointName     string
	Token            string
	SourceType       string
	RequestMethod    string
	RequestHeaders   string
	RequestQuery     string
	RequestPayload   string
	ParsedTitle      string
	ParsedMsg        string
	ParsedMsgType    string
	MeowStatusCode   int
	MeowResponseBody string
	Success          bool
	ErrorMessage     string
	CreatedAt        time.Time
}
