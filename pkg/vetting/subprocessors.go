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

package vetting

import (
	"encoding/json"
	"strings"

	"go.probo.inc/probo/pkg/llm"
)

const extractSubprocessorsToolName = "extract_subprocessors"

// subprocessorsFromOrchestratorMessages collects sub-processors from every
// extract_subprocessors sub-agent tool result in the orchestrator transcript.
// Later tool calls win when the same name appears more than once.
func subprocessorsFromOrchestratorMessages(messages []llm.Message) []Subprocessor {
	toolNames := toolCallNamesByID(messages)

	byName := make(map[string]Subprocessor)
	order := make([]string, 0)

	for _, msg := range messages {
		if msg.Role != llm.RoleTool {
			continue
		}

		if toolNames[msg.ToolCallID] != extractSubprocessorsToolName {
			continue
		}

		text := strings.TrimSpace(msg.Text())
		if text == "" || !json.Valid([]byte(text)) {
			continue
		}

		var output SubprocessorOutput
		if err := json.Unmarshal([]byte(text), &output); err != nil {
			continue
		}

		for _, sub := range output.Subprocessors {
			if sub.Name == "" {
				continue
			}

			key := normalizeSubprocessorName(sub.Name)
			if _, exists := byName[key]; !exists {
				order = append(order, key)
			}

			byName[key] = sub
		}
	}

	if len(order) == 0 {
		return nil
	}

	subs := make([]Subprocessor, 0, len(order))
	for _, key := range order {
		subs = append(subs, byName[key])
	}

	return subs
}

// mergeSubprocessors prefers entries from primary (tool output). Names only
// present in secondary (markdown extraction) are appended afterward.
func mergeSubprocessors(primary, secondary []Subprocessor) []Subprocessor {
	if len(primary) == 0 {
		return secondary
	}

	if len(secondary) == 0 {
		return primary
	}

	merged := make([]Subprocessor, len(primary), len(primary)+len(secondary))
	copy(merged, primary)

	seen := make(map[string]struct{}, len(primary))
	for _, sub := range primary {
		seen[normalizeSubprocessorName(sub.Name)] = struct{}{}
	}

	for _, sub := range secondary {
		if sub.Name == "" {
			continue
		}

		key := normalizeSubprocessorName(sub.Name)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		merged = append(merged, sub)
	}

	return merged
}

func subprocessorListURLFromOrchestratorMessages(messages []llm.Message) string {
	toolNames := toolCallNamesByID(messages)

	var source string

	for _, msg := range messages {
		if msg.Role != llm.RoleTool {
			continue
		}

		if toolNames[msg.ToolCallID] != extractSubprocessorsToolName {
			continue
		}

		text := strings.TrimSpace(msg.Text())
		if text == "" || !json.Valid([]byte(text)) {
			continue
		}

		var output SubprocessorOutput
		if err := json.Unmarshal([]byte(text), &output); err != nil {
			continue
		}

		if strings.TrimSpace(output.Source) != "" {
			source = strings.TrimSpace(output.Source)
		}
	}

	return source
}

func toolCallNamesByID(messages []llm.Message) map[string]string {
	toolNames := make(map[string]string)

	for _, msg := range messages {
		if msg.Role != llm.RoleAssistant {
			continue
		}

		for _, tc := range msg.ToolCalls {
			toolNames[tc.ID] = tc.Function.Name
		}
	}

	return toolNames
}

func normalizeSubprocessorName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
