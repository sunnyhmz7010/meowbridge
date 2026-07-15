package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var parserPresets = []ParserPreset{
	{
		ID:          "github_push_minimal",
		Name:        "GitHub 简化 Push",
		Description: "适配仅包含 event_type、hook.url、ref、service/sourcecontrol 的 GitHub Push payload",
		FieldMapping: map[string][]string{
			"title":    {"GitHub Push"},
			"msg":      {"仓库: ", "$.hook.url", "\\n分支: ", "$.ref", "\\n事件: ", "$.event_type", "\\n来源: ", "$.service"},
			"url":      {"$.hook.url"},
			"msg_type": {"markdown"},
		},
		DefaultValues: map[string]string{"msg_type": "markdown"},
	},
	{
		ID:          "github",
		Name:        "GitHub Webhook",
		Description: "标准 GitHub Webhook",
		FieldMapping: map[string][]string{
			"title":    {"GitHub: ", "$.action"},
			"msg":      {"仓库: ", "$.repository.full_name", "\\n引用: ", "$.ref", "\\n提交: ", "$.head_commit.message"},
			"url":      {"$.repository.html_url"},
			"msg_type": {"markdown"},
		},
		DefaultValues: map[string]string{"title": "GitHub Webhook", "msg_type": "markdown"},
	},
	{
		ID:          "github_action",
		Name:        "GitHub Actions",
		Description: "GitHub workflow_run 事件",
		FieldMapping: map[string][]string{
			"title":    {"GitHub Actions: ", "$.workflow_run.event"},
			"msg":      {"工作流: ", "$.workflow_run.name", "\\n提交: ", "$.workflow_run.head_commit.message"},
			"url":      {"$.workflow_run.html_url"},
			"msg_type": {"markdown"},
		},
		DefaultValues: map[string]string{"title": "GitHub Actions", "msg_type": "markdown"},
	},
	{
		ID:          "grafana",
		Name:        "Grafana",
		Description: "Grafana 告警消息",
		FieldMapping: map[string][]string{
			"title":    {"$.alerts[0].labels.alertname"},
			"msg":      {"$.alerts[0].annotations.message", "\\n状态: ", "$.status"},
			"url":      {"$.externalURL"},
			"msg_type": {"markdown"},
		},
		DefaultValues: map[string]string{"title": "Grafana Alert", "msg_type": "markdown"},
	},
	{
		ID:          "prometheus",
		Name:        "Prometheus Alertmanager",
		Description: "Prometheus Alertmanager 告警",
		FieldMapping: map[string][]string{
			"title":    {"$.alerts[0].labels.alertname"},
			"msg":      {"$.alerts[0].annotations.description", "\\n状态: ", "$.status"},
			"url":      {"$.externalURL"},
			"msg_type": {"markdown"},
		},
		DefaultValues: map[string]string{"title": "Prometheus Alert", "msg_type": "markdown"},
	},
	{
		ID:          "emby",
		Name:        "Emby",
		Description: "Emby 媒体库通知",
		FieldMapping: map[string][]string{
			"title":    {"$.Title"},
			"msg":      {"$.Description", "\\n事件: ", "$.Event", "\\n用户: ", "$.User.Name"},
			"msg_type": {"text"},
		},
		DefaultValues: map[string]string{"title": "Emby", "msg_type": "text"},
	},
	{
		ID:            "generic",
		Name:          "通用自定义",
		Description:   "自定义字段映射",
		FieldMapping:  map[string][]string{},
		DefaultValues: map[string]string{"title": "Webhook", "msg_type": "text"},
	},
}

func ParserPresets() []ParserPreset {
	presets := make([]ParserPreset, len(parserPresets))
	copy(presets, parserPresets)
	return presets
}

func ParseWithConfig(input ParseInput, config ParserConfig) (ParsedMessage, bool, error) {
	mode := strings.TrimSpace(config.Mode)
	if mode == "" || mode == "auto" {
		return ParsedMessage{}, false, nil
	}

	var payload any
	if err := json.Unmarshal(input.Body, &payload); err != nil {
		return ParsedMessage{}, false, errors.New("invalid json payload")
	}

	switch mode {
	case "preset":
		preset, ok := findParserPreset(config.Preset)
		if !ok {
			return ParsedMessage{}, false, nil
		}
		merged := ParserConfig{
			Mode:          "custom",
			Preset:        preset.ID,
			FieldMapping:  mergeFieldMapping(preset.FieldMapping, config.FieldMapping),
			DefaultValues: mergeDefaultValues(preset.DefaultValues, config.DefaultValues),
		}
		parsed, matched, err := parseMappedPayload(payload, merged, preset.ID)
		if !matched || err != nil {
			return parsed, matched, err
		}
		if preset.ID == "github_push_minimal" {
			parsed.Msg = renderGithubPushMinimalMessage(payload)
		}
		return parsed, true, nil
	case "custom":
		return parseMappedPayload(payload, config, "custom")
	default:
		return ParsedMessage{}, false, nil
	}
}

