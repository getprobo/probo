// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package vetting

import (
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

// subAgentSpec describes a vetting sub-agent. The generic builder
// `newSubAgent[T]` reads it once and constructs the agent. This avoids
// duplicating the same option boilerplate across 16 constructor functions.
type subAgentSpec struct {
	name           string
	outputName     string
	prompt         string
	maxTurns       int
	thinkingBudget int  // 0 disables extended thinking
	parallelTools  bool // true enables parallel tool calls
}

// newSubAgent builds a vetting sub-agent from its spec, the tools it
// should use, and any caller-supplied extra options (logger, hooks).
// The type parameter T is the structured output type the agent must
// produce.
func newSubAgent[T any](
	client *llm.Client,
	model string,
	spec subAgentSpec,
	tools []agent.Tool,
	extraOpts ...agent.Option,
) (*agent.Agent, error) {
	outputType, err := agent.NewOutputType[T](spec.outputName)
	if err != nil {
		return nil, fmt.Errorf("cannot create output type %q: %w", spec.outputName, err)
	}

	opts := []agent.Option{
		agent.WithInstructions(spec.prompt),
		agent.WithModel(model),
		agent.WithTools(tools...),
		agent.WithMaxTurns(spec.maxTurns),
		agent.WithOutputType(outputType),
	}
	if spec.thinkingBudget > 0 {
		opts = append(opts, agent.WithThinking(spec.thinkingBudget))
	}
	if spec.parallelTools {
		opts = append(opts, agent.WithParallelToolCalls(true))
	}
	opts = append(opts, extraOpts...)

	return agent.New(spec.name, client, opts...), nil
}
