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

package slackbot

import (
	"context"

	"go.probo.inc/probo/pkg/agent"
)

type (
	postMessageParams struct {
		Text string `json:"text" jsonschema_description:"Message text (supports Slack mrkdwn formatting)"`
	}

	addReactionParams struct {
		Timestamp string `json:"timestamp" jsonschema_description:"Timestamp of the message to react to (e.g. 1234567890.123456)"`
		Reaction  string `json:"reaction" jsonschema_description:"Emoji name without colons (e.g. thumbsup, white_check_mark, eyes)"`
	}
)

func Tools(c *Client) []agent.Tool {
	return []agent.Tool{
		agent.FunctionTool(
			"post_message",
			"Post a message in the current Slack conversation (channel/thread from the session).",
			func(ctx context.Context, p postMessageParams) (agent.ToolResult, error) {
				rc := agent.RunContextFrom[*RunContext](ctx)
				if err := c.PostMessage(ctx, rc.Channel, p.Text, rc.ThreadTS); err != nil {
					return agent.ResultErrorf("cannot post message: %s", err), nil
				}

				return agent.ToolResult{Content: "message sent"}, nil
			},
		),
		agent.FunctionTool(
			"add_reaction",
			"Add an emoji reaction to a Slack message in the current channel.",
			func(ctx context.Context, p addReactionParams) (agent.ToolResult, error) {
				rc := agent.RunContextFrom[*RunContext](ctx)
				if err := c.AddReaction(ctx, rc.Channel, p.Reaction, p.Timestamp); err != nil {
					return agent.ResultErrorf("cannot add reaction: %s", err), nil
				}

				return agent.ToolResult{Content: "reaction added"}, nil
			},
		),
	}
}
