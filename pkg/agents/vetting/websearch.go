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
	_ "embed"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed websearch_prompt.txt
var websearchSystemPrompt string

func newWebSearchAgent(
	client *llm.Client,
	model string,
	tools []agent.Tool,
	extraOpts ...agent.Option,
) (*agent.Agent, error) {
	outputType, err := agent.NewOutputType[WebSearchOutput]("web_search_output")
	if err != nil {
		return nil, fmt.Errorf("cannot create output type: %w", err)
	}

	opts := []agent.Option{
		agent.WithInstructions(websearchSystemPrompt),
		agent.WithModel(model),
		agent.WithTools(tools...),
		agent.WithMaxTurns(40),
		agent.WithOutputType(outputType),
		agent.WithParallelToolCalls(true),
	}
	opts = append(opts, extraOpts...)

	return agent.New("web_search_analyst", client, opts...), nil
}
