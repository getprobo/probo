// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"context"
	"crypto/sha256"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	TurnBinding struct {
		OrganizationID gid.GID
		ChannelID      string
		ThreadTS       string
		MessageTS      string
	}

	sendMessageParams struct {
		Text string `json:"text" jsonschema_description:"Plain-text message to send in the current conversation"`
	}

	addReactionParams struct {
		Reaction string `json:"reaction" jsonschema_description:"Emoji name without colons (e.g. thumbsup, white_check_mark, eyes)"`
	}
)

func Tools(queue *DeliveryService, turn TurnBinding) []agent.Tool {
	return []agent.Tool{
		agent.FunctionTool(
			"send_message",
			"Send a concise plain-text user-visible reply in the current trusted Slack conversation. Do not use Slack-specific markup or invent channel or thread IDs.",
			func(ctx context.Context, p sendMessageParams) (agent.ToolResult, error) {
				if turn.ChannelID == "" {
					return agent.ResultError("cannot queue message without a trusted Slack conversation"), nil
				}

				operationKey, ok := toolOperationKey(ctx, turn, "send_message")
				if !ok {
					return agent.ResultError("cannot queue message without a stable tool call ID"), nil
				}

				intent := bot.MessageIntent{FallbackText: p.Text}
				body := RenderMessageIntent(intent)

				_, _, err := queue.Queue(
					ctx,
					turn.OrganizationID,
					operationKey,
					coredata.SlackDeliveryOperationKindPostMessage,
					map[string]any{
						"channel":   turn.ChannelID,
						"text":      body["text"],
						"thread_ts": turn.ThreadTS,
					},
				)
				if err != nil {
					return agent.ResultErrorf("cannot queue message: %s", err), nil
				}

				return agent.ToolResult{Content: "message queued for durable Slack delivery"}, nil
			},
		),
		agent.FunctionTool(
			"add_reaction",
			"Add an emoji reaction to the current trusted Slack message. Never invent channel or message IDs.",
			func(ctx context.Context, p addReactionParams) (agent.ToolResult, error) {
				if turn.ChannelID == "" || turn.MessageTS == "" {
					return agent.ResultError("cannot queue reaction without a trusted current Slack message"), nil
				}

				operationKey, ok := toolOperationKey(ctx, turn, "add_reaction")
				if !ok {
					return agent.ResultError("cannot queue reaction without a stable tool call ID"), nil
				}

				_, _, err := queue.Queue(
					ctx,
					turn.OrganizationID,
					operationKey,
					coredata.SlackDeliveryOperationKindAddReaction,
					map[string]any{
						"channel":   turn.ChannelID,
						"reaction":  p.Reaction,
						"timestamp": turn.MessageTS,
					},
				)
				if err != nil {
					return agent.ResultErrorf("cannot queue reaction: %s", err), nil
				}

				return agent.ToolResult{Content: "reaction queued for durable Slack delivery"}, nil
			},
		),
	}
}

func toolOperationKey(ctx context.Context, turn TurnBinding, toolName string) (string, bool) {
	toolCallID, ok := agent.ToolCallIDFrom(ctx)
	if !ok {
		return "", false
	}

	sum := sha256.Sum256(
		fmt.Appendf(
			nil,
			"%s\x00%s\x00%s\x00%s\x00%s\x00%s",
			turn.OrganizationID.String(),
			turn.ChannelID,
			turn.ThreadTS,
			turn.MessageTS,
			toolName,
			toolCallID,
		),
	)

	return fmt.Sprintf("agent-tool:%x", sum[:]), true
}
