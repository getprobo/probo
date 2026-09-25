// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package slack

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/bot"
)

const (
	messageBlockLimit = 50
	actionLimit       = 5
	headlineLimit     = 150
	itemLabelLimit    = 150
	labelLimit        = 75
	fieldTextLimit    = 2000
	textLimit         = 3000
)

func RenderMessageIntent(intent bot.MessageIntent) map[string]any {
	blocks := renderIntentHeader(intent)

	itemCount := 0
	for _, group := range intent.Groups {
		itemCount += len(group.Items)
	}

	renderedCount := 0
	truncated := false

	for _, group := range intent.Groups {
		if len(group.Items) == 0 {
			continue
		}

		// A group costs a divider and a title before its first row, and one
		// block stays free for the truncation notice.
		if len(blocks)+4 > messageBlockLimit {
			truncated = true

			break
		}

		blocks = append(
			blocks,
			map[string]any{
				"type":     "divider",
				"block_id": blockID("divider", group.ID),
			},
			map[string]any{
				"type":     "section",
				"block_id": blockID("group", group.ID),
				"text":     markdownText("*"+escapeSlackText(group.Title)+"*", textLimit),
			},
		)

		for _, item := range group.Items {
			if len(blocks)+2 > messageBlockLimit {
				truncated = true

				break
			}

			blocks = append(blocks, renderItemBlock(item))
			renderedCount++
		}

		if truncated {
			break
		}
	}

	if truncated {
		blocks = append(blocks, renderTruncationBlock(itemCount-renderedCount))
	}

	payload := map[string]any{"text": intent.FallbackText}
	if len(blocks) > 0 {
		payload["blocks"] = blocks
	}

	return payload
}

func renderIntentHeader(intent bot.MessageIntent) []any {
	blocks := make([]any, 0, 4)
	if intent.Headline != "" {
		blocks = append(
			blocks,
			map[string]any{
				"type":     "header",
				"block_id": "headline",
				"text": map[string]any{
					"type": "plain_text",
					"text": truncateText(intent.Headline, headlineLimit),
				},
			},
		)
	}

	if intent.Context != "" {
		blocks = append(
			blocks,
			map[string]any{
				"type":     "context",
				"block_id": "context",
				"elements": []any{
					markdownText(escapeSlackText(intent.Context), textLimit),
				},
			},
		)
	}

	if intent.Body != "" {
		blocks = append(
			blocks,
			map[string]any{
				"type":     "section",
				"block_id": "body",
				"text":     markdownText(escapeSlackText(intent.Body), textLimit),
			},
		)
	}

	if len(intent.Actions) > 0 {
		blocks = append(
			blocks,
			map[string]any{
				"type":     "actions",
				"block_id": "actions",
				"elements": renderActionElements(intent.Actions),
			},
		)
	}

	return blocks
}

func renderItemBlock(item bot.ItemIntent) map[string]any {
	fields := []any{
		markdownText(renderItemLabel(item), fieldTextLimit),
	}
	if item.Status != "" {
		fields = append(fields, markdownText(escapeSlackText(item.Status), fieldTextLimit))
	}

	block := map[string]any{
		"type":     "section",
		"block_id": blockID("item", item.ID),
		"fields":   fields,
	}
	if item.Action != nil {
		block["accessory"] = renderActionElement(*item.Action)
	}

	return block
}

func renderActionElements(actions []bot.ActionIntent) []any {
	actionCount := min(len(actions), actionLimit)

	elements := make([]any, 0, actionCount)
	for _, action := range actions[:actionCount] {
		elements = append(elements, renderActionElement(action))
	}

	return elements
}

func renderActionElement(action bot.ActionIntent) map[string]any {
	if len(action.Options) > 0 {
		options := make([]any, 0, len(action.Options))
		for _, option := range action.Options {
			options = append(
				options,
				map[string]any{
					"text":  plainText(option.Label, labelLimit),
					"value": option.Value,
				},
			)
		}

		return map[string]any{
			"type":      "overflow",
			"action_id": action.ID,
			"options":   options,
		}
	}

	element := map[string]any{
		"type": "button",
		"text": plainText(action.Label, labelLimit),
	}
	if action.ID != "" {
		element["action_id"] = action.ID
	}

	if action.Value != "" {
		element["value"] = action.Value
	}

	if action.URL != "" {
		element["url"] = action.URL
	}

	switch action.Style {
	case bot.ActionStylePrimary:
		element["style"] = "primary"
	case bot.ActionStyleDanger:
		element["style"] = "danger"
	}

	return element
}

func renderTruncationBlock(hiddenCount int) map[string]any {
	noun := "items are"
	if hiddenCount == 1 {
		noun = "item is"
	}

	return map[string]any{
		"type":     "context",
		"block_id": "truncation",
		"elements": []any{
			markdownText(
				fmt.Sprintf("%d more %s only listed in Probo.", hiddenCount, noun),
				textLimit,
			),
		},
	}
}

func renderItemLabel(item bot.ItemIntent) string {
	if item.URL == "" {
		return escapeSlackText(truncateText(item.Label, itemLabelLimit))
	}

	available := fieldTextLimit - len([]rune(item.URL)) - 3
	if available < 1 {
		return escapeSlackText(truncateText(item.Label, itemLabelLimit))
	}

	available = min(available, itemLabelLimit)

	return fmt.Sprintf(
		"<%s|%s>",
		item.URL,
		escapeSlackText(truncateText(item.Label, available)),
	)
}

func blockID(prefix string, id string) string {
	return truncateText(prefix+"-"+id, 255)
}

func markdownText(text string, limit int) map[string]any {
	return map[string]any{
		"type":     "mrkdwn",
		"text":     truncateText(text, limit),
		"verbatim": false,
	}
}

func plainText(text string, limit int) map[string]any {
	return map[string]any{
		"type":  "plain_text",
		"text":  truncateText(text, limit),
		"emoji": false,
	}
}

func truncateText(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}

func escapeSlackText(text string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(text)
}
