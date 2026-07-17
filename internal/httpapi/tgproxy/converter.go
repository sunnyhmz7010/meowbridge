package tgproxy

import (
	"regexp"
	"strings"
)

func ConvertTGFormat(content, parseMode string) (string, string) {
	switch parseMode {
	case "MarkdownV2":
		return convertMarkdownV2(content), "markdown"
	case "HTML":
		return convertHTML(content), "html"
	default:
		return content, "text"
	}
}

func convertMarkdownV2(content string) string {
	// 先反转义 TG 特殊字符
	content = strings.ReplaceAll(content, `\\`, `\x00`)
	content = strings.ReplaceAll(content, `\*`, `\x01`)
	content = strings.ReplaceAll(content, `\_`, `\x02`)
	content = strings.ReplaceAll(content, `\~`, `\x03`)
	content = strings.ReplaceAll(content, "\x60", `\x04`)  // backtick
	content = strings.ReplaceAll(content, `\[`, `\x05`)
	content = strings.ReplaceAll(content, `\]`, `\x06`)
	content = strings.ReplaceAll(content, `\(`, `\x07`)
	content = strings.ReplaceAll(content, `\)`, `\x08`)
	content = strings.ReplaceAll(content, `\|`, `\x09`)

	// 转换格式
	// *bold* → **bold**
	boldRe := regexp.MustCompile(`\*([^\*]+)\*`)
	content = boldRe.ReplaceAllString(content, `**$1**`)

	// _italic_ → *italic*
	italicRe := regexp.MustCompile(`(?<!\w)_([^_]+)_(?!\w)`)
	content = italicRe.ReplaceAllString(content, `*$1*`)

	// __underline__ → <u>underline</u>
	underlineRe := regexp.MustCompile(`__([^_]+)__`)
	content = underlineRe.ReplaceAllString(content, `<u>$1</u>`)

	// ~strikethrough~ → ~~strikethrough~~
	strikeRe := regexp.MustCompile(`(?<!\~)~([^~]+)~(?!\~)`)
	content = strikeRe.ReplaceAllString(content, `~~$1~~`)

	// ||spoiler|| → <details><summary>spoiler</summary></details>
	spoilerRe := regexp.MustCompile(`\|\|([^|]+)\|\|`)
	content = spoilerRe.ReplaceAllString(content, `<details><summary>$1</summary></details>`)

	// 恢复转义字符
	content = strings.ReplaceAll(content, `\x00`, `\`)
	content = strings.ReplaceAll(content, `\x01`, `*`)
	content = strings.ReplaceAll(content, `\x02`, `_`)
	content = strings.ReplaceAll(content, `\x03`, `~`)
	content = strings.ReplaceAll(content, "\x04", "`")
	content = strings.ReplaceAll(content, `\x05`, `[`)
	content = strings.ReplaceAll(content, `\x06`, `]`)
	content = strings.ReplaceAll(content, `\x07`, `(`)
	content = strings.ReplaceAll(content, `\x08`, `)`)
	content = strings.ReplaceAll(content, `\x09`, `|`)

	return content
}

func convertHTML(content string) string {
	// 直接透传支持的标签，忽略不支持的
	// 移除 <tg-emoji> 和 <tg-spoiler>
	re := regexp.MustCompile(`</?tg-(emoji|spoiler)[^>]*>`)
	return re.ReplaceAllString(content, "")
}
