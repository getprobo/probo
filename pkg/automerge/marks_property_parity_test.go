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

// This file reproduces the upstream marks_are_okay property test
// (rust/automerge/tests/text.rs). Like the upstream proptest, many random
// sequences of insert, delete, split-block, and mark operations are applied and
// the resulting spans must (1) stay consolidated (no two adjacent text spans
// carry an identical mark set) and (2) reproduce the accumulated text. The
// upstream test asserts these structural invariants rather than specific mark
// values or a differential against another engine, so both the native Go and
// Rust/WASM reference engines are checked against the same invariants here.

package automerge_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func marksAreConsolidated(spans []automerge.Span) bool {
	haveLast := false

	var last map[string]any

	for _, span := range spans {
		if span.Type != automerge.SpanTypeText {
			haveLast = false

			continue
		}

		if haveLast && sameMarkSet(last, span.Marks) {
			return false
		}

		last = span.Marks
		haveLast = true
	}

	return true
}

func sameMarkSet(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, value := range a {
		other, ok := b[key]
		if !ok || other != value {
			return false
		}
	}

	return true
}

// TestRustText_MarksAreOkay reproduces marks_are_okay across randomized scenarios.
func TestRustText_MarksAreOkay(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewSource(0x2545f4914f6cdd1d))

	const scenarios = 300

	markNames := []string{"bold", "italic", "underline"}

	for scenario := range scenarios {
		engineSpans := make(map[string][]automerge.Span)

		var expected []rune

		steps := 3 + random.Intn(18)
		actions := make([]func(*testing.T, *automerge.Text), 0, steps)

		length := 0

		for range steps {
			switch random.Intn(4) {
			case 0: // insert
				index := random.Intn(length + 1)
				value := randomLetters(random, 1+random.Intn(6))
				runes := []rune(value)

				actions = append(
					actions,
					func(t *testing.T, text *automerge.Text) {
						require.NoError(t, text.Splice(uint32(index), 0, value))
					},
				)

				expected = append(expected[:index], append(append([]rune{}, runes...), expected[index:]...)...)
				length += len(runes)
			case 1: // delete
				if length == 0 {
					continue
				}

				deleteLen := 1 + random.Intn(length)
				index := random.Intn(length - deleteLen + 1)

				actions = append(
					actions,
					func(t *testing.T, text *automerge.Text) {
						require.NoError(t, text.Splice(uint32(index), int32(deleteLen), ""))
					},
				)

				expected = append(expected[:index], expected[index+deleteLen:]...)
				length -= deleteLen
			case 2: // split block
				index := random.Intn(length + 1)

				actions = append(
					actions,
					func(t *testing.T, text *automerge.Text) {
						_, err := text.SplitBlock(uint32(index))
						require.NoError(t, err)
					},
				)

				expected = append(expected[:index], append([]rune{'\n'}, expected[index:]...)...)
				length++
			case 3: // add mark
				if length == 0 {
					continue
				}

				markLen := 1 + random.Intn(length)
				index := random.Intn(length - markLen + 1)
				name := markNames[random.Intn(len(markNames))]

				actions = append(
					actions,
					func(t *testing.T, text *automerge.Text) {
						require.NoError(
							t,
							text.Mark(

								uint32(index),
								uint32(index+markLen),
								name,
								automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
								automerge.MarkExpandBoth,
							),
						)
					},
				)
			}
		}

		for _, engine := range rustParityEngines() {
			document, err := engine.open(actor(1))
			require.NoError(t, err)
			closeDocument(t, document)

			text, err := document.CreateText("text")
			require.NoError(t, err)

			for _, action := range actions {
				action(t, text)
			}

			_, err = document.Commit("scenario", commitTime)
			require.NoError(t, err)

			spans, err := text.Spans()
			require.NoError(t, err)

			engineSpans[engine.name] = spans
		}

		for _, engine := range rustParityEngines() {
			spans := engineSpans[engine.name]

			require.True(
				t,
				marksAreConsolidated(spans),
				"scenario %d marks not consolidated on %s: %+v",
				scenario,
				engine.name,
				spans,
			)

			var builder strings.Builder

			for _, span := range spans {
				if span.Type == automerge.SpanTypeBlock {
					builder.WriteRune('\n')

					continue
				}

				builder.WriteString(span.Text)
			}

			require.Equal(
				t,
				string(expected),
				builder.String(),
				"scenario %d span text diverged on %s",
				scenario,
				engine.name,
			)
		}
	}
}

func randomLetters(random *rand.Rand, count int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var builder strings.Builder

	for range count {
		builder.WriteByte(alphabet[random.Intn(len(alphabet))])
	}

	return builder.String()
}
