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
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

// TestToSpans_RoundTripsCanonicalDocuments is the seeding-correctness gate: for
// every document in the shared corpus, converting the canonical ProseMirror JSON
// to spans, writing them into a fresh Automerge document, and rendering it back
// must reproduce the same ProseMirror JSON. This is exactly what server-side
// seeding does, so a green run means the server can bootstrap a document
// version's CRDT from its stored content and materialize it unchanged.
func TestToSpans_RoundTripsCanonicalDocuments(t *testing.T) {
	t.Parallel()

	file, err := os.Open("testdata/upstream-render.json.gz")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	reader, err := gzip.NewReader(file)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	var fixtures []struct {
		Name     string          `json:"name"`
		Expected json.RawMessage `json:"expected"`
	}
	require.NoError(t, json.NewDecoder(reader).Decode(&fixtures))
	require.NotEmpty(t, fixtures)

	for _, fixture := range fixtures {
		t.Run(
			fixture.Name,
			func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()

				spans, err := automergeprosemirror.ToSpans(string(fixture.Expected))
				require.NoError(t, err)

				actorID, err := automerge.NewActorID()
				require.NoError(t, err)

				document, err := automerge.New(ctx, actorID)
				require.NoError(t, err)

				defer func() { _ = document.Close(ctx) }()

				text, err := document.CreateText(ctx, "body")
				require.NoError(t, err)

				require.NoError(t, text.UpdateSpans(ctx, spans, automergeprosemirror.UpdateSpansConfig()))

				readback, err := text.Spans(ctx)
				require.NoError(t, err)

				rendered, err := automergeprosemirror.Render(readback)
				require.NoError(t, err)

				assert.JSONEq(
					t,
					string(fixture.Expected),
					rendered,
					"seeding %s must round-trip through spans",
					fixture.Name,
				)
			},
		)
	}
}
