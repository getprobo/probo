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

package prosemirror

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDocumentContentJSON_Schema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "empty", in: "", wantErr: false},
		{name: "whitespace only", in: "   \n", wantErr: false},
		{
			name:    "valid paragraph",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: false,
		},
		{name: "plain text", in: "not json", wantErr: true},
		{name: "non-doc root", in: `{"type":"paragraph","content":[]}`, wantErr: true},
		{name: "unknown node", in: `{"type":"doc","content":[{"type":"unknownWidget"}]}`, wantErr: true},
		{
			name:    "unknown mark",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"glow"}],"text":"hi"}]}]}`,
			wantErr: true,
		},
		{
			name:    "invalid heading level",
			in:      `{"type":"doc","content":[{"type":"heading","attrs":{"level":9},"content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: true,
		},
		{
			name:    "list item at root",
			in:      `{"type":"doc","content":[{"type":"listItem","content":[{"type":"paragraph"}]}]}`,
			wantErr: true,
		},
		{
			name:    "empty doc content",
			in:      `{"type":"doc","content":[]}`,
			wantErr: true,
		},
		{
			name:    "heading inside paragraph",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "text at doc root",
			in:      `{"type":"doc","content":[{"type":"text","text":"hi"}]}`,
			wantErr: true,
		},
		{
			name:    "list item without paragraph",
			in:      `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "code block with paragraph child",
			in:      `{"type":"doc","content":[{"type":"codeBlock","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "valid list",
			in:      `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "valid task list",
			in:      `{"type":"doc","content":[{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"open"}]}]},{"type":"taskItem","attrs":{"checked":true},"content":[{"type":"paragraph","content":[{"type":"text","text":"done"}]}]}]}]}`,
			wantErr: false,
		},
		{
			name:    "task item in bullet list",
			in:      `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph"}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "task item checked is not a bool",
			in:      `{"type":"doc","content":[{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":"yes"},"content":[{"type":"paragraph"}]}]}]}`,
			wantErr: true,
		},
		{
			name:    "valid table",
			in:      `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph"}]}]}]}]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := testDocumentSchema.ValidateJSON(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateDocumentContentJSON_TestdataDocument(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(loadTestDocument(t))
	require.NoError(t, err)
	require.NoError(t, testDocumentSchema.ValidateJSON(string(raw)))
}

func TestLinkMark_InvalidAttrs(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"x","marks":[{"type":"link","attrs":{"href":1}}]}]}]}`

	err := testDocumentSchema.ValidateJSON(raw)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot parse link attrs")

	got, err := testDocumentSchema.Sanitize(raw)
	require.NoError(t, err)
	assert.Contains(t, got, `"href":"#"`)
}

func TestSanitizeDocumentJSON_EmptyUnchanged(t *testing.T) {
	t.Parallel()

	got, err := testDocumentSchema.Sanitize("")
	require.NoError(t, err)
	assert.Equal(t, "", got)

	got, err = testDocumentSchema.Sanitize("   ")
	require.NoError(t, err)
	assert.Equal(t, "   ", got)
}

func TestSanitizeDocumentJSON_NonJSONError(t *testing.T) {
	t.Parallel()

	_, err := testDocumentSchema.Sanitize("plain text is not valid document JSON")
	require.Error(t, err)
}

func TestSanitizeDocumentJSON_NonDocRootError(t *testing.T) {
	t.Parallel()

	_, err := testDocumentSchema.Sanitize(`{"type":"paragraph","content":[]}`)
	require.Error(t, err)
}

func TestSanitizeDocumentJSON_LinkHref(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)","target":"_blank"}}],"text":"click"}]}]}`

	out, err := testDocumentSchema.Sanitize(raw)
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	txt := doc.Content[0].Content[0]
	require.Len(t, txt.Marks, 1)
	attrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "#", attrs.Href)
	require.NotNil(t, attrs.Target)
	assert.Equal(t, "_blank", *attrs.Target)
}

func TestSanitizeDocumentJSON_PreservesSafeHref(t *testing.T) {
	t.Parallel()

	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com"}}],"text":"ok"}]}]}`

	out, err := testDocumentSchema.Sanitize(raw)
	require.NoError(t, err)

	var doc Node
	require.NoError(t, json.Unmarshal([]byte(out), &doc))
	txt := doc.Content[0].Content[0]
	attrs, err := txt.Marks[0].LinkAttrs()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", attrs.Href)
}
