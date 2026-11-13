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
	"slices"
	"unicode/utf16"

	"go.probo.inc/probo/pkg/automerge"
)

const originalBlockMarker uint32 = 1 << 16

type (
	highlightSegmentKind uint8

	// highlightSegment is a run of the document while a diff replays. Patch
	// indexes are sequential, so each one addresses the document as the
	// previous patches left it.
	highlightSegment struct {
		kind highlightSegmentKind
		// start is the first element of an original run in the before document.
		start int
		// length is the live width of the run. Deleted runs have no width.
		length int
		// text holds the units of an inserted text run. It is nil when the
		// run inserts blocks.
		text []uint16
		// deleted is the original text a deleted run removed.
		deleted string
	}
)

const (
	highlightSegmentOriginal highlightSegmentKind = iota
	highlightSegmentInserted
	highlightSegmentDeleted
)

func historyHighlights(
	document *automerge.Document,
	before []automerge.Hash,
	after []automerge.Hash,
) ([]DocumentHistoryHighlight, error) {
	patches, err := document.Diff(before, after)
	if err != nil {
		return nil, fmt.Errorf("cannot diff revision: %w", err)
	}

	text, err := document.Text("body")
	if err != nil {
		return nil, fmt.Errorf("cannot get collaboration body for highlights: %w", err)
	}

	bodyID, err := text.ObjectID()
	if err != nil {
		return nil, err
	}

	var original []uint32

	if len(before) > 0 {
		beforeSpans, err := text.SpansAt(before)
		if err != nil {
			return nil, fmt.Errorf("cannot read before spans for highlights: %w", err)
		}

		original = originalElements(beforeSpans)
	}

	segments := []highlightSegment{}
	if len(original) > 0 {
		segments = append(
			segments,
			highlightSegment{kind: highlightSegmentOriginal, length: len(original)},
		)
	}

	for _, patch := range patches {
		if patch.Object != bodyID {
			continue
		}

		switch patch.Action {
		case automerge.PatchSpliceText:
			if patch.Text == "" {
				continue
			}

			units := utf16.Encode([]rune(patch.Text))

			segments = insertSegment(
				segments,
				int(patch.Index),
				highlightSegment{
					kind:   highlightSegmentInserted,
					length: len(units),
					text:   units,
				},
			)
		case automerge.PatchInsert:
			index := int(patch.Index)

			for _, run := range insertedRuns(patch.Values) {
				segments = insertSegment(segments, index, run)
				index += run.length
			}
		case automerge.PatchDeleteSeq:
			segments = deleteSegments(
				segments,
				original,
				int(patch.Index),
				int(patch.Length),
			)
		}
	}

	return highlightsFromSegments(segments), nil
}

// insertedRuns groups the values of an insert patch into runs of text and runs
// of other elements, which are block markers.
func insertedRuns(values []automerge.InsertedValue) []highlightSegment {
	runs := make([]highlightSegment, 0, 1)

	for _, inserted := range values {
		scalar := inserted.Value.Scalar
		if scalar != nil && scalar.Type == automerge.ScalarTypeString {
			units := utf16.Encode([]rune(scalar.String))

			if last := len(runs) - 1; last >= 0 && runs[last].text != nil {
				runs[last].text = append(runs[last].text, units...)
				runs[last].length += len(units)

				continue
			}

			runs = append(
				runs,
				highlightSegment{
					kind:   highlightSegmentInserted,
					length: len(units),
					text:   units,
				},
			)

			continue
		}

		if last := len(runs) - 1; last >= 0 && runs[last].text == nil {
			runs[last].length++

			continue
		}

		runs = append(runs, highlightSegment{kind: highlightSegmentInserted, length: 1})
	}

	return runs
}

func originalElements(spans []automerge.Span) []uint32 {
	elements := make([]uint32, 0, len(spans))

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock || span.Block != nil {
			elements = append(elements, originalBlockMarker)

			continue
		}

		for _, unit := range utf16.Encode([]rune(span.Text)) {
			elements = append(elements, uint32(unit))
		}
	}

	return elements
}

