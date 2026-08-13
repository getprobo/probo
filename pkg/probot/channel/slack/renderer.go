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

	"go.probo.inc/probo/pkg/probot"
)

const (
	messageBlockLimit = 50
	cardActionLimit   = 5
	cardTitleLimit    = 150
	cardSubtitleLimit = 150
	cardBodyLimit     = 200
)

func RenderMessageIntent(intent probot.MessageIntent) map[string]any {
	cards := intent.Cards
	if len(cards) > messageBlockLimit {
		hiddenCount := len(cards) - messageBlockLimit + 1
		visible := append([]probot.CardIntent(nil), cards[:messageBlockLimit-1]...)
		cards = append(
			visible,
			probot.CardIntent{
				ID:    "truncation-summary",
				Title: "Additional resources not shown",
				Body:  fmt.Sprintf("%d additional item(s) are available in Probo.", hiddenCount),
			},
		)
	}

	blocks := make([]any, 0, len(cards))
	for _, card := range cards {
		block := map[string]any{
			"type":     "card",
			"block_id": truncateText("card-"+card.ID, 255),
			"title": rawTextObject(
				renderCardTitle(card.Title, card.TitleURL),
				cardTitleLimit,
			),
		}
		if card.Subtitle != "" {
			block["subtitle"] = textObject(card.Subtitle, cardSubtitleLimit)
		}

		if card.Body != "" {
			block["body"] = textObject(card.Body, cardBodyLimit)
		}

		if len(card.Actions) > 0 {
			actionCount := min(len(card.Actions), cardActionLimit)

			actions := make([]any, 0, actionCount)
			for _, action := range card.Actions[:actionCount] {
				element := map[string]any{
					"type": "button",
					"text": map[string]any{
						"type":  "plain_text",
						"text":  truncateText(action.Label, 150),
						"emoji": false,
					},
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
				case probot.ActionStylePrimary:
					element["style"] = "primary"
				case probot.ActionStyleDanger:
					element["style"] = "danger"
				}

				actions = append(actions, element)
			}

			block["actions"] = actions
		}

		blocks = append(blocks, block)
	}

	return map[string]any{
		"text":   intent.FallbackText,
		"blocks": blocks,
	}
}

func textObject(text string, limit int) map[string]any {
	return map[string]any{
		"type":     "mrkdwn",
		"text":     escapeSlackText(truncateText(text, limit)),
		"verbatim": false,
	}
}

func rawTextObject(text string, limit int) map[string]any {
	return map[string]any{
		"type":     "mrkdwn",
		"text":     truncateText(text, limit),
		"verbatim": false,
	}
}

func renderCardTitle(title string, titleURL string) string {
	if titleURL == "" {
		return escapeSlackText(title)
	}

	available := cardTitleLimit - len([]rune(titleURL)) - 3
	if available < 1 {
		return escapeSlackText(truncateText(title, cardTitleLimit))
	}

	return fmt.Sprintf(
		"<%s|%s>",
		titleURL,
		escapeSlackText(truncateText(title, available)),
	)
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
