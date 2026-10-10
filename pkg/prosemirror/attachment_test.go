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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestDropRemoteImageSrcs(t *testing.T) {
	t.Parallel()

	node, err := Parse(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/a.png","alt":"diagram"}}]}`)
	require.NoError(t, err)

	DropRemoteImageSrcs(&node)

	attrs, err := node.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.Empty(t, attrs.Src)
}

func TestEmbedImageDataURLs(t *testing.T) {
	t.Parallel()

	node, err := Parse(`{"type":"doc","content":[{"type":"image","attrs":{"fileId":"file-1","src":""}}]}`)
	require.NoError(t, err)

	EmbedImageDataURLs(&node, map[string]string{
		"file-1": "data:image/png;base64,aaaa",
	})

	attrs, err := node.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.Nil(t, attrs.FileID)
	require.Equal(t, "data:image/png;base64,aaaa", attrs.Src)
}

func TestParseMarkdown_RestoresUploadedImageAndFile(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	path := AttachmentPath(fileID)

	doc, err := ParseMarkdown("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")\n\nsee [policy.pdf](" + path + ") here")
	require.NoError(t, err)
	require.Len(t, doc.Content, 3)

	imageAttrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, imageAttrs.FileID)
	require.Equal(t, fileID, *imageAttrs.FileID)
	require.Empty(t, imageAttrs.Src)

	require.Equal(t, NodeFile, doc.Content[1].Type)
	fileAttrs, err := doc.Content[1].FileAttrs()
	require.NoError(t, err)
	require.Equal(t, fileID, fileAttrs.FileID)
	require.Equal(t, "policy.pdf", fileAttrs.FileName)

	require.Equal(t, NodeParagraph, doc.Content[2].Type)

	short, err := ParseMarkdown("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")")
	require.NoError(t, err)
	require.Len(t, short.Content, 2)
	require.Equal(t, NodeFile, short.Content[1].Type)

	absolute, err := ParseMarkdown("![diagram](https://app.example/api/files/v1/" + fileID + ")")
	require.NoError(t, err)
	require.Len(t, absolute.Content, 1)
	absoluteAttrs, err := absolute.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, absoluteAttrs.FileID)
	require.Equal(t, fileID, *absoluteAttrs.FileID)
}

func TestAttachmentFromParagraph_IgnoresSurroundingSpace(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(LinkAttrs{Href: AttachmentPath(fileID)})
	require.NoError(t, err)

	space := " "
	name := "policy.pdf"
	paragraph := Node{
		Type: NodeParagraph,
		Content: []Node{
			{Type: NodeText, Text: &space},
			{Type: NodeText, Text: &name, Marks: []Mark{{Type: MarkLink, Attrs: href}}},
			{Type: NodeText, Text: &space},
		},
	}

	file, ok := attachmentFromParagraph(paragraph)
	require.True(t, ok)

	attrs, err := file.FileAttrs()
	require.NoError(t, err)
	require.Equal(t, fileID, attrs.FileID)
	require.Equal(t, "policy.pdf", attrs.FileName)
}
