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
