package webhook

import (
	"encoding/json"
	"errors"
)

type parser func(ParseInput, map[string]any) (ParsedMessage, bool)

func Parse(input ParseInput) (ParsedMessage, error) {
	var payload map[string]any
	if err := json.Unmarshal(input.Body, &payload); err != nil {
		return ParsedMessage{}, errors.New("invalid json payload")
	}
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
		parseFallback,
	} {
		if parsed, ok := parse(input, payload); ok {
			return parsed, nil
		}
	}
	return ParsedMessage{}, errors.New("payload could not be parsed")
}
