package webhook

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestParseGitHubPullRequest(t *testing.T) {
	payload := []byte(`{"action":"opened","repository":{"full_name":"sunny/meowbridge","html_url":"https://github.com/sunny/meowbridge"},"pull_request":{"title":"Add webhook","body":"Adds support","html_url":"https://github.com/sunny/meowbridge/pull/1"}}`)
	parsed, err := Parse(ParseInput{
		Headers: http.Header{"X-GitHub-Event": []string{"pull_request"}},
		Body:    payload,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SourceType != "github_pr" || parsed.Title != "Add webhook" || parsed.Msg != "Adds support" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseTopLevelArrayFallsBackToFullJSON(t *testing.T) {
	parsed, err := Parse(ParseInput{Headers: http.Header{}, Body: []byte(`[{"event":"started"},{"event":"finished"}]`)})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SourceType != "fallback" || parsed.Msg == "" {
		t.Fatalf("parsed = %#v", parsed)
	}
	var payload []map[string]string
	if err := json.Unmarshal([]byte(parsed.Msg), &payload); err != nil {
		t.Fatalf("fallback message is not full JSON: %v", err)
	}
	if len(payload) != 2 || payload[1]["event"] != "finished" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestParseProviderWithEmptyMessageFallsBackToFullJSON(t *testing.T) {
	const body = `{"title":"Gotify title","message":"","priority":5}`
	parsed, err := Parse(ParseInput{Headers: http.Header{}, Body: []byte(body)})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SourceType != "fallback" || parsed.Msg == "" {
		t.Fatalf("parsed = %#v", parsed)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(parsed.Msg), &payload); err != nil {
		t.Fatalf("fallback message is not full JSON: %v", err)
	}
	if payload["title"] != "Gotify title" || payload["priority"] != float64(5) {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestParseKnownProvidersAndFallback(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantSource string
	}{
		{"github_action", `{"workflow_run":{"event":"push","head_commit":{"message":"build passed"},"html_url":"https://github.test/run"}}`, "github_action"},
		{"github", `{"action":"push","repository":{"full_name":"sunny/meowbridge","html_url":"https://github.test/repo"}}`, "github"},
		{"jenkins", `{"project":{"name":"build"},"build":{"full_display_url":"https://jenkins.test/1"}}`, "jenkins"},
		{"grafana", `{"alerts":[{"labels":{"alertname":"CPUHigh"},"annotations":{"message":"CPU high"}}],"externalURL":"https://grafana.test"}`, "grafana"},
		{"prometheus", `{"receiver":"default","alerts":[{"labels":{"alertname":"DiskFull"},"annotations":{"description":"disk full"}}]}`, "prometheus"},
		{"zabbix", `{"trigger":{"description":"Host down"},"event":{"description":"host unavailable"}}`, "zabbix"},
		{"gotify", `{"title":"Gotify title","message":"Gotify message","priority":5}`, "gotify"},
		{"emby", `{"Title":"Playback started","Description":"Movie"}`, "emby"},
		{"generic", `{"title":"Generic title","message":"Generic message"}`, "generic"},
		{"fallback", `{"unexpected":{"nested":true}}`, "fallback"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := Parse(ParseInput{Headers: http.Header{}, Body: []byte(tc.body)})
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if parsed.SourceType != tc.wantSource {
				t.Fatalf("SourceType = %q", parsed.SourceType)
			}
			if parsed.Msg == "" {
				t.Fatalf("Msg is empty: %#v", parsed)
			}
		})
	}
}