func highlightsFromSegments(segments []highlightSegment) []DocumentHistoryHighlight {
	highlights := make([]DocumentHistoryHighlight, 0)

	var (
		offset int
		last   = -1
	)

	for _, segment := range segments {
		switch segment.kind {
		case highlightSegmentOriginal:
			offset += segment.length
			last = -1
		case highlightSegmentInserted:
			if segment.text == nil {
				offset += segment.length
				last = -1

				continue
			}

			inserted := string(utf16.Decode(segment.text))

			if last >= 0 && highlights[last].Kind == DocumentHistoryHighlightKindInsert && highlights[last].To == offset {
				highlights[last].To += segment.length
				highlights[last].Text += inserted
			} else {
				highlights = append(
					highlights,
					DocumentHistoryHighlight{
						Kind: DocumentHistoryHighlightKindInsert,
						From: offset,
						To:   offset + segment.length,
						Text: inserted,
					},
				)
				last = len(highlights) - 1
			}

			offset += segment.length
		case highlightSegmentDeleted:
			if last >= 0 && highlights[last].Kind == DocumentHistoryHighlightKindDelete {
				highlights[last].Text += segment.deleted

				continue
			}

			highlights = append(
				highlights,
				DocumentHistoryHighlight{
					Kind: DocumentHistoryHighlightKindDelete,
					From: offset,
					To:   offset,
					Text: segment.deleted,
				},
			)
			last = len(highlights) - 1
		}
	}

	return highlights
}

// insertSegment places inserted at the live index, after any deleted run that
// already sits there so a replacement reads as deletion then insertion.
func insertSegment(
	segments []highlightSegment,
	index int,
	inserted highlightSegment,
) []highlightSegment {
	segments = splitSegmentLive(segments, index)
	position := livePosition(segments, index)

	return slices.Insert(segments, position, inserted)
}

// livePosition returns the position in segments of the run that starts at the
// live index, or the end when the index is past the document. Deleted runs at
// the boundary stay before it.
func livePosition(segments []highlightSegment, index int) int {
	offset := 0

	for position, segment := range segments {
		if segment.kind == highlightSegmentDeleted {
			continue
		}

		if index >= offset+segment.length {
			offset += segment.length

			continue
		}

		return position
	}

	return len(segments)
}

func deleteSegments(
	segments []highlightSegment,
	original []uint32,
	index int,
	length int,
) []highlightSegment {
	if length <= 0 {
		return segments
	}

	segments = splitSegmentLive(segments, index)
	segments = splitSegmentLive(segments, index+length)

	offset := 0
	next := make([]highlightSegment, 0, len(segments))

	for _, segment := range segments {
		if segment.kind == highlightSegmentDeleted {
			next = append(next, segment)

			continue
		}

		inside := offset >= index && offset+segment.length <= index+length

		offset += segment.length

		if !inside {
			next = append(next, segment)

			continue
		}

		if segment.kind != highlightSegmentOriginal {
			continue
		}

		if deleted := originalText(original, segment.start, segment.length); deleted != "" {
			next = append(
				next,
				highlightSegment{kind: highlightSegmentDeleted, deleted: deleted},
			)
		}
	}

	return next
}

// splitSegmentLive cuts the run that strictly contains the live index so a
// range edit can address whole runs.
func splitSegmentLive(segments []highlightSegment, index int) []highlightSegment {
	offset := 0

	for position, segment := range segments {
		if segment.kind == highlightSegmentDeleted {
			continue
		}

		if index >= offset+segment.length {
			offset += segment.length

			continue
		}

		if index == offset {
			return segments
		}

		cut := index - offset
		head, tail := cutSegment(segment, cut)

		result := make([]highlightSegment, 0, len(segments)+1)
		result = append(result, segments[:position]...)
		result = append(result, head, tail)
		result = append(result, segments[position+1:]...)

		return result
	}

	return segments
}

func cutSegment(segment highlightSegment, at int) (highlightSegment, highlightSegment) {
	head := highlightSegment{kind: segment.kind, start: segment.start, length: at}
	tail := highlightSegment{
		kind:   segment.kind,
		start:  segment.start + at,
		length: segment.length - at,
	}

	if segment.text != nil {
		head.text = segment.text[:at:at]
		tail.text = segment.text[at:]
	}

	return head, tail
}

func originalText(original []uint32, start int, length int) string {
	units := make([]uint16, 0, length)

	for _, element := range original[start : start+length] {
		if element == originalBlockMarker {
			continue
		}

		units = append(units, uint16(element))
	}

	return string(utf16.Decode(units))
}
