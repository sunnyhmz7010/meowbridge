package webhook

import "net/http"

type ParseInput struct {
	Headers http.Header
	Body    []byte
}

type ParsedMessage struct {
	SourceType string
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
}

type EndpointDefaults struct {
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
}

type QueryOverrides struct {
	Title      string
	MsgType    string
	HTMLHeight int
	URL        string
	ImgURL     string
}

type FinalMessage struct {
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
	HTMLHeight int
}
