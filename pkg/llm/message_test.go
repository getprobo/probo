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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "text only",
			msg: Message{
				Role:  RoleUser,
				Parts: []Part{TextPart{Text: "hello"}},
			},
		},
		{
			name: "image part",
			msg: Message{
				Role:  RoleUser,
				Parts: []Part{ImagePart{URL: "https://example.com/img.png"}},
			},
		},
		{
			name: "file part",
			msg: Message{
				Role: RoleUser,
				Parts: []Part{FilePart{
					Data: "aGVsbG8=", MimeType: "text/plain", Filename: "hello.txt",
				}},
			},
		},
		{
			name: "mixed parts",
			msg: Message{
				Role: RoleUser,
				Parts: []Part{
					TextPart{Text: "see attached"},
					FilePart{Data: "aGVsbG8=", MimeType: "text/plain", Filename: "hello.txt"},
				},
			},
		},
		{
			name: "assistant with tool calls",
			msg: Message{
				Role:  RoleAssistant,
				Parts: []Part{TextPart{Text: "calling tool"}},
				ToolCalls: []ToolCall{{
					ID:       "call_1",
					Function: FunctionCall{Name: "search", Arguments: `{"q":"test"}`},
				}},
			},
		},
		{
			name: "tool response",
			msg: Message{
				Role:       RoleTool,
				ToolCallID: "call_1",
				Parts:      []Part{TextPart{Text: "result"}},
			},
		},
		{
			name: "empty parts",
			msg:  Message{Role: RoleAssistant},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.msg)
			require.NoError(t, err)

			var got Message
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.msg, got)
		})
	}
}
