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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/html"
	"go.probo.inc/probo/pkg/prosemirror/markdown"
	"go.probo.inc/probo/pkg/richtext"
	"go.probo.inc/probo/pkg/richtext/mention"
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
			mention.Document(),
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

func TestMentionProfileVsDocument(t *testing.T) {
	t.Parallel()

	profileSchema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			mention.Profile(),
		},
		nil,
	)
	documentSchema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			mention.Document(),
		},
		nil,
	)

	profileID := gid.New(gid.TenantID{}, coredata.MembershipProfileEntityType).String()
	otherID := gid.New(gid.TenantID{}, 25).String()

	require.NoError(
		t,
		profileSchema.ValidateJSON(fmt.Sprintf(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
			profileID,
		)),
	)

	err := profileSchema.ValidateJSON(fmt.Sprintf(
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
		otherID,
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")

	require.NoError(
		t,
		documentSchema.ValidateJSON(fmt.Sprintf(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
			otherID,
		)),
	)
	require.NoError(
		t,
		documentSchema.ValidateJSON(fmt.Sprintf(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"%s","label":"Alice"}}]}]}`,
			profileID,
		)),
	)
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

type (
	renderCallout struct{}

	calloutHTMLRenderer     struct{}
	calloutMarkdownRenderer struct{}
)

func (renderCallout) Name() prosemirror.NodeType { return "callout" }

func (renderCallout) Group() prosemirror.Group { return prosemirror.GroupBlock }

func (renderCallout) Content() prosemirror.ContentRule {
	return prosemirror.ContentRule{Groups: []prosemirror.Group{prosemirror.GroupInline}}
}

func (renderCallout) Check(_ *prosemirror.Node, _ prosemirror.CheckMode) error {
	return nil
}

func (calloutHTMLRenderer) Node() prosemirror.NodeType { return "callout" }

func (calloutHTMLRenderer) Render(w *html.Writer, n prosemirror.Node) error {
	w.WriteString(`<aside class="callout">`)

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</aside>")

	return nil
}

func (calloutMarkdownRenderer) Node() prosemirror.NodeType { return "callout" }

func (calloutMarkdownRenderer) Render(w *markdown.Writer, n prosemirror.Node) error {
	return w.RenderInline(n.Content)
}

func TestRender_CustomNodeSpec(t *testing.T) {
	t.Parallel()

	schema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			renderCallout{},
		},
		nil,
	)
	raw := `{"type":"doc","content":[{"type":"callout","content":[{"type":"text","text":"hello"}]}]}`
	node, err := schema.Parse(raw)
	require.NoError(t, err)

	htmlRenderer := html.NewRenderer(schema, []html.NodeRenderer{
		html.Doc(),
		html.Paragraph(),
		html.Text(),
		calloutHTMLRenderer{},
	}, html.Marks())
	gotHTML, err := htmlRenderer.Render(node)
	require.NoError(t, err)
	assert.Equal(t, `<aside class="callout">hello</aside>`, gotHTML)

	mdRenderer := markdown.NewRenderer(schema, htmlRenderer, []markdown.NodeRenderer{
		markdown.Doc(),
		markdown.Paragraph(),
		markdown.Text(),
		calloutMarkdownRenderer{},
	}, markdown.Marks())
	gotMD, err := mdRenderer.Render(node)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", gotMD)
}

type (
	replacedHeading struct{}

	replacedHeadingHTMLRenderer     struct{}
	replacedHeadingMarkdownRenderer struct{}
)

func (replacedHeading) Name() prosemirror.NodeType { return prosemirror.NodeHeading }

func (replacedHeading) Group() prosemirror.Group { return prosemirror.GroupBlock }

func (replacedHeading) Content() prosemirror.ContentRule {
	return prosemirror.ContentRule{Groups: []prosemirror.Group{prosemirror.GroupInline}}
}

func (replacedHeading) Check(_ *prosemirror.Node, _ prosemirror.CheckMode) error {
	return nil
}

func (replacedHeadingHTMLRenderer) Node() prosemirror.NodeType { return prosemirror.NodeHeading }

func (replacedHeadingHTMLRenderer) Render(w *html.Writer, n prosemirror.Node) error {
	w.WriteString(`<h1 class="replaced">`)

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	w.WriteString("</h1>")

	return nil
}

func (replacedHeadingMarkdownRenderer) Node() prosemirror.NodeType { return prosemirror.NodeHeading }

func (replacedHeadingMarkdownRenderer) Render(w *markdown.Writer, n prosemirror.Node) error {
	return w.RenderInline(n.Content)
}

func TestRender_ReplacementHeadingSpec(t *testing.T) {
	t.Parallel()

	schema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
			replacedHeading{},
		},
		nil,
	)
	raw := `{"type":"doc","content":[{"type":"heading","content":[{"type":"text","text":"Hi"}]}]}`
	node, err := schema.Parse(raw)
	require.NoError(t, err)

	nodes := []html.NodeRenderer{
		DocHTMLRenderer(),
		ParagraphHTMLRenderer(),
		TextHTMLRenderer(),
		replacedHeadingHTMLRenderer{},
	}
	htmlRenderer := html.NewRenderer(schema, nodes, html.Marks())
	gotHTML, err := htmlRenderer.Render(node)
	require.NoError(t, err)
	assert.Equal(t, `<h1 class="replaced">Hi</h1>`, gotHTML)

	mdNodes := []markdown.NodeRenderer{
		markdown.Doc(),
		markdown.Paragraph(),
		markdown.Text(),
		replacedHeadingMarkdownRenderer{},
	}
	mdRenderer := markdown.NewRenderer(schema, htmlRenderer, mdNodes, markdown.Marks())
	gotMD, err := mdRenderer.Render(node)
	require.NoError(t, err)
	assert.Equal(t, "Hi\n", gotMD)
}

func DocHTMLRenderer() html.NodeRenderer {
	for _, n := range html.Nodes() {
		if n.Node() == prosemirror.NodeDoc {
			return n
		}
	}

	panic("doc renderer missing")
}

func ParagraphHTMLRenderer() html.NodeRenderer {
	for _, n := range html.Nodes() {
		if n.Node() == prosemirror.NodeParagraph {
			return n
		}
	}

	panic("paragraph renderer missing")
}

func TextHTMLRenderer() html.NodeRenderer {
	for _, n := range html.Nodes() {
		if n.Node() == prosemirror.NodeText {
			return n
		}
	}

	panic("text renderer missing")
}

type (
	highlightMark struct{}

	highlightHTMLRenderer     struct{}
	highlightMarkdownRenderer struct{}
)

func (highlightMark) Name() prosemirror.MarkType { return "highlight" }

func (highlightMark) Check(_ *prosemirror.Mark, _ prosemirror.CheckMode) error {
	return nil
}

func (highlightHTMLRenderer) Mark() prosemirror.MarkType { return "highlight" }

func (highlightHTMLRenderer) Open(w *html.Writer, _ prosemirror.Mark) error {
	w.WriteString("<mark>")

	return nil
}

func (highlightHTMLRenderer) Close(w *html.Writer, _ prosemirror.Mark) {
	w.WriteString("</mark>")
}

func (highlightMarkdownRenderer) Mark() prosemirror.MarkType { return "highlight" }

func (highlightMarkdownRenderer) Open(_ *markdown.Writer, _ prosemirror.Mark) error {
	return nil
}

func (highlightMarkdownRenderer) Close(_ *markdown.Writer, _ prosemirror.Mark) error {
	return nil
}

func TestRender_CustomMarkSpec(t *testing.T) {
	t.Parallel()

	schema := prosemirror.NewSchema(
		[]prosemirror.NodeSpec{
			prosemirror.Doc(),
			prosemirror.Paragraph(),
			prosemirror.Text(),
		},
		[]prosemirror.MarkSpec{highlightMark{}},
	)
	raw := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi","marks":[{"type":"highlight"}]}]}]}`
	node, err := schema.Parse(raw)
	require.NoError(t, err)

	htmlRenderer := html.NewRenderer(
		schema,
		html.Nodes(),
		append(html.Marks(), highlightHTMLRenderer{}),
	)
	gotHTML, err := htmlRenderer.Render(node)
	require.NoError(t, err)
	assert.Equal(t, "<p><mark>hi</mark></p>", gotHTML)
}

func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		`{"type":"doc"}`,
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hi"}]}]}`,
		`not json`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		node, err := prosemirror.Parse(raw)
		if err != nil {
			return
		}

		encoded, err := json.Marshal(node)
		require.NoError(t, err)

		again, err := prosemirror.Parse(string(encoded))
		require.NoError(t, err)
		assert.Equal(t, node.Type, again.Type)
	})
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
		once, err := richtext.DocumentSchema.Sanitize(raw)
		if err != nil {
			return
		}

		twice, err := richtext.DocumentSchema.Sanitize(once)
		require.NoError(t, err)
		assert.Equal(t, once, twice)

		require.NoError(t, richtext.DocumentSchema.ValidateJSON(once))

		if strings.TrimSpace(once) == "" {
			return
		}

		var n prosemirror.Node
		require.NoError(t, json.Unmarshal([]byte(once), &n))
		assert.Equal(t, prosemirror.NodeDoc, n.Type)
	})
}
