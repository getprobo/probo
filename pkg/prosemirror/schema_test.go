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

package prosemirror_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

func commentSchema() *prosemirror.Schema {
	return prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			prosemirror.Blockquote(),
			prosemirror.BulletList(),
			prosemirror.OrderedList(),
			prosemirror.ListItem(),
			prosemirror.HardBreak(),
			prosemirror.Mention(prosemirror.MentionOptions{}),
		},
		[]prosemirror.MarkSpec{
			prosemirror.Bold(),
			prosemirror.Italic(),
			prosemirror.Strike(),
			prosemirror.Code(),
			prosemirror.Link(),
		},
	)
}

func TestRestrictedSchemaRejectsExcludedNodesAndMarks(t *testing.T) {
	t.Parallel()

	schema := commentSchema()
	profileID := gid.New(gid.TenantID{}, 1).String()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{
			name: "paragraph allowed",
			in:   `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`,
		},
		{
			name: "mention allowed",
			in: fmt.Sprintf(
				`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
				profileID,
			),
		},
		{
			name:    "heading rejected",
			in:      `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}`,
			wantErr: `heading node is not allowed in this schema`,
		},
		{
			name:    "image rejected",
			in:      `{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/img.png"}}]}`,
			wantErr: `image node is not allowed in this schema`,
		},
		{
			name:    "underline rejected",
			in:      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"hi"}]}]}`,
			wantErr: `cannot validate mark: unknown type "underline"`,
		},
		{
			name:    "heading inside blockquote rejected",
			in:      `{"type":"doc","content":[{"type":"blockquote","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"hi"}]}]}]}`,
			wantErr: `heading node is not allowed in this schema`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := schema.ValidateJSON(tt.in)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMentionOptionsEntityTypes(t *testing.T) {
	t.Parallel()

	const allowedType uint16 = 7
	const otherType uint16 = 25

	schema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			prosemirror.Mention(prosemirror.MentionOptions{EntityTypes: []uint16{allowedType}}),
		},
		nil,
	)

	allowedID := gid.New(gid.TenantID{}, allowedType).String()
	otherID := gid.New(gid.TenantID{}, otherType).String()

	require.NoError(
		t,
		schema.ValidateJSON(fmt.Sprintf(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
			allowedID,
		)),
	)

	err := schema.ValidateJSON(fmt.Sprintf(
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
		otherID,
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

func TestNewSchemaPanics(t *testing.T) {
	t.Parallel()

	t.Run("missing doc", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			prosemirror.NewSchema([]prosemirror.NodeSpec{prosemirror.Paragraph(), prosemirror.Text()}, nil)
		})
	})

	t.Run("missing text", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			prosemirror.NewSchema([]prosemirror.NodeSpec{prosemirror.Doc(), prosemirror.Paragraph()}, nil)
		})
	})

	t.Run("duplicate node", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			prosemirror.NewSchema(
				[]prosemirror.NodeSpec{prosemirror.Doc(), prosemirror.Text(), prosemirror.Text()},
				nil,
			)
		})
	})

	t.Run("duplicate mark", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			prosemirror.NewSchema(
				[]prosemirror.NodeSpec{prosemirror.Doc(), prosemirror.Text()},
				[]prosemirror.MarkSpec{prosemirror.Bold(), prosemirror.Bold()},
			)
		})
	})

	t.Run("unknown content type", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			prosemirror.NewSchema(
				[]prosemirror.NodeSpec{
					prosemirror.Doc(),
					prosemirror.Text(),
					prosemirror.BulletList(),
				},
				nil,
			)
		})
	})
}

func TestSchemaGroupResolutionFollowsSchema(t *testing.T) {
	t.Parallel()

	schema := commentSchema()
	err := schema.ValidateJSON(
		`{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"ok"}]}]}]}`,
	)
	require.NoError(t, err)

	err = schema.ValidateJSON(
		`{"type":"doc","content":[{"type":"blockquote","content":[{"type":"codeBlock","content":[{"type":"text","text":"no"}]}]}]}`,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `codeBlock node is not allowed in this schema`)
}

func FuzzSanitize(f *testing.F) {
	seeds := []string{
		"",
		`{"type":"doc","content":[{"type":"paragraph"}]}`,
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi","marks":[{"type":"bold"},{"type":"link","attrs":{"href":"javascript:alert(1)"}}]}]}]}`,
		`{"type":"doc","content":[{"type":"image","attrs":{"src":"data:image/png;base64,abc","fileId":"  "}}]}`,
		`not json`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		once, err := prosemirror.DocumentSchema.Sanitize(raw)
		if err != nil {
			return
		}

		twice, err := prosemirror.DocumentSchema.Sanitize(once)
		require.NoError(t, err)
		assert.Equal(t, once, twice)

		require.NoError(t, prosemirror.DocumentSchema.ValidateJSON(once))

		if strings.TrimSpace(once) == "" {
			return
		}

		var n prosemirror.Node
		require.NoError(t, json.Unmarshal([]byte(once), &n))
		assert.Equal(t, prosemirror.NodeDoc, n.Type)
	})
}
