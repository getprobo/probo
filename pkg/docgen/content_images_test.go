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
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

type stubImageReader struct {
	data string
	mime string
}

func (s stubImageReader) GetFileBase64(context.Context, *coredata.File) (string, string, error) {
	return s.data, s.mime, nil
}

func TestContentImageFiles_LogsSkippedImage(t *testing.T) {
	t.Parallel()

	var logged bytes.Buffer

	logger := log.NewLogger(log.WithOutput(&logged))
	content := `{"type":"doc","content":[{"type":"image","attrs":{"fileId":"not-a-gid"}}]}`

	files := ContentImageFiles(context.Background(), nil, nil, content, logger)

	assert.Empty(t, files)
	assert.Contains(t, logged.String(), "cannot parse image file id")
}

func TestEmbedContentImages_LogsUnreadableContent(t *testing.T) {
	t.Parallel()

	var logged bytes.Buffer

	logger := log.NewLogger(log.WithOutput(&logged))
	got := EmbedContentImages(context.Background(), logger, "{", nil, nil)

	assert.Equal(t, "{", got)
	assert.Contains(t, logged.String(), "cannot parse document content for PDF images")
}

func TestEmbedContentImages(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger(log.WithOutput(io.Discard))
	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
	reader := stubImageReader{data: "aaaa", mime: "image/png"}
	files := []*coredata.File{{ID: fileID}}

	t.Run("embeds a known file and drops a remote source", func(t *testing.T) {
		t.Parallel()

		content := fmt.Sprintf(
			`{"type":"doc","content":[{"type":"image","attrs":{"src":"","fileId":%q,"alt":null,"title":null}},{"type":"image","attrs":{"src":"https://example.com/a.png","alt":null,"title":null}}]}`,
			fileID.String(),
		)

		out := EmbedContentImages(context.Background(), logger, content, files, reader)

		node, err := prosemirror.Parse(out)
		require.NoError(t, err)
		require.Len(t, node.Content, 2)

		uploaded, err := node.Content[0].ImageAttrs()
		require.NoError(t, err)
		assert.Nil(t, uploaded.FileID)
		assert.Equal(t, "data:image/png;base64,aaaa", uploaded.Src)

		remote, err := node.Content[1].ImageAttrs()
		require.NoError(t, err)
		assert.Empty(t, remote.Src)
	})

	t.Run("leaves an unresolved file id", func(t *testing.T) {
		t.Parallel()

		missingID := gid.New(gid.NewTenantID(), coredata.FileEntityType)
		content := fmt.Sprintf(
			`{"type":"doc","content":[{"type":"image","attrs":{"src":"","fileId":%q,"alt":null,"title":null}}]}`,
			missingID.String(),
		)

		out := EmbedContentImages(context.Background(), logger, content, files, reader)

		node, err := prosemirror.Parse(out)
		require.NoError(t, err)
		require.Len(t, node.Content, 1)

		attrs, err := node.Content[0].ImageAttrs()
		require.NoError(t, err)
		require.NotNil(t, attrs.FileID)
		assert.Equal(t, missingID.String(), *attrs.FileID)
		assert.Empty(t, attrs.Src)
	})

	t.Run("returns malformed content unchanged", func(t *testing.T) {
		t.Parallel()

		content := "not json"

		out := EmbedContentImages(context.Background(), logger, content, files, reader)

		assert.Equal(t, content, out)
	})

	t.Run("embeds an image nested in a paragraph", func(t *testing.T) {
		t.Parallel()

		content := fmt.Sprintf(
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"image","attrs":{"src":"","fileId":%q,"alt":null,"title":null}}]}]}`,
			fileID.String(),
		)

		out := EmbedContentImages(context.Background(), logger, content, files, reader)

		node, err := prosemirror.Parse(out)
		require.NoError(t, err)
		require.Len(t, node.Content, 1)
		require.Len(t, node.Content[0].Content, 1)

		attrs, err := node.Content[0].Content[0].ImageAttrs()
		require.NoError(t, err)
		assert.Nil(t, attrs.FileID)
		assert.Equal(t, "data:image/png;base64,aaaa", attrs.Src)
	})
}
