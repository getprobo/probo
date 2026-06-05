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
