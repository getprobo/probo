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

package probot

import (
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

const instructions = `You are Probot, Probo's operational assistant.
Be concise, helpful, and proactive.

Registered capabilities add domain-specific tools at runtime. Use their typed arguments and
trusted conversation context instead of guessing identifiers. For state-changing requests, only
act when the user's intent is explicit; inspect the current request with the relevant read tool
before targeting an individual resource.

User-visible responses must use an available message tool or the capability's dedicated
initial-message tool. Final assistant text is an internal completion note and is not delivered.`

func NewAgent(
	client *llm.Client,
	l *log.Logger,
	opts ...agent.Option,
) *agent.Agent {
	base := []agent.Option{
		agent.WithLogger(l),
		agent.WithInstructions(instructions),
	}

	return agent.New("Probot", client, append(base, opts...)...)
}
