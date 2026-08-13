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

const (
	SlashCommandName = "/probot"

	bindRequiredText = "I can't help until your Probo account is linked. " +
		"Run `/probot bind` — I'll send you a private link."

	bindSlashAlreadyLinkedText = "Your Probo account is already linked."
	bindSlashLinkedText        = "Your Probo account is linked."
	bindSlashUsageText         = "Usage: `/probot bind`"
	bindSlashUnavailableText   = "Probot is not available in this workspace."
	bindSlashFallbackText      = "Link your Probo account to use the Probo Slack assistant."
	bindSlashFailedText        = "I couldn't create a link right now. Try `/probot bind` again."
)

func bindRequiredBlocks(bindURL string) []any {
	return []any{
		map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "Link your *Probo* account before I can help you here. " +
					"You'll confirm the connection while signed in to Probo.",
			},
		},
		map[string]any{
			"type": "actions",
			"elements": []any{
				map[string]any{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "Link Probo account",
					},
					"url":   bindURL,
					"style": "primary",
				},
			},
		},
	}
}
