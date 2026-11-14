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

package automerge_test

import (
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

const blockElement = 0xfffc

func newSpansDocument(t *testing.T, spans []automerge.SpanInput) (*automerge.Document, *automerge.Text) {
	t.Helper()

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	text, err := document.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.UpdateSpans(spans, automerge.UpdateSpansConfig{}))

	_, err = document.CommitNow("seed")
	require.NoError(t, err)

	return document, text
}

func paragraphBlock() automerge.SpanInput {
	return automerge.SpanInput{
		Block: map[string]any{
			"type":    "paragraph",
			"parents": []any{},
			"attrs":   map[string]any{},
			"isEmbed": false,
		},
	}
}

func headingBlock() automerge.SpanInput {
	return automerge.SpanInput{
		Block: map[string]any{
			"type":    "heading",
			"parents": []any{},
			"attrs":   map[string]any{"level": float64(2)},
			"isEmbed": false,
		},
	}
}

func cellBlock(attrs map[string]any) automerge.SpanInput {
	return automerge.SpanInput{
		Block: map[string]any{
			"type":    "table-cell",
			"parents": []any{"table", "table-row"},
			"attrs":   attrs,
			"isEmbed": false,
		},
	}
}

func TestUpdateSpans_NullBlockAttributeEqualsOmittedAttribute(t *testing.T) {
	t.Parallel()

	document, text := newSpansDocument(
		t,
		[]automerge.SpanInput{
			cellBlock(map[string]any{"colspan": float64(1), "colwidth": nil, "rowspan": float64(1)}),
			{Text: "Right"},
			cellBlock(map[string]any{"colspan": float64(1), "colwidth": nil, "rowspan": float64(1)}),
			{Text: "Access"},
		},
	)

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(
		t,
		text.UpdateSpans(
			[]automerge.SpanInput{
				cellBlock(map[string]any{"colspan": float64(1), "rowspan": float64(1)}),
				{Text: "Right!"},
				cellBlock(map[string]any{"colspan": float64(1), "rowspan": float64(1)}),
				{Text: "Access"},
			},
			automerge.UpdateSpansConfig{},
		),
	)

	_, err = document.CommitNow("edit")
	require.NoError(t, err)

	after, err := document.Heads()
	require.NoError(t, err)

	patches, err := document.Diff(before, after)
	require.NoError(t, err)

	require.Len(t, patches, 1, "only the text edit may produce a patch")
	assert.Equal(t, automerge.PatchSpliceText, patches[0].Action)
	assert.Equal(t, "!", patches[0].Text)
}

func TestUpdateSpans_NestedNullAttributeIsNotDropped(t *testing.T) {
	t.Parallel()

	meta := func(note any) automerge.SpanInput {
		attrs := map[string]any{"meta": map[string]any{}}
		if note != struct{}{} {
			attrs["meta"] = map[string]any{"note": note}
		}

		return automerge.SpanInput{
			Block: map[string]any{
				"type":    "paragraph",
				"parents": []any{},
				"attrs":   attrs,
				"isEmbed": false,
			},
		}
	}

	document, text := newSpansDocument(
		t,
		[]automerge.SpanInput{
			meta(nil),
			{Text: "Hi"},
		},
	)

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(t, text.UpdateSpans(
		[]automerge.SpanInput{
			meta(struct{}{}),
			{Text: "Hi"},
		},
		automerge.UpdateSpansConfig{},
	))

	_, err = document.CommitNow("edit")
	require.NoError(t, err)

	after, err := document.Heads()
	require.NoError(t, err)

	patches, err := document.Diff(before, after)
	require.NoError(t, err)
	assert.NotEmpty(t, patches, "a nested null that becomes absent is a real block change")
}

func TestDiff_IndexesStayInUTF16UnitsAfterMixedInsertRun(t *testing.T) {
	t.Parallel()

	document, text := newSpansDocument(
		t,
		[]automerge.SpanInput{
			paragraphBlock(),
			{Text: "start"},
			paragraphBlock(),
			{Text: "tail"},
		},
	)

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(
		t,
		text.UpdateSpans(
			[]automerge.SpanInput{
				paragraphBlock(),
				{Text: "start"},
				headingBlock(),
				{Text: "😀 smile"},
				paragraphBlock(),
				{Text: "tail!"},
			},
			automerge.UpdateSpansConfig{},
		),
	)

	_, err = document.CommitNow("edit")
	require.NoError(t, err)

	after, err := document.Heads()
	require.NoError(t, err)

	patches, err := document.Diff(before, after)
	require.NoError(t, err)

	id, err := text.ObjectID()
	require.NoError(t, err)

	elements := elementsAt(t, text, before)

	for _, patch := range patches {
		if patch.Object != id {
			continue
		}

		switch patch.Action {
		case automerge.PatchSpliceText:
			inserted := utf16.Encode([]rune(patch.Text))
			elements = append(elements[:patch.Index], append(inserted, elements[patch.Index:]...)...)
		case automerge.PatchInsert:
			var inserted []uint16

			for _, value := range patch.Values {
				if value.Value.Scalar != nil && value.Value.Scalar.Type == automerge.ScalarTypeString {
					inserted = append(inserted, utf16.Encode([]rune(value.Value.Scalar.String))...)

					continue
				}

				inserted = append(inserted, blockElement)
			}

			elements = append(elements[:patch.Index], append(inserted, elements[patch.Index:]...)...)
		case automerge.PatchDeleteSeq:
			elements = append(elements[:patch.Index], elements[patch.Index+patch.Length:]...)
		}
	}

	assert.Equal(t, elementsAt(t, text, after), elements)
}

func elementsAt(t *testing.T, text *automerge.Text, heads []automerge.Hash) []uint16 {
	t.Helper()

	spans, err := text.SpansAt(heads)
	require.NoError(t, err)

	elements := make([]uint16, 0)

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			elements = append(elements, blockElement)

			continue
		}

		elements = append(elements, utf16.Encode([]rune(span.Text))...)
	}

	return elements
}
