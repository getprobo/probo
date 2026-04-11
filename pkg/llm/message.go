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

package llm

import "strings"

import "encoding/json"

type (
	Message struct {
		Role       Role
		Parts      []Part
		ToolCalls  []ToolCall
		ToolCallID string // set when Role is RoleTool
	}

	ToolCall struct {
		ID       string
		Function FunctionCall
	}

	FunctionCall struct {
		Name      string
		Arguments string // JSON-encoded arguments
	}

	Tool struct {
		Name        string
		Description string
		Parameters  json.RawMessage // JSON Schema
	}
)

func (m Message) Text() string {
	var s strings.Builder
	for _, p := range m.Parts {
		if tp, ok := p.(TextPart); ok {
			s.WriteString(tp.Text)
		}
	}
	return s.String()
}

func (m Message) Thinking() string {
	var s strings.Builder
	for _, p := range m.Parts {
		if tp, ok := p.(ThinkingPart); ok {
			s.WriteString(tp.Text)
		}
	}
	return s.String()
}
