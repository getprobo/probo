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
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

// TestBridge_MatchesUpstreamInBothDirections is the differential parity gate for
// both halves of the ProseMirror bridge. The fixture is produced by the official
// @automerge/prosemirror library through the real frontend schema adapter (see
// packages/ui/src/RichEditor/prosemirrorRenderParity.test.ts). For each document,
// native Go must materialize the exact spans produced by pmNodeToSpans, and the Go
// renderer must turn those spans back into the canonical ProseMirror JSON produced
// by pmDocFromSpans.
func TestBridge_MatchesUpstreamInBothDirections(t *testing.T) {
	t.Parallel()

	file, err := os.Open("testdata/upstream-render.json.gz")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	reader, err := gzip.NewReader(file)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	var fixtures []struct {
		Name     string          `json:"name"`
		Document string          `json:"document"`
		Expected json.RawMessage `json:"expected"`
		Spans    json.RawMessage `json:"spans"`
	}
	require.NoError(t, json.NewDecoder(reader).Decode(&fixtures))
	require.NotEmpty(t, fixtures)

	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()

			data, err := base64.StdEncoding.DecodeString(fixture.Document)
			require.NoError(t, err)

			actorID, err := automerge.NewActorID()
			require.NoError(t, err)

			document, err := automerge.Load(context.Background(), data, actorID)
			require.NoError(t, err)

			defer func() { _ = document.Close(context.Background()) }()

			text, err := document.Text(context.Background(), "body")
			require.NoError(t, err)

			spans, err := text.Spans(context.Background())
			require.NoError(t, err)

			actualSpans := make([]map[string]any, 0, len(spans))
			for _, span := range spans {
				switch span.Type {
				case automerge.SpanTypeBlock:
					actualSpans = append(
						actualSpans,
						map[string]any{
							"type":  string(span.Type),
							"value": span.Block,
						},
					)
				case automerge.SpanTypeText:
					actual := map[string]any{
						"type":  string(span.Type),
						"value": span.Text,
					}
					if len(span.Marks) > 0 {
						actual["marks"] = span.Marks
					}

					actualSpans = append(actualSpans, actual)
				}
			}

			encodedSpans, err := json.Marshal(actualSpans)
			require.NoError(t, err)
			assert.JSONEq(
				t,
				string(fixture.Spans),
				string(encodedSpans),
				"native spans must match pmNodeToSpans for %s",
				fixture.Name,
			)

			rendered, err := automergeprosemirror.Render(spans)
			require.NoError(t, err)

			assert.JSONEq(t, string(fixture.Expected), rendered)
		})
	}
}
