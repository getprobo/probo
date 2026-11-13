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
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

func highlightParagraph(text string) string {
	return fmt.Sprintf(`{"type":"paragraph","content":[{"type":"text","text":%q}]}`, text)
}

func highlightHeading(level int, text string) string {
	return fmt.Sprintf(`{"type":"heading","attrs":{"level":%d},"content":[{"type":"text","text":%q}]}`, level, text)
}

func highlightTable(rows [][]string) string {
	var table strings.Builder

	table.WriteString(`{"type":"table","content":[`)

	for i, row := range rows {
		if i > 0 {
			table.WriteByte(',')
		}

		table.WriteString(`{"type":"tableRow","content":[`)

		for j, cell := range row {
			if j > 0 {
				table.WriteByte(',')
			}

			kind := "tableCell"
			if i == 0 {
				kind = "tableHeader"
			}

			fmt.Fprintf(&table, `{"type":%q,"content":[%s]}`, kind, highlightParagraph(cell))
		}

		table.WriteString(`]}`)
	}

	table.WriteString(`]}`)

	return table.String()
}

func highlightDocument(blocks ...string) string {
	return `{"type":"doc","content":[` + strings.Join(blocks, ",") + `]}`
}

func editedDocument(
	t *testing.T,
	seed string,
	edited string,
	browserAuthored bool,
) (*automerge.Document, []automerge.Hash, []automerge.Hash) {
	t.Helper()

	document, err := newAutomergeDocument(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	if browserAuthored {
		spans, err := automergeprosemirror.ToSpans(seed)
		require.NoError(t, err)

		for i := range spans {
			attrs, ok := spans[i].Block["attrs"].(map[string]any)
			if !ok {
				continue
			}

			if _, ok := attrs["colspan"]; ok {
				attrs["colwidth"] = nil
			}
		}

		text, err := document.CreateText("body")
		require.NoError(t, err)
		require.NoError(t, text.UpdateSpans(spans, automergeprosemirror.UpdateSpansConfig()))
	} else {
		require.NoError(t, writeSpans(document, seed, true))
	}

	_, err = document.Commit("seed", time.Now())
	require.NoError(t, err)

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(t, writeSpans(document, edited, false))

	_, err = document.Commit("edit", time.Now())
	if !errors.Is(err, automerge.ErrNoOperations) {
		require.NoError(t, err)
	}

	after, err := document.Heads()
	require.NoError(t, err)

	return document, before, after
}

// reconstructBefore replays the highlights over the after document and
// returns the text the before document held.
func reconstructBefore(
	t *testing.T,
	document *automerge.Document,
	after []automerge.Hash,
	highlights []DocumentHistoryHighlight,
) string {
	t.Helper()

	text, err := document.Text("body")
	require.NoError(t, err)

	spans, err := text.SpansAt(after)
	require.NoError(t, err)

	elements := originalElements(spans)

	var rebuilt []uint16

	emitDeleted := func(at int) {
		for _, highlight := range highlights {
			if highlight.Kind == DocumentHistoryHighlightKindDelete && highlight.From == at {
				rebuilt = append(rebuilt, utf16.Encode([]rune(highlight.Text))...)
			}
		}
	}

	for index, element := range elements {
		emitDeleted(index)

		inserted := false

		for _, highlight := range highlights {
			if highlight.Kind == DocumentHistoryHighlightKindInsert &&
				index >= highlight.From && index < highlight.To {
				inserted = true
			}
		}

		if inserted || element == originalBlockMarker {
			continue
		}

		rebuilt = append(rebuilt, uint16(element))
	}

	emitDeleted(len(elements))

	return string(utf16.Decode(rebuilt))
}

func plainTextAt(
	t *testing.T,
	document *automerge.Document,
	heads []automerge.Hash,
) string {
	t.Helper()

	text, err := document.Text("body")
	require.NoError(t, err)

	spans, err := text.SpansAt(heads)
	require.NoError(t, err)

	var plain strings.Builder

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText {
			plain.WriteString(span.Text)
		}
	}

	return plain.String()
}

