package webhook

import "net/http"

type ParseInput struct {
	Headers http.Header
	Body    []byte
}

type ParserConfig struct {
	Mode          string              `json:"mode"`
	Preset        string              `json:"preset"`
	FieldMapping  map[string][]string `json:"field_mapping"`
	DefaultValues map[string]string   `json:"default_values"`
}

type ParserPreset struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	FieldMapping  map[string][]string `json:"field_mapping"`
	DefaultValues map[string]string   `json:"default_values"`
}

type ParsedMessage struct {
	SourceType string `json:"source_type"`
	Title      string `json:"title"`
	Msg        string `json:"msg"`
	URL        string `json:"url"`
	ImgURL     string `json:"img_url"`
	MsgType    string `json:"msg_type"`
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
