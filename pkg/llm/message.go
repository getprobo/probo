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

import (
	"encoding/json"
	"fmt"
	"strings"
)

type (
	Message struct {
		Role       Role
		Parts      []Part
		ToolCalls  []ToolCall
		ToolCallID string // set when Role is RoleTool
	}

	ToolCall struct {
		ID       string       `json:"id"`
		Function FunctionCall `json:"function"`
	}

	FunctionCall struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}

	Tool struct {
		Name        string
		Description string
		Parameters  json.RawMessage // JSON Schema
	}
)

type partEnvelope struct {
	Type string `json:"type"`
	// TextPart fields
	Text string `json:"text,omitempty"`
	// ImagePart fields
	URL string `json:"url,omitempty"`
	// FilePart fields
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type messageJSON struct {
	Role       Role           `json:"role"`
	Parts      []partEnvelope `json:"parts,omitempty"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

func (m Message) MarshalJSON() ([]byte, error) {
	mj := messageJSON{
		Role:       m.Role,
		ToolCalls:  m.ToolCalls,
		ToolCallID: m.ToolCallID,
	}

	for _, p := range m.Parts {
		switch v := p.(type) {
		case TextPart:
			mj.Parts = append(mj.Parts, partEnvelope{Type: "text", Text: v.Text})
		case ImagePart:
			mj.Parts = append(mj.Parts, partEnvelope{Type: "image", URL: v.URL})
		case FilePart:
			mj.Parts = append(mj.Parts, partEnvelope{
				Type: "file", Data: v.Data, MimeType: v.MimeType, Filename: v.Filename,
			})
		default:
			return nil, fmt.Errorf("cannot marshal unknown Part type %T", p)
		}
	}

	return json.Marshal(mj)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var mj messageJSON
	if err := json.Unmarshal(data, &mj); err != nil {
		return err
	}

	m.Role = mj.Role
	m.ToolCalls = mj.ToolCalls
	m.ToolCallID = mj.ToolCallID

	if len(mj.Parts) == 0 {
		m.Parts = nil
		return nil
	}

	m.Parts = make([]Part, len(mj.Parts))
	for i, env := range mj.Parts {
		switch env.Type {
		case "text":
			m.Parts[i] = TextPart{Text: env.Text}
		case "image":
			m.Parts[i] = ImagePart{URL: env.URL}
		case "file":
			m.Parts[i] = FilePart{Data: env.Data, MimeType: env.MimeType, Filename: env.Filename}
		default:
			return fmt.Errorf("cannot unmarshal unknown Part type %q", env.Type)
		}
	}

	return nil
}

func (m Message) Text() string {
	var s strings.Builder
	for _, p := range m.Parts {
		if tp, ok := p.(TextPart); ok {
			s.WriteString(tp.Text)
		}
	}
	return s.String()
}
