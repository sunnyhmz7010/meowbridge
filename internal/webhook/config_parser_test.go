package webhook

import (
	"net/http"
	"strings"
	"testing"
)

func TestParseWithConfigUsesCustomFieldMapping(t *testing.T) {
	body := []byte(`{"title":"Deploy","repository":{"full_name":"sunny/meowbridge"},"ref":"refs/heads/main","url":"https://github.com/sunnyhmz7010/meowbridge"}`)

	parsed, matched, err := ParseWithConfig(ParseInput{Headers: http.Header{}, Body: body}, ParserConfig{
		Mode: "custom",
		FieldMapping: map[string][]string{
			"title":    {"GitHub: ", "$.title"},
			"msg":      {"仓库: ", "$.repository.full_name", "\\n分支: ", "$.ref"},
			"url":      {"$.url"},
			"msg_type": {"markdown"},
		},
	})
	if err != nil {
		t.Fatalf("ParseWithConfig: %v", err)
	}
	if !matched {
		t.Fatalf("ParseWithConfig did not match")
	}
	if parsed.SourceType != "custom" {
		t.Fatalf("SourceType = %q", parsed.SourceType)
	}
	if parsed.Title != "GitHub: Deploy" {
		t.Fatalf("Title = %q", parsed.Title)
	}
	if parsed.Msg != "仓库: sunny/meowbridge\n分支: refs/heads/main" {
		t.Fatalf("Msg = %q", parsed.Msg)
	}
	if parsed.URL != "https://github.com/sunnyhmz7010/meowbridge" || parsed.MsgType != "markdown" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseWithConfigUsesGithubPushMinimalPreset(t *testing.T) {
	body := []byte(`{"sourcecontrol":"github","service":"github","event_type":"push","hook":{"url":"https://github.com/sunnyhmz7010/meowbridge"},"ref":"refs/heads/main"}`)

	parsed, matched, err := ParseWithConfig(ParseInput{Headers: http.Header{}, Body: body}, ParserConfig{
		Mode:   "preset",
		Preset: "github_push_minimal",
	})
	if err != nil {
		t.Fatalf("ParseWithConfig: %v", err)
	}
	if !matched {
		t.Fatalf("ParseWithConfig did not match")
	}
	if parsed.SourceType != "github_push_minimal" {
		t.Fatalf("SourceType = %q", parsed.SourceType)
	}
	if parsed.Title != "GitHub Push" {
		t.Fatalf("Title = %q", parsed.Title)
	}
	for _, want := range []string{
		"仓库: https://github.com/sunnyhmz7010/meowbridge",
		"分支: main",
		"事件: push",
		"来源: github",
	} {
		if !strings.Contains(parsed.Msg, want) {
			t.Fatalf("Msg %q does not contain %q", parsed.Msg, want)
		}
	}
	if parsed.URL != "https://github.com/sunnyhmz7010/meowbridge" || parsed.MsgType != "markdown" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseWithConfigUsesSourceControlWhenServiceIsMissing(t *testing.T) {
	body := []byte(`{"sourcecontrol":"github","event_type":"push","hook":{"url":"https://github.com/sunnyhmz7010/meowbridge"},"ref":"refs/heads/main"}`)

	parsed, matched, err := ParseWithConfig(ParseInput{Headers: http.Header{}, Body: body}, ParserConfig{
		Mode:   "preset",
		Preset: "github_push_minimal",
	})
	if err != nil {
		t.Fatalf("ParseWithConfig: %v", err)
	}
	if !matched {
		t.Fatalf("ParseWithConfig did not match")
	}
	if !strings.Contains(parsed.Msg, "来源: github") {
		t.Fatalf("Msg %q does not contain sourcecontrol value", parsed.Msg)
	}
}

func TestParseWithConfigReturnsNoMatchForDisabledOrInvalidConfig(t *testing.T) {
	body := []byte(`{"message":"hello"}`)

	for _, config := range []ParserConfig{
		{},
		{Mode: "auto"},
		{Mode: "preset", Preset: "missing"},
		{Mode: "custom", FieldMapping: map[string][]string{"title": {"$.message"}}},
	} {
		parsed, matched, err := ParseWithConfig(ParseInput{Headers: http.Header{}, Body: body}, config)
		if err != nil {
			t.Fatalf("ParseWithConfig(%#v): %v", config, err)
		}
		if matched {
			t.Fatalf("ParseWithConfig(%#v) matched unexpectedly: %#v", config, parsed)
		}
	}
}
