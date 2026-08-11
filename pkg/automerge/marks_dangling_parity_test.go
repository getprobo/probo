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
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// markScenarioStep is one editing step in a reproducible mark scenario.
type markScenarioStep struct {
	kind   string
	index  uint32
	end    uint32
	count  int32
	value  string
	name   string
	expand automerge.MarkExpand
}

func (s markScenarioStep) String() string {
	switch s.kind {
	case "insert":
		return fmt.Sprintf("insert(%d,%q)", s.index, s.value)
	case "delete":
		return fmt.Sprintf("delete(%d,%d)", s.index, s.count)
	case "split":
		return fmt.Sprintf("split(%d)", s.index)
	default:
		return fmt.Sprintf("mark(%d,%d,%s,%s)", s.index, s.end, s.name, s.expand)
	}
}

// TestRustText_DanglingMarkBoundaries gates the dangling-begin behavior: a mark
// applied with an out-of-range end boundary fails, but the reference has already
// recorded the begin. That dangling begin then captures text according to its
// expand direction, including text inserted to its left when the mark expands
// before. Native now starts a leftward-expanding dangling begin at the position
// just after its own anchor, matching the reference.
//
// The block-boundary and multi-dangling cases below were minimized by delta
// debugging from randomized differential runs. They all shared one cause: a
// split block did not resolve its insertion anchor against neighbouring mark
// boundaries the way a text insertion does, so a block, and every insertion
// anchored after it, landed on the wrong side of a dangling begin. SplitBlock
// now resolves its anchor identically to Splice, and a value-level randomized
// sweep including out-of-range marks and every expand mode no longer diverges.
func TestRustText_DanglingMarkBoundaries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name  string
		steps []markScenarioStep
	}{
		{
			name: "expand before captures a later left insertion",
			steps: []markScenarioStep{
				{kind: "mark", index: 0, end: 3, name: "bold", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "w"},
			},
		},
		{
			name: "expand both captures a later left insertion",
			steps: []markScenarioStep{
				{kind: "mark", index: 0, end: 5, name: "underline", expand: automerge.MarkExpandBoth},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "i"},
			},
		},
		{
			name: "dangling begin survives a full delete",
			steps: []markScenarioStep{
				{kind: "mark", index: 0, end: 1, name: "italic", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 0},
				{kind: "mark", index: 0, end: 1, name: "italic", expand: automerge.MarkExpandBoth},
				{kind: "delete", index: 0, count: 5},
				{kind: "insert", index: 0, value: "G"},
			},
		},
		{
			name: "dangling begin spans a block boundary",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 1, end: 3, name: "underline", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 1},
				{kind: "insert", index: 0, value: "Roa"},
				{kind: "insert", index: 4, value: "J"},
			},
		},
		{
			name: "two dangling begins both survive",
			steps: []markScenarioStep{
				{kind: "mark", index: 0, end: 3, name: "italic", expand: automerge.MarkExpandBoth},
				{kind: "mark", index: 0, end: 3, name: "bold", expand: automerge.MarkExpandBoth},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "u"},
			},
		},
		{
			name: "overlapping begins with mixed expand both survive",
			steps: []markScenarioStep{
				{kind: "mark", index: 0, end: 1, name: "italic", expand: automerge.MarkExpandBoth},
				{kind: "mark", index: 0, end: 1, name: "underline", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "vJ"},
			},
		},
		{
			name: "dangling begin after a full delete and resplit",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 1, end: 4, name: "bold", expand: automerge.MarkExpandBoth},
				{kind: "delete", index: 0, count: 4},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "rG"},
			},
		},
		{
			name: "none expand mark does not leak past a block",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 0, end: 1, name: "bold", expand: automerge.MarkExpandNone},
				{kind: "split", index: 1},
				{kind: "insert", index: 1, value: "Plk"},
			},
		},
		{
			name: "before expand mark does not leak across a block",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 0, end: 1, name: "bold", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 1},
				{kind: "insert", index: 1, value: "lMf"},
			},
		},
		{
			name: "before expand mark bounded by two blocks",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 0, end: 1, name: "bold", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 1},
				{kind: "split", index: 1},
				{kind: "insert", index: 2, value: "ed"},
			},
		},
		{
			name: "before expand mark across a block at head",
			steps: []markScenarioStep{
				{kind: "split", index: 0},
				{kind: "mark", index: 0, end: 1, name: "italic", expand: automerge.MarkExpandBefore},
				{kind: "split", index: 0},
				{kind: "insert", index: 0, value: "C"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equalf(t,
				runMarkScenario(t, ctx, rustParityEngines()[1], tt.steps),
				runMarkScenario(t, ctx, rustParityEngines()[0], tt.steps),
				"steps: %s", renderMarkScenario(tt.steps),
			)
		})
	}
}

