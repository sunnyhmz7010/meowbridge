package webhook

import (
	"errors"
	"strings"
)

func Merge(parsed ParsedMessage, endpoint EndpointDefaults, query QueryOverrides) (FinalMessage, error) {
	msg := strings.TrimSpace(parsed.Msg)
	if msg == "" {
		return FinalMessage{}, errors.New("message is required")
	}
	final := FinalMessage{
		Title:      firstNonEmpty(query.Title, parsed.Title, endpoint.DefaultTitle, "Meow"),
		Msg:        msg,
		URL:        firstNonEmpty(query.URL, parsed.URL, endpoint.DefaultURL),
		ImgURL:     firstNonEmpty(query.ImgURL, parsed.ImgURL, endpoint.DefaultImgURL),
		MsgType:    firstNonEmpty(query.MsgType, parsed.MsgType, endpoint.MsgType, "text"),
		HTMLHeight: firstPositive(query.HTMLHeight, endpoint.HTMLHeight, 200),
	}
	return final, nil
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
