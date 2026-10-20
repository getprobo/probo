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
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestParseAttachmentID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	path := AttachmentPath(fileID)

	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{name: "path", raw: path, ok: true},
		{name: "https", raw: "https://files.example.com" + path, ok: true},
		{name: "http", raw: "http://files.example.com" + path, ok: true},
		{name: "file scheme", raw: "file://" + path, ok: false},
		{name: "javascript scheme", raw: "javascript:" + path, ok: false},
		{name: "protocol relative", raw: "//evil.example" + path, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseAttachmentID(tt.raw)
			assert.Equal(t, tt.ok, ok)

			if tt.ok {
				assert.Equal(t, fileID, got)

				return
			}

			assert.Empty(t, got)
		})
	}
}

func TestIsImageMime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mime string
		ok   bool
	}{
		{name: "png", mime: "image/png", ok: true},
		{name: "uppercase", mime: "IMAGE/PNG", ok: true},
		{name: "padded mixed case", mime: " Image/Jpeg ", ok: true},
		{name: "pdf", mime: "application/pdf", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.ok, IsImageMime(tt.mime))
		})
	}
}

func TestDropRemoteImageSrcs(t *testing.T) {
	t.Parallel()

	node, err := Parse(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/a.png","alt":"diagram"}}]}`)
	require.NoError(t, err)

	DropRemoteImageSrcs(&node)

	attrs, err := node.Content[0].ImageAttrs()
	require.NoError(t, err)
	assert.Empty(t, attrs.Src)
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
	assert.Nil(t, attrs.FileID)
	assert.Equal(t, "data:image/png;base64,aaaa", attrs.Src)
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
	assert.Equal(t, fileID, *imageAttrs.FileID)
	assert.Empty(t, imageAttrs.Src)

	assert.Equal(t, NodeFile, doc.Content[1].Type)
	fileAttrs, err := doc.Content[1].FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, fileAttrs.FileID)
	assert.Equal(t, "policy.pdf", fileAttrs.FileName)

	assert.Equal(t, NodeParagraph, doc.Content[2].Type)

	short, err := ParseMarkdown("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")")
	require.NoError(t, err)
	require.Len(t, short.Content, 2)
	assert.Equal(t, NodeFile, short.Content[1].Type)

	cellDoc, err := ParseMarkdown("| File |\n| --- |\n| [policy.pdf](" + path + ") |\n")
	require.NoError(t, err)
	require.NoError(t, validateDocument(cellDoc))
	cell := cellDoc.Content[0].Content[1].Content[0]
	require.Len(t, cell.Content, 1)
	assert.Equal(t, NodeFile, cell.Content[0].Type)
	cellAttrs, err := cell.Content[0].FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, cellAttrs.FileID)

	refs := ReferencedFiles(cellDoc)
	require.Len(t, refs, 1)
	assert.Equal(t, fileID, refs[0].ID)

	absolute, err := ParseMarkdown("![diagram](https://app.example/api/files/v1/" + fileID + ")")
	require.NoError(t, err)
	require.Len(t, absolute.Content, 1)
	absoluteAttrs, err := absolute.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, absoluteAttrs.FileID)
	assert.Equal(t, fileID, *absoluteAttrs.FileID)
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
	assert.Equal(t, fileID, attrs.FileID)
	assert.Equal(t, "policy.pdf", attrs.FileName)
}

func TestAttachmentFromParagraph_KeepsSpaceBetweenLinks(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(LinkAttrs{Href: AttachmentPath(fileID)})
	require.NoError(t, err)

	left := "report"
	space := " "
	right := "final.pdf"
	paragraph := Node{
		Type: NodeParagraph,
		Content: []Node{
			{Type: NodeText, Text: &left, Marks: []Mark{{Type: MarkLink, Attrs: href}}},
			{Type: NodeText, Text: &space},
			{Type: NodeText, Text: &right, Marks: []Mark{{Type: MarkLink, Attrs: href}}},
		},
	}

	file, ok := attachmentFromParagraph(paragraph)
	require.True(t, ok)

	attrs, err := file.FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, attrs.FileID)
	assert.Equal(t, "report final.pdf", attrs.FileName)
}

func TestAttachmentFromParagraph_TruncatesFileNameOnRuneBoundary(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(LinkAttrs{Href: AttachmentPath(fileID)})
	require.NoError(t, err)

	name := strings.Repeat("a", 1023) + "é"
	paragraph := Node{
		Type: NodeParagraph,
		Content: []Node{
			{Type: NodeText, Text: &name, Marks: []Mark{{Type: MarkLink, Attrs: href}}},
		},
	}

	file, ok := attachmentFromParagraph(paragraph)
	require.True(t, ok)

	attrs, err := file.FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 1023), attrs.FileName)
	require.True(t, utf8.ValidString(attrs.FileName))
	assert.LessOrEqual(t, len(attrs.FileName), 1024)
}
