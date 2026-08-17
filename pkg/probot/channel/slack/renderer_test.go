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

package slack_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/probot/channel/slack"
)

func blockAt(t *testing.T, blocks []any, index int) map[string]any {
	t.Helper()

	require.Greater(t, len(blocks), index)

	block, ok := blocks[index].(map[string]any)
	require.True(t, ok)

	return block
}

func nested(t *testing.T, block map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := block[key].(map[string]any)
	require.True(t, ok)

	return value
}

func elements(t *testing.T, block map[string]any, key string) []any {
	t.Helper()

	value, ok := block[key].([]any)
	require.True(t, ok)

	return value
}

func reviewIntent(items []bot.ItemIntent) bot.MessageIntent {
	return bot.MessageIntent{
		FallbackText: "New compliance portal access request",
		Headline:     "🔒 New compliance portal access request",
		Context:      "Requested by Jane Requester <jane@example.com>",
		Actions: []bot.ActionIntent{
			{
				ID:    "compliance_access.approve_all",
				Label: "Grant all",
				Style: bot.ActionStylePrimary,
			},
			{
				ID:    "compliance_access.deny_all",
				Label: "Reject/revoke all",
				Style: bot.ActionStyleDanger,
			},
			{
				Label: "Open in Probo",
				URL:   "https://app.example.com/access",
			},
		},
		Groups: []bot.GroupIntent{
			{
				ID:    "documents",
				Title: fmt.Sprintf("Documents (%d)", len(items)),
				Items: items,
			},
		},
	}
}

func TestRenderMessageIntent_RendersHeaderContextAndBulkActions(t *testing.T) {
	t.Parallel()

	payload := slack.RenderMessageIntent(reviewIntent(nil))

	assert.Equal(t, "New compliance portal access request", payload["text"])

	blocks := elements(t, payload, "blocks")
	require.Len(t, blocks, 3)

	headline := blockAt(t, blocks, 0)
	assert.Equal(t, "header", headline["type"])
	assert.Equal(
		t,
		"🔒 New compliance portal access request",
		nested(t, headline, "text")["text"],
	)

	context := blockAt(t, blocks, 1)
	assert.Equal(t, "context", context["type"])

	contextText, ok := elements(t, context, "elements")[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(
		t,
		"Requested by Jane Requester &lt;jane@example.com&gt;",
		contextText["text"],
	)

	actions := blockAt(t, blocks, 2)
	assert.Equal(t, "actions", actions["type"])

	buttons := elements(t, actions, "elements")
	require.Len(t, buttons, 3)

	grantAll, ok := buttons[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "primary", grantAll["style"])
	assert.Equal(t, "compliance_access.approve_all", grantAll["action_id"])

	openInProbo, ok := buttons[2].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "https://app.example.com/access", openInProbo["url"])
	assert.NotContains(t, openInProbo, "action_id")
}

func TestRenderMessageIntent_RendersPendingItemAsSingleRowWithMenu(t *testing.T) {
	t.Parallel()

	payload := slack.RenderMessageIntent(
		reviewIntent(
			[]bot.ItemIntent{
				{
					ID:    "document-1",
					Label: "Operations security policy",
					URL:   "https://app.example.com/documents/1",
					Action: &bot.ActionIntent{
						ID: "compliance_access.review_item",
						Options: []bot.ActionOptionIntent{
							{Label: "Grant", Value: "approve/document-1"},
							{Label: "Reject", Value: "reject/document-1"},
						},
					},
				},
			},
		),
	)

	blocks := elements(t, payload, "blocks")
	require.Len(t, blocks, 6)

	divider := blockAt(t, blocks, 3)
	assert.Equal(t, "divider", divider["type"])

	group := blockAt(t, blocks, 4)
	assert.Equal(t, "*Documents (1)*", nested(t, group, "text")["text"])

	item := blockAt(t, blocks, 5)
	assert.Equal(t, "section", item["type"])

	fields := elements(t, item, "fields")
	require.Len(t, fields, 1, "a pending row must not repeat a requested status")

	label, ok := fields[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(
		t,
		"<https://app.example.com/documents/1|Operations security policy>",
		label["text"],
	)

	accessory := nested(t, item, "accessory")
	assert.Equal(t, "overflow", accessory["type"])
	assert.Equal(t, "compliance_access.review_item", accessory["action_id"])

	options := elements(t, accessory, "options")
	require.Len(t, options, 2)

	grant, ok := options[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "approve/document-1", grant["value"])
	assert.Equal(t, "Grant", nested(t, grant, "text")["text"])
}

func TestRenderMessageIntent_RendersDecidedItemWithStatusAndButton(t *testing.T) {
	t.Parallel()

	payload := slack.RenderMessageIntent(
		reviewIntent(
			[]bot.ItemIntent{
				{
					ID:     "document-1",
					Label:  "Operations security policy",
					URL:    "https://app.example.com/documents/1",
					Status: "Granted",
					Action: &bot.ActionIntent{
						ID:    "compliance_access.deny_item",
						Label: "Revoke",
						Style: bot.ActionStyleDanger,
						Value: "document-1",
					},
				},
			},
		),
	)

	item := blockAt(t, elements(t, payload, "blocks"), 5)

	fields := elements(t, item, "fields")
	require.Len(t, fields, 2)

	status, ok := fields[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Granted", status["text"])

	accessory := nested(t, item, "accessory")
	assert.Equal(t, "button", accessory["type"])
	assert.Equal(t, "danger", accessory["style"])
	assert.Equal(t, "document-1", accessory["value"])
}

func TestRenderMessageIntent_TruncatesWithinSlackBlockLimit(t *testing.T) {
	t.Parallel()

	items := make([]bot.ItemIntent, 60)
	for i := range items {
		items[i] = bot.ItemIntent{
			ID:    fmt.Sprintf("document-%d", i),
			Label: fmt.Sprintf("Document %d", i),
		}
	}

	payload := slack.RenderMessageIntent(reviewIntent(items))

	blocks := elements(t, payload, "blocks")
	assert.LessOrEqual(t, len(blocks), 50)

	notice := blockAt(t, blocks, len(blocks)-1)
	assert.Equal(t, "context", notice["type"])

	noticeText, ok := elements(t, notice, "elements")[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "16 more items are only listed in Probo.", noticeText["text"])
}

func TestRenderMessageIntent_OmitsBlocksForPlainText(t *testing.T) {
	t.Parallel()

	payload := slack.RenderMessageIntent(bot.MessageIntent{FallbackText: "hello"})

	assert.Equal(t, "hello", payload["text"])
	assert.NotContains(t, payload, "blocks")
}