// TestRustText_MarkValuesMatchReferenceUnderErrors is the standing gate for the
// dangling-mark behavior: many randomized scenarios, deliberately including marks
// whose end boundary runs past the text and every expand mode, must produce the
// exact same marked spans on the native and reference engines. This is stronger
// than the consolidation-and-text invariants of marks_are_okay because it
// compares the mark values run for run, which is what caught the block-anchor
// divergence this gate now protects against.
func TestRustText_MarkValuesMatchReferenceUnderErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	random := rand.New(rand.NewSource(0x1e3779b97f4a7c15))

	const scenarios = 2000

	for scenario := range scenarios {
		steps := randomDanglingMarkSteps(random)

		reference := runMarkScenario(t, ctx, rustParityEngines()[1], steps)
		native := runMarkScenario(t, ctx, rustParityEngines()[0], steps)

		require.Equalf(t, reference, native,
			"scenario %d diverged; steps: %s", scenario, renderMarkScenario(steps))
	}
}

// randomDanglingMarkSteps builds a random editing scenario over text, list and
// mark operations. Marks may target an end boundary past the current length so
// the error path that leaves a dangling begin is exercised, and every expand
// mode appears.
func randomDanglingMarkSteps(random *rand.Rand) []markScenarioStep {
	names := []string{"bold", "italic", "underline"}
	expands := []automerge.MarkExpand{
		automerge.MarkExpandNone,
		automerge.MarkExpandBefore,
		automerge.MarkExpandAfter,
		automerge.MarkExpandBoth,
	}

	steps := 3 + random.Intn(10)
	out := make([]markScenarioStep, 0, steps)
	length := 0

	for range steps {
		switch random.Intn(4) {
		case 0:
			index := random.Intn(length + 1)
			value := randomLetters(random, 1+random.Intn(3))
			out = append(out, markScenarioStep{kind: "insert", index: uint32(index), value: value})
			length += len(value)
		case 1:
			if length == 0 {
				continue
			}

			index := random.Intn(length)
			count := 1 + random.Intn(length-index)
			out = append(out, markScenarioStep{kind: "delete", index: uint32(index), count: int32(count)})
			length -= count
		case 2:
			index := random.Intn(length + 1)
			out = append(out, markScenarioStep{kind: "split", index: uint32(index)})
			length++
		case 3:
			if length == 0 {
				continue
			}

			index := random.Intn(length)
			// end may exceed the length so the dangling-begin path is covered.
			end := index + 1 + random.Intn(length+2-index)
			out = append(out, markScenarioStep{
				kind:   "mark",
				index:  uint32(index),
				end:    uint32(end),
				name:   names[random.Intn(len(names))],
				expand: expands[random.Intn(len(expands))],
			})
		}
	}

	return out
}

func runMarkScenario(
	t *testing.T,
	ctx context.Context,
	engine rustParityEngine,
	steps []markScenarioStep,
) string {
	t.Helper()

	document, err := engine.open(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText(ctx, "text")
	require.NoError(t, err)

	// Out-of-range marks are expected to fail on both engines; the divergence is
	// in the spans they leave behind, so step errors are deliberately ignored.
	for _, step := range steps {
		switch step.kind {
		case "insert":
			_ = text.Splice(ctx, step.index, 0, step.value)
		case "delete":
			_ = text.Splice(ctx, step.index, step.count, "")
		case "split":
			_, _ = text.SplitBlock(ctx, step.index)
		case "mark":
			_ = text.Mark(ctx, step.index, step.end, step.name,
				automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true}, step.expand)
		}
	}

	_, _ = document.Commit(ctx, "scenario", commitTime)

	spans, err := text.Spans(ctx)
	require.NoError(t, err)

	return renderMarkedSpans(spans)
}

func renderMarkedSpans(spans []automerge.Span) string {
	var builder strings.Builder

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			builder.WriteString("|block")

			continue
		}

		names := make([]string, 0, len(span.Marks))
		for name := range span.Marks {
			names = append(names, name)
		}

		sort.Strings(names)
		fmt.Fprintf(&builder, "|%q{%s}", span.Text, strings.Join(names, ","))
	}

	return builder.String()
}

func renderMarkScenario(steps []markScenarioStep) string {
	rendered := make([]string, 0, len(steps))
	for _, step := range steps {
		rendered = append(rendered, step.String())
	}

	return strings.Join(rendered, " ")
}
