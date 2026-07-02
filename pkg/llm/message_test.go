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
