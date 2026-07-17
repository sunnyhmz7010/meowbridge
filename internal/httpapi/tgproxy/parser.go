package tgproxy

import "encoding/json"

type TGMessage struct {
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func ParseTGRequest(body []byte, method string) (*TGMessage, error) {
	var msg TGMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (m *TGMessage) GetContent() string {
	if m.Text != "" {
		return m.Text
	}
	return m.Caption
}
