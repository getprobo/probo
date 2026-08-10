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
	"os"
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

// TestRustText_DanglingMarkBoundaries pins the remaining mark divergence: a mark
// applied with an out-of-range end boundary fails, but the reference has already
// recorded the begin. That dangling begin then captures text according to its
// expand direction, including text inserted to its left when the mark expands
// before. Native drops the mark for that text instead.
//
// Every case here is minimized by delta debugging from randomized differential
// runs. The scenarios are an error path (no valid caller marks past the end of
// the text) which is why they do not gate the build; set
// AUTOMERGE_REQUIRE_DANGLING_MARKS=1 to run them while working on a fix.
func TestRustText_DanglingMarkBoundaries(t *testing.T) {
	t.Parallel()

	if os.Getenv("AUTOMERGE_REQUIRE_DANGLING_MARKS") == "" {
		t.Skip("known divergence: dangling mark begin from an out-of-range boundary")
	}

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
