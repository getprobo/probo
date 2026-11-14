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
	"fmt"

	"go.probo.inc/probo/pkg/automerge"
	automergedm "go.probo.inc/probo/pkg/automerge/prosemirror"
)

// writePreservingSuggestions replaces the body with content and keeps every
// open suggestion whose block was not part of the edit. A block that still
// carries a suggestion and whose accepted text changed is refused.
func writePreservingSuggestions(document *automerge.Document, content string) error {
	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	if !HasOpenSuggestions(spans) {
		return writeSpans(document, content, false)
	}

	inputs, err := automergedm.ToSpans(content)
	if err != nil {
		return fmt.Errorf("cannot convert edited content to collaboration spans: %w", err)
	}

	next, err := mergePreservingSuggestions(spans, inputs)
	if err != nil {
		return err
	}

	if err := WriteBodySpans(document, next); err != nil {
		return fmt.Errorf("cannot write edited collaboration spans: %w", err)
	}

	return nil
}

func mergePreservingSuggestions(before []automerge.Span, after []automerge.SpanInput) ([]automerge.Span, error) {
	oldBlocks := topLevelSpans(before)
	newBlocks := topLevelInputs(after)
	oldText := make([]string, len(oldBlocks))
	newText := make([]string, len(newBlocks))

	for index, block := range oldBlocks {
		oldText[index] = acceptedBlockText(block)
	}

	for index, block := range newBlocks {
		newText[index] = inputBlockText(block)
	}

	alignment := alignBlockText(oldText, newText)
	next := make([]automerge.Span, 0, len(before))

	for _, step := range alignment {
		switch {
		// The incoming content of a suggested block is its reader projection,
		// which drops the redline, so an unchanged suggested block keeps its
		// stored spans. Every other block takes the new content, including
		// type, mark, and embed changes that leave the text alone.
		case step.old >= 0 && step.new >= 0 &&
			oldText[step.old] == newText[step.new] &&
			blockHasSuggestion(oldBlocks[step.old]):
			next = append(next, oldBlocks[step.old]...)
		case step.old >= 0 && blockHasSuggestion(oldBlocks[step.old]):
			return nil, ErrSuggestionOverlaps
		case step.new >= 0:
			for _, input := range newBlocks[step.new] {
				next = append(next, spanFromInput(input))
			}
		}
	}

	return next, nil
}

type blockAlign struct {
	old int
	new int
}

func alignBlockText(before []string, after []string) []blockAlign {
	rows := len(before) + 1
	cols := len(after) + 1
	score := make([][]int, rows)

	for row := range score {
		score[row] = make([]int, cols)
	}

	for row := 1; row < rows; row++ {
		for col := 1; col < cols; col++ {
			if before[row-1] == after[col-1] {
				score[row][col] = score[row-1][col-1] + 1

				continue
			}

			score[row][col] = score[row-1][col]
			if score[row][col-1] > score[row][col] {
				score[row][col] = score[row][col-1]
			}
		}
	}

	reversed := make([]blockAlign, 0, len(before)+len(after))
	row := len(before)
	col := len(after)

	for row > 0 || col > 0 {
		switch {
		case row > 0 && col > 0 && before[row-1] == after[col-1]:
			reversed = append(reversed, blockAlign{old: row - 1, new: col - 1})
			row--
			col--
		case col > 0 && (row == 0 || score[row][col-1] >= score[row-1][col]):
			reversed = append(reversed, blockAlign{old: -1, new: col - 1})
			col--
		default:
			reversed = append(reversed, blockAlign{old: row - 1, new: -1})
			row--
		}
	}

	aligned := make([]blockAlign, len(reversed))
	for index := range reversed {
		aligned[len(reversed)-1-index] = reversed[index]
	}

	return aligned
}

func topLevelSpans(spans []automerge.Span) [][]automerge.Span {
	blocks := make([][]automerge.Span, 0)

	var current []automerge.Span

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock && len(blockParents(span)) == 0 {
			if current != nil {
				blocks = append(blocks, current)
			}

			current = []automerge.Span{span}

			continue
		}

		if current == nil {
			current = []automerge.Span{}
		}

		current = append(current, span)
	}

	if current != nil {
		blocks = append(blocks, current)
	}

	return blocks
}

func topLevelInputs(inputs []automerge.SpanInput) [][]automerge.SpanInput {
	blocks := make([][]automerge.SpanInput, 0)

	var current []automerge.SpanInput

	for _, input := range inputs {
		if input.Block != nil && len(inputParents(input)) == 0 {
			if current != nil {
				blocks = append(blocks, current)
			}

			current = []automerge.SpanInput{input}

			continue
		}

		if current == nil {
			current = []automerge.SpanInput{}
		}

		current = append(current, input)
	}

	if current != nil {
		blocks = append(blocks, current)
	}

	return blocks
}

func inputParents(input automerge.SpanInput) []string {
	switch values := input.Block["parents"].(type) {
	case []string:
		return values
	case []any:
		parents := make([]string, 0, len(values))
		for _, value := range values {
			if name, ok := value.(string); ok {
				parents = append(parents, name)
			}
		}

		return parents
	default:
		return nil
	}
}

func acceptedBlockText(spans []automerge.Span) string {
	text := ""

	for _, span := range spans {
		if span.Type != automerge.SpanTypeText {
			continue
		}

		if markID(span.Marks, SuggestionInsertMark) != "" {
			continue
		}

		text += span.Text
	}

	return text
}

func inputBlockText(inputs []automerge.SpanInput) string {
	text := ""

	for _, input := range inputs {
		text += input.Text
	}

	return text
}

func blockHasSuggestion(spans []automerge.Span) bool {
	return HasOpenSuggestions(spans)
}

func spanFromInput(input automerge.SpanInput) automerge.Span {
	if input.Block != nil {
		return automerge.Span{Type: automerge.SpanTypeBlock, Block: input.Block}
	}

	marks := make(map[string]any, len(input.Marks))
	for name, value := range input.Marks {
		marks[name] = scalarValue(value)
	}

	span := automerge.Span{Type: automerge.SpanTypeText, Text: input.Text}
	if len(marks) > 0 {
		span.Marks = marks
	}

	return span
}

func scalarValue(value automerge.Scalar) any {
	switch value.Type {
	case automerge.ScalarTypeBoolean:
		return value.Bool
	case automerge.ScalarTypeString:
		return value.String
	case automerge.ScalarTypeFloat64:
		return value.Float
	case automerge.ScalarTypeInt, automerge.ScalarTypeCounter, automerge.ScalarTypeTimestamp:
		return value.Int
	case automerge.ScalarTypeUint:
		return value.Uint
	case automerge.ScalarTypeNull:
		return nil
	default:
		return value.String
	}
}
