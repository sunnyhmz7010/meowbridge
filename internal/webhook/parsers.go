package webhook

import (
	"encoding/json"
	"errors"
	"strings"
)

type parser func(ParseInput, map[string]any) (ParsedMessage, bool)

func Parse(input ParseInput) (ParsedMessage, error) {
	var payload any
	if err := json.Unmarshal(input.Body, &payload); err != nil {
		return ParsedMessage{}, errors.New("invalid json payload")
	}
	if object, ok := payload.(map[string]any); ok {
		for _, parse := range []parser{
			parseGitHubPR,
			parseGitHubAction,
			parseGitHub,
			parseJenkins,
			parseGrafana,
			parsePrometheus,
			parseZabbix,
			parseGotify,
			parseEmby,
			parseGeneric,
		} {
			if parsed, matched := parse(input, object); matched && strings.TrimSpace(parsed.Msg) != "" {
				return parsed, nil
			}
		}
	}
	parsed, _ := parseFallback(input, payload)
	return parsed, nil
}