func findParserPreset(id string) (ParserPreset, bool) {
	for _, preset := range parserPresets {
		if preset.ID == id {
			return preset, true
		}
	}
	return ParserPreset{}, false
}

func NormalizeParserConfig(raw []byte) (string, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var config ParserConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return "", fmt.Errorf("invalid parser_config: %w", err)
	}

	config.Mode = strings.TrimSpace(config.Mode)
	config.Preset = strings.TrimSpace(config.Preset)
	if config.FieldMapping == nil {
		config.FieldMapping = map[string][]string{}
	}
	if config.DefaultValues == nil {
		config.DefaultValues = map[string]string{}
	}

	switch config.Mode {
	case "", "auto":
		return "", nil
	case "preset":
		if config.Preset == "" {
			return "", errors.New("invalid parser_config: preset is required")
		}
		if _, ok := findParserPreset(config.Preset); !ok {
			return "", errors.New("invalid parser_config: unknown preset")
		}
	case "custom":
		if len(config.FieldMapping["msg"]) == 0 {
			return "", errors.New("invalid parser_config: custom msg mapping is required")
		}
	default:
		return "", errors.New("invalid parser_config: unsupported mode")
	}

	normalized, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("invalid parser_config: %w", err)
	}
	return string(normalized), nil
}

func renderGithubPushMinimalMessage(payload any) string {
	repo, _ := ParseJSONPath(payload, "$.hook.url")
	ref, _ := ParseJSONPath(payload, "$.ref")
	event, _ := ParseJSONPath(payload, "$.event_type")
	source, _ := firstJSONPathValue(payload, "$.service", "$.sourcecontrol")
	branch := strings.TrimPrefix(ref, "refs/heads/")
	return strings.TrimSpace(fmt.Sprintf("仓库: %s\n分支: %s\n事件: %s\n来源: %s", repo, branch, event, source))
}

func firstJSONPathValue(payload any, paths ...string) (string, bool) {
	for _, path := range paths {
		value, ok := ParseJSONPath(payload, path)
		if ok && strings.TrimSpace(value) != "" {
			return value, true
		}
	}
	return "", false
}

func parseMappedPayload(payload any, config ParserConfig, sourceType string) (ParsedMessage, bool, error) {
	values := map[string]string{}
	for field, fragments := range config.FieldMapping {
		if text := renderFragments(payload, fragments); text != "" {
			values[field] = text
		}
	}
	for field, value := range config.DefaultValues {
		if strings.TrimSpace(values[field]) == "" {
			values[field] = value
		}
	}

	msg := strings.TrimSpace(values["msg"])
	if msg == "" {
		return ParsedMessage{}, false, nil
	}
	return ParsedMessage{
		SourceType: sourceType,
		Title:      strings.TrimSpace(values["title"]),
		Msg:        msg,
		URL:        strings.TrimSpace(values["url"]),
		ImgURL:     strings.TrimSpace(values["img_url"]),
		MsgType:    strings.TrimSpace(values["msg_type"]),
	}, true, nil
}

func renderFragments(payload any, fragments []string) string {
	var builder strings.Builder
	for _, fragment := range fragments {
		switch {
		case strings.HasPrefix(fragment, "$."):
			value, ok := ParseJSONPath(payload, fragment)
			if ok {
				builder.WriteString(value)
			}
		case strings.HasPrefix(fragment, "$["):
			value, ok := ParseJSONPath(payload, fragment)
			if ok {
				builder.WriteString(value)
			}
		default:
			builder.WriteString(unescapeMappingLiteral(fragment))
		}
	}
	return strings.TrimSpace(builder.String())
}

func unescapeMappingLiteral(value string) string {
	return strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\\`, `\`).Replace(value)
}

func mergeFieldMapping(base, override map[string][]string) map[string][]string {
	merged := map[string][]string{}
	for key, value := range base {
		merged[key] = append([]string(nil), value...)
	}
	for key, value := range override {
		if len(value) > 0 {
			merged[key] = append([]string(nil), value...)
		}
	}
	return merged
}

func mergeDefaultValues(base, override map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range override {
		if strings.TrimSpace(value) != "" {
			merged[key] = value
		}
	}
	return merged
}
