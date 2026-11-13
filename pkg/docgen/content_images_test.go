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

package docgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
)

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

			assert.Equal(t, tt.ok, isImageMime(tt.mime))
		})
	}
}

func TestDropRemoteImageSrcs(t *testing.T) {
	t.Parallel()

	node, err := prosemirror.Parse(`{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com/a.png","alt":"diagram"}}]}`)
	require.NoError(t, err)

	dropRemoteImageSrcs(&node)

	attrs, err := node.Content[0].ImageAttrs()
	require.NoError(t, err)
	assert.Empty(t, attrs.Src)
}

func TestEmbedImageDataURLs(t *testing.T) {
	t.Parallel()

	node, err := prosemirror.Parse(`{"type":"doc","content":[{"type":"image","attrs":{"fileId":"file-1","src":""}}]}`)
	require.NoError(t, err)

	embedImageDataURLs(&node, map[string]string{
		"file-1": "data:image/png;base64,aaaa",
	})

	attrs, err := node.Content[0].ImageAttrs()
	require.NoError(t, err)
	assert.Nil(t, attrs.FileID)
	assert.Equal(t, "data:image/png;base64,aaaa", attrs.Src)
}
