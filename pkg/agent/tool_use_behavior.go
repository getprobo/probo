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

package agent

import "context"

// ToolUseBehavior decides whether tool-call results constitute a final
// output. When isFinal is true the run stops and finalOutput is returned
// to the caller without sending results back to the LLM.
type ToolUseBehavior func(ctx context.Context, results []ToolCallResult) (finalOutput string, isFinal bool, err error)

type ToolCallResult struct {
	ToolName  string
	Arguments string
	Result    ToolResult
}

// RunLLMAgain is the default behavior: tools run, then the LLM receives the
// results and gets to respond.
func RunLLMAgain() ToolUseBehavior {
	return func(_ context.Context, _ []ToolCallResult) (string, bool, error) {
		return "", false, nil
	}
}

// StopOnFirstTool treats the output from the first tool call as the final
// result, without sending it back to the LLM.
func StopOnFirstTool() ToolUseBehavior {
	return func(_ context.Context, results []ToolCallResult) (string, bool, error) {
		if len(results) == 0 {
			return "", false, nil
		}
		return results[0].Result.Content, true, nil
	}
}

// StopAtTools stops the agent when any of the listed tool names is called.
// The matching tool's output becomes the final output.
func StopAtTools(names ...string) ToolUseBehavior {
	stopSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		stopSet[n] = struct{}{}
	}
	return func(_ context.Context, results []ToolCallResult) (string, bool, error) {
		for _, r := range results {
			if _, ok := stopSet[r.ToolName]; ok {
				return r.Result.Content, true, nil
			}
		}
		return "", false, nil
	}
}
