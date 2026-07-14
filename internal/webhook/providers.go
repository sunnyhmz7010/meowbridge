package webhook

import (
	"encoding/json"
	"strings"
)

func parseGitHubPR(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if input.Headers.Get("X-GitHub-Event") != "pull_request" && payload["pull_request"] == nil {
		return ParsedMessage{}, false
	}
	pr, _ := payload["pull_request"].(map[string]any)
	return ParsedMessage{
		SourceType: "github_pr",
		Title:      stringValue(pr, "title"),
		Msg:        firstNonEmpty(stringValue(pr, "body"), stringValue(payload, "action")),
		URL:        stringValue(pr, "html_url"),
		MsgType:    "markdown",
	}, true
}

func parseGitHubAction(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	run, ok := payload["workflow_run"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	commit, _ := run["head_commit"].(map[string]any)
	return ParsedMessage{
		SourceType: "github_action",
		Title:      firstNonEmpty(stringValue(run, "event"), "GitHub Actions"),
		Msg:        firstNonEmpty(stringValue(commit, "message"), stringValue(run, "name")),
		URL:        stringValue(run, "html_url"),
		MsgType:    "markdown",
	}, true
}

func parseGitHub(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	repo, ok := payload["repository"].(map[string]any)
	if !ok && input.Headers.Get("X-GitHub-Event") == "" {
		return ParsedMessage{}, false
	}
	return ParsedMessage{
		SourceType: "github",
		Title:      firstNonEmpty(stringValue(payload, "action"), input.Headers.Get("X-GitHub-Event"), "GitHub Webhook"),
		Msg:        firstNonEmpty(stringValue(repo, "full_name"), compactJSON(payload)),
		URL:        stringValue(repo, "html_url"),
		MsgType:    "text",
	}, true
}

func parseJenkins(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	project, ok := payload["project"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	build, _ := payload["build"].(map[string]any)
	return ParsedMessage{SourceType: "jenkins", Title: stringValue(project, "name"), Msg: firstNonEmpty(stringValue(build, "full_display_url"), compactJSON(payload)), URL: stringValue(build, "full_display_url"), MsgType: "text"}, true
}

func parseGrafana(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if payload["alerts"] == nil || payload["receiver"] != nil {
		return ParsedMessage{}, false
	}
	return parseAlertPayload("grafana", payload, "message")
}

func parsePrometheus(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if payload["alerts"] == nil || payload["receiver"] == nil {
		return ParsedMessage{}, false
	}
	return parseAlertPayload("prometheus", payload, "description")
}

func parseAlertPayload(source string, payload map[string]any, annotationKey string) (ParsedMessage, bool) {
	alert := firstAlert(payload)
	labels, _ := alert["labels"].(map[string]any)
	annotations, _ := alert["annotations"].(map[string]any)
	return ParsedMessage{SourceType: source, Title: stringValue(labels, "alertname"), Msg: firstNonEmpty(stringValue(annotations, annotationKey), compactJSON(payload)), URL: stringValue(payload, "externalURL"), MsgType: "markdown"}, true
}

func parseZabbix(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	trigger, ok := payload["trigger"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	event, _ := payload["event"].(map[string]any)
	return ParsedMessage{SourceType: "zabbix", Title: stringValue(trigger, "description"), Msg: firstNonEmpty(stringValue(event, "description"), compactJSON(payload)), MsgType: "markdown"}, true
}

func parseGotify(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if _, ok := payload["message"]; !ok {
		return ParsedMessage{}, false
	}
	if payload["priority"] == nil && payload["extras"] == nil {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "gotify", Title: stringValue(payload, "title"), Msg: stringValue(payload, "message"), MsgType: "markdown"}, true
}

func parseEmby(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if _, ok := payload["Title"]; !ok {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "emby", Title: stringValue(payload, "Title"), Msg: firstNonEmpty(stringValue(payload, "Description"), compactJSON(payload)), MsgType: "text"}, true
}

func parseGeneric(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	msg := firstNonEmpty(stringValue(payload, "msg"), stringValue(payload, "message"), stringValue(payload, "text"), stringValue(payload, "content"))
	if msg == "" {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "generic", Title: stringValue(payload, "title"), Msg: msg, URL: stringValue(payload, "url"), ImgURL: stringValue(payload, "imgUrl"), MsgType: stringValue(payload, "msgType")}, true
}

func parseFallback(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	return ParsedMessage{SourceType: "fallback", Title: "Webhook", Msg: prettyJSON(payload), MsgType: "markdown"}, true
}

func firstAlert(payload map[string]any) map[string]any {
	alerts, _ := payload["alerts"].([]any)
	if len(alerts) == 0 {
		return map[string]any{}
	}
	first, _ := alerts[0].(map[string]any)
	return first
}

func stringValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	value, _ := m[key].(string)
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func compactJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func prettyJSON(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return string(data)
}