func TestHistoryHighlights_EditAfterBlocksStaysOnTheEditedText(t *testing.T) {
	t.Parallel()

	rows := [][]string{
		{"Right", "What it means"},
		{"Access", "Obtain confirmation and a copy of your personal data."},
		{"Erasure", "Have your data deleted where the conditions apply."},
	}

	seed := highlightDocument(
		highlightHeading(2, "Your rights"),
		highlightParagraph("Where we act as controller, you have the right to:"),
		highlightTable(rows),
		highlightParagraph("Contact privacy@wafer.ai."),
	)
	edited := highlightDocument(
		highlightHeading(2, "Your rights and more"),
		highlightParagraph("Where we act as controller, you have the right to:"),
		highlightTable(rows),
		highlightParagraph("Contact privacy@wafer.ai or use the form."),
	)

	for _, browserAuthored := range []bool{false, true} {
		t.Run(
			fmt.Sprintf("browser authored %t", browserAuthored),
			func(t *testing.T) {
				t.Parallel()

				document, before, after := editedDocument(t, seed, edited, browserAuthored)

				highlights, err := historyHighlights(document, before, after)
				require.NoError(t, err)

				require.Len(t, highlights, 2)
				assert.Equal(t, DocumentHistoryHighlightKindInsert, highlights[0].Kind)
				assert.Equal(t, " and more", highlights[0].Text)
				assert.Equal(t, 12, highlights[0].From, "the heading block marker counts as one position")
				assert.Equal(t, 21, highlights[0].To)
				assert.Equal(t, DocumentHistoryHighlightKindInsert, highlights[1].Kind)
				assert.Equal(t, " or use the form", highlights[1].Text)

				assert.Equal(
					t,
					plainTextAt(t, document, before),
					reconstructBefore(t, document, after, highlights),
				)
			},
		)
	}
}

func TestHistoryHighlights_UnchangedBrowserBlocksAreNotRewritten(t *testing.T) {
	t.Parallel()

	rows := [][]string{{"Right", "What it means"}, {"Access", "Obtain a copy."}}

	seed := highlightDocument(highlightParagraph("Intro"), highlightTable(rows))
	edited := highlightDocument(highlightParagraph("Intro text"), highlightTable(rows))

	document, before, after := editedDocument(t, seed, edited, true)

	patches, err := document.Diff(before, after)
	require.NoError(t, err)

	for _, patch := range patches {
		assert.NotEqual(t, automerge.PatchDeleteSeq, patch.Action, "no block marker may be deleted")
		assert.NotEqual(t, automerge.PatchInsert, patch.Action, "no block marker may be inserted")
	}
}

func TestHistoryHighlights_ReplayReconstructsTheBeforeText(t *testing.T) {
	t.Parallel()

	words := []string{"alpha", "beta", "HIPAA", "ePHI", "Right", "Access", "data", "the", "of", "BAA", "é", "日本", "😀"}

	randomText := func(r *rand.Rand) string {
		parts := make([]string, 1+r.Intn(8))
		for i := range parts {
			parts[i] = words[r.Intn(len(words))]
		}

		return strings.Join(parts, " ")
	}

	randomBlock := func(r *rand.Rand) string {
		switch r.Intn(4) {
		case 0:
			return highlightHeading(1+r.Intn(3), randomText(r))
		case 1:
			rows := make([][]string, 1+r.Intn(3))
			for i := range rows {
				rows[i] = []string{randomText(r), randomText(r)}
			}

			return highlightTable(rows)
		default:
			return highlightParagraph(randomText(r))
		}
	}

	for seed := range int64(60) {
		t.Run(
			fmt.Sprintf("seed %d", seed),
			func(t *testing.T) {
				t.Parallel()

				r := rand.New(rand.NewSource(seed))

				blocks := make([]string, 3+r.Intn(8))
				for i := range blocks {
					blocks[i] = randomBlock(r)
				}

				edited := append([]string(nil), blocks...)

				for range 1 + r.Intn(3) {
					switch r.Intn(3) {
					case 0:
						edited[r.Intn(len(edited))] = randomBlock(r)
					case 1:
						at := r.Intn(len(edited) + 1)
						edited = append(edited[:at], append([]string{randomBlock(r)}, edited[at:]...)...)
					default:
						if len(edited) > 1 {
							at := r.Intn(len(edited))
							edited = append(edited[:at], edited[at+1:]...)
						}
					}
				}

				document, before, after := editedDocument(
					t,
					highlightDocument(blocks...),
					highlightDocument(edited...),
					seed%2 == 0,
				)

				highlights, err := historyHighlights(document, before, after)
				require.NoError(t, err)

				assert.Equal(
					t,
					plainTextAt(t, document, before),
					reconstructBefore(t, document, after, highlights),
				)
			},
		)
	}
}
