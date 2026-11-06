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

package probo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/validator"
)

func TestCommentSchemaParity(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "comment-schema.json"))
	require.NoError(t, err)

	var ledger struct {
		Nodes []string `json:"nodes"`
		Marks []string `json:"marks"`
	}
	require.NoError(t, json.Unmarshal(raw, &ledger))

	gotNodes := make([]string, 0, len(CommentSchema.NodeTypes()))
	for _, nodeType := range CommentSchema.NodeTypes() {
		gotNodes = append(gotNodes, string(nodeType))
	}

	gotMarks := make([]string, 0, len(CommentSchema.MarkTypes()))
	for _, markType := range CommentSchema.MarkTypes() {
		gotMarks = append(gotMarks, string(markType))
	}

	assert.ElementsMatch(t, ledger.Nodes, gotNodes)
	assert.ElementsMatch(t, ledger.Marks, gotMarks)

	// Sanity: CommentSchema stays a strict subset of DocumentSchema.
	docNodes := map[prosemirror.NodeType]struct{}{}
	for _, nodeType := range prosemirror.DocumentSchema.NodeTypes() {
		docNodes[nodeType] = struct{}{}
	}
	for _, nodeType := range CommentSchema.NodeTypes() {
		_, ok := docNodes[nodeType]
		assert.Truef(t, ok, "comment node %q missing from DocumentSchema", nodeType)
	}
}

func TestCommentSchema_ValidateJSON(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, coredata.MembershipProfileEntityType).String()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "empty", in: "", wantErr: false},
		{
			name:    "paragraph",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: false,
		},
		{
			name: "mention",
			in: fmt.Sprintf(
				`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
				profileID,
			),
			wantErr: false,
		},
		{
			name: "mention wrong entity type",
			in: fmt.Sprintf(
				`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
				gid.New(gid.TenantID{}, coredata.DocumentEntityType).String(),
			),
			wantErr: true,
		},
		{
			name:    "heading rejected",
			in:      `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: true,
		},
		{
			name:    "image rejected",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}`,
			wantErr: true,
		},
		{
			name:    "underline rejected",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"hi"}]}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := CommentSchema.ValidateJSON(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestProseMirrorContent_CommentSchema(t *testing.T) {
	t.Parallel()

	validDoc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`
	heading := `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}`
	image := `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}`
	underline := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"hi"}]}]}`
	list := `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}]}]}`
	blockquote := `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}]}`

	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{"empty string", "", false},
		{"valid paragraph", validDoc, false},
		{"valid list", list, false},
		{"valid blockquote", blockquote, false},
		{"heading rejected", heading, true},
		{"image rejected", image, true},
		{"underline rejected", underline, true},
		{"plain text", "not json", true},
		{"non-string", 1, true},
	}

	fn := validator.ProseMirrorContent(CommentSchema)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fn(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ProseMirrorContent(CommentSchema) error = %v, wantError %v", err, tt.wantError)
			}

			if err != nil && err.Code != validator.ErrorCodeInvalidFormat {
				t.Errorf("expected code %s, got %s", validator.ErrorCodeInvalidFormat, err.Code)
			}
		})
	}
}
