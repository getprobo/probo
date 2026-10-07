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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestRecoverDocument_UnwrapsLiteralDiagramRequest(t *testing.T) {
	t.Parallel()

	raw := mermaidDiagramRequestDoc(t)
	formatted := "Add three Mermaid diagrams\n\n- Network\n- Data flow\n- SDLC\n"

	t.Run("document json", func(t *testing.T) {
		t.Parallel()

		doc, err := prosemirror.RecoverDocument(raw)
		require.NoError(t, err)

		markdown, err := prosemirror.RenderMarkdown(doc)
		require.NoError(t, err)
		assert.Equal(t, formatted, markdown)
		assert.NotContains(t, markdown, "bulletList")
	})

	t.Run("json string", func(t *testing.T) {
		t.Parallel()

		encoded, err := json.Marshal(raw)
		require.NoError(t, err)

		doc, err := prosemirror.RecoverDocument(string(encoded))
		require.NoError(t, err)

		markdown, err := prosemirror.RenderMarkdown(doc)
		require.NoError(t, err)
		assert.Equal(t, formatted, markdown)
	})

	t.Run("paragraphs of pretty json", func(t *testing.T) {
		t.Parallel()

		var pretty bytes.Buffer

		require.NoError(t, json.Indent(&pretty, []byte(raw), "", "  "))

		doc, err := prosemirror.RecoverDocument(prosemirror.FromPlainText(pretty.String()))
		require.NoError(t, err)

		markdown, err := prosemirror.RenderMarkdown(doc)
		require.NoError(t, err)
		assert.Equal(t, formatted, markdown)
		assert.NotContains(t, markdown, `"type"`)
	})
}

func TestRecoverDocument_KeepsRealBulletList(t *testing.T) {
	t.Parallel()

	doc, err := prosemirror.ParseMarkdown("- Network\n- Data flow\n- SDLC")
	require.NoError(t, err)

	encoded, err := json.Marshal(doc)
	require.NoError(t, err)

	recovered, err := prosemirror.RecoverDocument(string(encoded))
	require.NoError(t, err)
	require.NotEmpty(t, recovered.Content)
	assert.Equal(t, prosemirror.NodeBulletList, recovered.Content[0].Type)

	markdown, err := prosemirror.RenderMarkdown(recovered)
	require.NoError(t, err)
	assert.Contains(t, markdown, "- Network")
	assert.Contains(t, markdown, "- Data flow")
	assert.Contains(t, markdown, "- SDLC")
}

func TestRecoverDocument_KeepsProse(t *testing.T) {
	t.Parallel()

	doc, err := prosemirror.RecoverDocument(prosemirror.FromPlainText("Keep this sentence"))
	require.NoError(t, err)

	markdown, err := prosemirror.RenderMarkdown(doc)
	require.NoError(t, err)
	assert.Equal(t, "Keep this sentence\n", markdown)
}

func TestRecoverDocument_RejectsPlainText(t *testing.T) {
	t.Parallel()

	_, err := prosemirror.RecoverDocument("not json")
	require.Error(t, err)
}

func mermaidDiagramRequestDoc(t *testing.T) string {
	t.Helper()

	doc := prosemirror.Node{
		Type: prosemirror.NodeDoc,
		Content: []prosemirror.Node{
			mermaidParagraph("Add three Mermaid diagrams"),
			{
				Type: prosemirror.NodeBulletList,
				Content: []prosemirror.Node{
					mermaidItem("Network"),
					mermaidItem("Data flow"),
					mermaidItem("SDLC"),
				},
			},
		},
	}

	encoded, err := json.Marshal(doc)
	require.NoError(t, err)

	return string(encoded)
}

func mermaidParagraph(text string) prosemirror.Node {
	value := text

	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{{
			Type: prosemirror.NodeText,
			Text: &value,
		}},
	}
}

func mermaidItem(text string) prosemirror.Node {
	return prosemirror.Node{
		Type:    prosemirror.NodeListItem,
		Content: []prosemirror.Node{mermaidParagraph(text)},
	}
}
