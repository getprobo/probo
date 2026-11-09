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

package attachment_test

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/richtext/attachment"
)

func TestAttachmentFromParagraph_IgnoresSurroundingSpace(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(prosemirror.LinkAttrs{Href: attachment.Path(fileID)})
	require.NoError(t, err)

	paragraph := prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{Type: prosemirror.NodeText, Text: " "},
			{
				Type: prosemirror.NodeText,
				Text: "policy.pdf",
				Marks: []prosemirror.Mark{
					{Type: prosemirror.MarkLink, Attrs: href},
				},
			},
			{Type: prosemirror.NodeText, Text: " "},
		},
	}

	file, ok := attachment.AttachmentFromParagraph(paragraph)
	require.True(t, ok)
	attrs, err := file.FileAttrs()
	require.NoError(t, err)
	assert.Equal(t, fileID, attrs.FileID)
}

func TestAttachmentFromParagraph_RejectsPartialLinkText(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(prosemirror.LinkAttrs{Href: attachment.Path(fileID)})
	require.NoError(t, err)

	paragraph := prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{Type: prosemirror.NodeText, Text: "see "},
			{
				Type: prosemirror.NodeText,
				Text: "policy.pdf",
				Marks: []prosemirror.Mark{
					{Type: prosemirror.MarkLink, Attrs: href},
				},
			},
		},
	}

	_, ok := attachment.AttachmentFromParagraph(paragraph)
	assert.False(t, ok)
}

func TestAttachmentFromParagraph_RejectsNonFileHref(t *testing.T) {
	t.Parallel()

	href, err := json.Marshal(prosemirror.LinkAttrs{Href: "https://example.com/file.pdf"})
	require.NoError(t, err)

	paragraph := prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeText,
				Text: "file.pdf",
				Marks: []prosemirror.Mark{
					{Type: prosemirror.MarkLink, Attrs: href},
				},
			},
		},
	}

	_, ok := attachment.AttachmentFromParagraph(paragraph)
	assert.False(t, ok)
}

func TestAttachmentFromParagraph_RejectsEmptyFileName(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	href, err := json.Marshal(prosemirror.LinkAttrs{Href: attachment.Path(fileID)})
	require.NoError(t, err)

	paragraph := prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeText,
				Text: strings.Repeat(" ", utf8.UTFMax),
				Marks: []prosemirror.Mark{
					{Type: prosemirror.MarkLink, Attrs: href},
				},
			},
		},
	}

	_, ok := attachment.AttachmentFromParagraph(paragraph)
	assert.False(t, ok)
}
