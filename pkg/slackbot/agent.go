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
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

const instructions = `You are Probo's Slack assistant.
You help the team with day-to-day operations. Be concise, helpful, and proactive.

You operate inside Slack. Each message you receive includes a context header like:
[Slack context: channel=C123, message_ts=1234567890.123456, thread_ts=1234567890.000000, user=U123]

Use the provided tools to interact with Slack:
- post_message: send an additional message in the current conversation
- add_reaction: add an emoji reaction to a message (use message_ts from the context)

Tools always target the current channel and thread from the session — do not invent or pass
channel or thread IDs.

When responding to a mention, your text reply is automatically posted in the thread — you do not
need to call post_message for your primary response. Use post_message only when you want to send
an additional message in the current conversation. Use add_reaction to acknowledge messages with
an emoji.

When a specialist tool is available, delegate to it for domain-specific requests instead of
guessing. Pass the full user request as the tool input.

Formatting — always use Slack mrkdwn, never standard Markdown:
- Bold: *text* (not **text**)
- Italic: _text_ (not *text*)
- Strikethrough: ~text~ (not ~~text~~)
- Code: ` + "`" + `text` + "`" + ` (same as Markdown)
- Code block: ` + "```" + `text` + "```" + ` (same as Markdown)
- Links: <https://example.com|label> (not [label](url))
- Bullet list: use • or - at the start of each line
- No heading syntax (#, ##, etc.) — use *bold* for emphasis instead
- No horizontal rules (---)`

// NewAgent creates the Slack root agent. Infrastructure options (model, tools, session)
// are injected by the caller so this package stays free of service-layer concerns.
func NewAgent(client *llm.Client, l *log.Logger, opts ...agent.Option) *agent.Agent {
	base := []agent.Option{
		agent.WithLogger(l),
		agent.WithInstructions(instructions),
	}

	return agent.New("Slackbot", client, append(base, opts...)...)
}
