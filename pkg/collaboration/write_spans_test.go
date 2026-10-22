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

package collaboration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestWriteSpans_LargeDocument(t *testing.T) {
	t.Parallel()

	const paragraphs = 2000

	var content strings.Builder
	content.WriteString(`{"type":"doc","content":[`)

	for i := range paragraphs {
		if i > 0 {
			content.WriteByte(',')
		}

		fmt.Fprintf(
			&content,
			`{"type":"paragraph","content":[{"type":"text","text":"paragraph %d"}]}`,
			i,
		)
	}

	content.WriteString(`]}`)

	document, err := newAutomergeDocument(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	started := time.Now()

	require.NoError(t, writeSpans(document, content.String(), true))

	elapsed := time.Since(started)

	assert.Less(t, elapsed, 15*time.Second)

	got, err := materialize(context.Background(), document)
	require.NoError(t, err)

	want, err := prosemirror.SanitizeDocumentJSON(content.String())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
