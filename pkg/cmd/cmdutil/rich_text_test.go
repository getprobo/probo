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

package cmdutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestFormatRichText(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.FormatRichText("")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("prosemirror json", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.FormatRichText(prosemirror.FromPlainText("Review access controls"))
		require.NoError(t, err)
		assert.Equal(t, "Review access controls", got)
	})

	t.Run("empty document", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.FormatRichText(prosemirror.FromPlainText(""))
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		_, err := cmdutil.FormatRichText("not json")
		require.Error(t, err)
	})
}

func TestCLIContent(t *testing.T) {
	t.Parallel()

	t.Run("plain text", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.CLIContent("Review access controls")
		require.NoError(t, err)
		assert.Equal(t, prosemirror.FromPlainText("Review access controls"), got)
	})

	t.Run("json passthrough", func(t *testing.T) {
		t.Parallel()

		raw := "  " + prosemirror.FromPlainText("kept")
		got, err := cmdutil.CLIContent(raw)
		require.NoError(t, err)
		assert.Equal(t, raw, got)
	})

	t.Run("markdown image and file", func(t *testing.T) {
		t.Parallel()

		fileID := gid.New(gid.TenantID{}, 25).String()
		path := prosemirror.AttachmentPath(fileID)
		got, err := cmdutil.CLIContent("![diagram](" + path + ")\n\n[policy.pdf](" + path + ")")
		require.NoError(t, err)

		doc, err := prosemirror.Parse(got)
		require.NoError(t, err)
		require.Len(t, doc.Content, 2)
		assert.Equal(t, prosemirror.NodeImage, doc.Content[0].Type)
		assert.Equal(t, prosemirror.NodeFile, doc.Content[1].Type)
	})
}
