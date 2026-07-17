package tgproxy

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestConvertMarkdownV2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"*bold*", "**bold**"},
		{"_italic_", "*italic*"},
		{"__underline__", "<u>underline</u>"},
		{"~strike~", "~~strike~~"},
	}

	for _, tt := range tests {
		result, msgType := ConvertTGFormat(tt.input, "MarkdownV2")
		if result != tt.expected {
			t.Errorf("ConvertMarkdownV2(%q) = %q, want %q", tt.input, result, tt.expected)
		}
		if msgType != "markdown" {
			t.Errorf("msgType = %q, want markdown", msgType)
		}
	}
}

func TestConvertHTML(t *testing.T) {
	input := `<b>bold</b> <tg-emoji>emoji</tg-emoji>`
	expected := `<b>bold</b> `
	result, msgType := ConvertTGFormat(input, "HTML")
	if result != expected {
		t.Errorf("ConvertHTML(%q) = %q, want %q", input, result, expected)
	}
	if msgType != "html" {
		t.Errorf("msgType = %q, want html", msgType)
	}
}

func TestRespondTGSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	RespondTGSuccess(w, "test message")

	var resp TGSuccessResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.OK {
		t.Error("expected ok=true")
	}
}

func TestRespondTGError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondTGError(w, 401, "Unauthorized")

	var resp TGErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.OK {
		t.Error("expected ok=false")
	}
	if resp.ErrorCode != 401 {
		t.Errorf("error_code = %d, want 401", resp.ErrorCode)
	}
}

func TestParseTGRequest(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		method   string
		wantText string
		wantMode string
	}{
		{
			name:     "sendMessage with MarkdownV2",
			input:    `{"text": "*Hello* _World_", "parse_mode": "MarkdownV2"}`,
			method:   "sendMessage",
			wantText: "*Hello* _World_",
			wantMode: "MarkdownV2",
		},
		{
			name:     "sendPhoto with caption",
			input:    `{"caption": "photo caption", "parse_mode": "HTML"}`,
			method:   "sendPhoto",
			wantText: "photo caption",
			wantMode: "HTML",
		},
		{
			name:     "sendMessage without parse_mode",
			input:    `{"text": "plain text"}`,
			method:   "sendMessage",
			wantText: "plain text",
			wantMode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseTGRequest([]byte(tt.input), tt.method)
			if err != nil {
				t.Fatalf("ParseTGRequest() error = %v", err)
			}
			if msg.GetContent() != tt.wantText {
				t.Errorf("GetContent() = %q, want %q", msg.GetContent(), tt.wantText)
			}
			if msg.ParseMode != tt.wantMode {
				t.Errorf("ParseMode = %q, want %q", msg.ParseMode, tt.wantMode)
			}
		})
	}
}

func TestParseTGRequestInvalidJSON(t *testing.T) {
	_, err := ParseTGRequest([]byte("invalid json"), "sendMessage")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
