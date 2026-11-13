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

package richtext

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/richtext/attachment"
)

func TestParseMarkdown_RestoresUploadedImageAndFile(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	path := attachment.Path(fileID)

	doc, err := ParseMarkdown("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")\n\nsee [policy.pdf](" + path + ") here")
	require.NoError(t, err)
	require.Len(t, doc.Content, 3)

	imageAttrs, err := doc.Content[0].ImageAttrs()
	require.NoError(t, err)
	require.NotNil(t, imageAttrs.FileID)
	assert.Equal(t, fileID, *imageAttrs.FileID)
	assert.Empty(t, imageAttrs.Src)

	assert.Equal(t, prosemirror.NodeFile, doc.Content[1].Type)
	fileAttrs, err := doc.Content[1].FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, fileAttrs.FileID)
	assert.Equal(t, "policy.pdf", fileAttrs.FileName)

	assert.Equal(t, prosemirror.NodeParagraph, doc.Content[2].Type)

	short, err := ParseMarkdown("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")")
	require.NoError(t, err)
	require.Len(t, short.Content, 2)
	assert.Equal(t, prosemirror.NodeFile, short.Content[1].Type)

	cellDoc, err := ParseMarkdown("| File |\n| --- |\n| [policy.pdf](" + path + ") |\n")
	require.NoError(t, err)
	require.NoError(t, DocumentSchema.Validate(cellDoc))
	cell := cellDoc.Content[0].Content[1].Content[0]
	require.Len(t, cell.Content, 1)
	assert.Equal(t, prosemirror.NodeFile, cell.Content[0].Type)
	cellAttrs, err := cell.Content[0].FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, cellAttrs.FileID)

	refs := attachment.Referenced(cellDoc)
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
