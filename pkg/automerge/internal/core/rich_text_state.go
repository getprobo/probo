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

package core

import (
	"fmt"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"reflect"
	"sort"
)

func (s *State) RichTextSpans(object opset.OpID) ([]RichSpan, error) {
	elements := s.sequenceElements(object)
	marks := s.richTextMarks(object, elements)
	spans := make([]RichSpan, 0)

	for i, operation := range elements {
		switch operation.Action {
		case opset.ActionMakeMap:
			value, err := s.mapValue(operation.ID, make(map[opset.OpID]struct{}))
			if err != nil {
				return nil, fmt.Errorf("cannot hydrate block %v: %w", operation.ID, err)
			}

			spans = append(spans, RichSpan{Type: "block", Value: value})
		case opset.ActionSet:
			if operation.Value == nil || operation.Value.Type != opset.ScalarString {
				continue
			}

			activeMarks := make(map[string]any)

			for _, mark := range marks {
				if i >= mark.start && i < mark.end {
					if mark.value == nil {
						delete(activeMarks, mark.name)
					} else {
						activeMarks[mark.name] = mark.value
					}
				}
			}

			if len(activeMarks) == 0 {
				activeMarks = nil
			}

			if len(spans) > 0 &&
				spans[len(spans)-1].Type == "text" &&
				reflect.DeepEqual(spans[len(spans)-1].Marks, activeMarks) {
				spans[len(spans)-1].Value = spans[len(spans)-1].Value.(string) +
					operation.Value.String
			} else {
				spans = append(
					spans,
					RichSpan{
						Type:  "text",
						Value: operation.Value.String,
						Marks: activeMarks,
					},
				)
			}
		}
	}

	return spans, nil
}

type richTextMark struct {
	start  int
	end    int
	name   string
	value  any
	scalar *opset.Scalar
	// id is the mark's begin operation, which orders precedence: a later mark
	// (an unmark, or a new value) overrides an earlier one where they overlap.
	id opset.OpID
}

// MarkRange is one active mark over a UTF-16 range of a text object.
type MarkRange struct {
	Start uint32
	End   uint32
	Name  string
	Value *opset.Scalar
}

// insertAnchorKey adjusts an insertion anchor so a new element lands on the
// correct side of the mark boundaries that follow it, mirroring the reference's
// insert query. Scanning forward from the anchor, an expanding mark begin and a
// non-expanding mark end each offer a position after themselves, so the new
// element joins the expanding range. Reaching the end of a mark whose begin
// offered a position withdraws that offer, because a begin/end pair with no
// visible content between them must not capture the insertion. The scan stops at
// the first visible element; tombstones are stepped over.
func (s *State) insertAnchorKey(object opset.OpID, base opset.Key) opset.Key {
	index := s.sequenceIndex(object)
	if index.markCount == 0 {
		return base
	}

	start := 0

	if !base.IsHead {
		if base.Element == nil {
			return base
		}

		position, ok := index.rawPosition(*base.Element)
		if !ok {
			return base
		}

		start = position + 1
	}

	type candidate struct {
		key opset.Key
		id  *opset.OpID
	}

	candidates := []candidate{{key: base}}

	index.eachRaw(start, func(entry sequenceIndexEntry) bool {
		operation, ok := s.operation(entry.insertion)
		if !ok {
			return true
		}

		if operation.Action == opset.ActionMark {
			expand := operation.MarkExpand != nil && *operation.MarkExpand
			isEnd := operation.MarkName == nil
			withdrawn := false

			if isEnd {
				begin := opset.OpID{Actor: operation.ID.Actor, Counter: operation.ID.Counter - 1}

				for index := range candidates {
					if candidates[index].id != nil && *candidates[index].id == begin {
						candidates = candidates[:index]
						withdrawn = true

						break
					}
				}
			}

			if !withdrawn && ((!isEnd && expand) || (isEnd && !expand)) {
				candidates = append(
					candidates,
					candidate{
						key: opset.Key{Element: new(operation.ID)},
						id:  new(operation.ID),
					},
				)
			}

			return true
		}

		if !s.isSuperseded(operation.ID) && len(candidates) > 0 {
			return false
		}

		return true
	})

	if len(candidates) == 0 {
		return base
	}

	return candidates[len(candidates)-1].key
}

// richTextMarks computes the active mark ranges of a text object by walking the
// sequence order and running a mark state machine, mirroring the reference. A
// mark begin opens a range at the current visible index and its matching end
// closes it. Because mark operations hold positions in the sequence, text
// inserted at an expanding boundary sits inside the range and keeps the mark
// even after the originally marked content is deleted.
func (s *State) richTextMarks(object opset.OpID, elements []opset.Operation) []richTextMark {
	order := s.insertOrder(object)

	type openMark struct {
		start     int
		operation opset.Operation
	}

	open := make(map[opset.OpID]openMark)
	marks := make([]richTextMark, 0)
	elementIndex := make(map[opset.OpID]int)
	index := 0

	closeMark := func(begin openMark, end int) {
		if end <= begin.start || begin.operation.MarkName == nil {
			return
		}

		marks = append(
			marks,
			richTextMark{
				start:  begin.start,
				end:    end,
				name:   *begin.operation.MarkName,
				value:  scalarMaterializedValue(begin.operation.Value),
				scalar: begin.operation.Value,
				id:     begin.operation.ID,
			},
		)
	}

	for _, id := range order {
		operation, ok := s.operation(id)
		if !ok || s.isSuperseded(id) {
			continue
		}

		if operation.Action != opset.ActionMark {
			elementIndex[id] = index
			index++

			continue
		}

		if operation.MarkName != nil {
			open[operation.ID] = openMark{start: index, operation: operation}

			continue
		}

		begin := opset.OpID{Actor: operation.ID.Actor, Counter: operation.ID.Counter - 1}
		if opened, ok := open[begin]; ok {
			delete(open, begin)
			closeMark(opened, index)
		}
	}

	// A begin whose matching end operation was never created extends to the end
	// of the text. This happens when a mark was applied with an out-of-range end
	// boundary: the reference records the begin and then fails on the end, so the
	// begin dangles. A begin whose end operation exists but was simply visited
	// first (a zero-length mark, where begin and end share an anchor and sibling
	// insertions are ordered by descending operation ID) covers nothing.
	remaining := make([]openMark, 0, len(open))

	for _, opened := range open {
		endID := opset.OpID{Actor: opened.operation.ID.Actor, Counter: opened.operation.ID.Counter + 1}

		// The end operation exists only when the following operation is actually
		// a mark end. A begin whose end insert failed leaves that counter free
		// for a later operation (a delete, say), so checking the action avoids
		// mistaking such an operation for the missing end.
		if end, ok := s.operation(endID); ok &&
			end.Action == opset.ActionMark && end.MarkName == nil {
			continue
		}

		remaining = append(remaining, opened)
	}

	sort.Slice(
		remaining,
		func(i, j int) bool {
			return remaining[i].operation.ID.Compare(remaining[j].operation.ID) < 0
		},
	)

	for _, opened := range remaining {
		// A dangling begin that expands leftward (expand "before" or "both")
		// covers text back to its own anchor rather than only from where it sorts
		// in the RGA order. The begin sorts after same-anchor insertions by
		// descending operation ID, so its walk index lands past text it should
		// cover; the reference instead starts the mark at the position just after
		// the begin's anchor element (or at the document start for a head anchor).
		if opened.operation.MarkExpand != nil && *opened.operation.MarkExpand {
			opened.start = danglingBeginStart(opened.operation.Key, elementIndex, opened.start)
		}

		closeMark(opened, index)
	}

	// Precedence follows creation order, so a later unmark or replacement value
	// wins over an earlier mark where the two overlap.
	sort.SliceStable(
		marks,
		func(i, j int) bool {
			return marks[i].id.Compare(marks[j].id) < 0
		},
	)

	_ = elements

	return marks
}

// danglingBeginStart returns the visible index a leftward-expanding dangling
// begin should start from: the document start for a head anchor, the position
// immediately after the anchor element otherwise, and the walk index as a
// fallback when the anchor is no longer visible.
func danglingBeginStart(anchor opset.Key, elementIndex map[opset.OpID]int, fallback int) int {
	if anchor.IsHead {
		return 0
	}

	if anchor.Element != nil {
		if position, ok := elementIndex[*anchor.Element]; ok {
			return position + 1
		}
	}

	return fallback
}

// Marks returns the active marks over a text object as UTF-16 ranges, matching
// upstream Rust's marks(): contiguous runs of an identical (name, value) mark
// are merged, block markers occupy one position, and marks removed by a null
// value are excluded.
func (s *State) Marks(object opset.OpID) []MarkRange {
	elements := s.sequenceElements(object)
	marks := s.richTextMarks(object, elements)

	type openMark struct {
		start uint32
		value *opset.Scalar
	}

	open := make(map[string]openMark)

	result := make([]MarkRange, 0)

	var position uint32

	closeMark := func(name string, mark openMark, end uint32) {
		result = append(
			result,
			MarkRange{
				Start: mark.start,
				End:   end,
				Name:  name,
				Value: mark.value,
			},
		)
	}

	for index, element := range elements {
		active := make(map[string]*opset.Scalar)

		for _, mark := range marks {
			if index < mark.start || index >= mark.end {
				continue
			}

			if mark.scalar == nil || mark.scalar.Type == opset.ScalarNull {
				delete(active, mark.name)
			} else {
				active[mark.name] = mark.scalar
			}
		}

		for name, mark := range open {
			value, ok := active[name]
			if !ok || !scalarValuesEqual(*value, *mark.value) {
				closeMark(name, mark, position)
				delete(open, name)
			}
		}

		for name, value := range active {
			if _, ok := open[name]; !ok {
				open[name] = openMark{start: position, value: value}
			}
		}

		position += elementLength(element)
	}

	for name, mark := range open {
		closeMark(name, mark, position)
	}

	sort.Slice(
		result,
		func(i, j int) bool {
			if result[i].Start != result[j].Start {
				return result[i].Start < result[j].Start
			}

			return result[i].Name < result[j].Name
		},
	)

	return result
}

func (s *State) markRangeHasSurvivingElement(
	object opset.OpID,
	begin opset.Operation,
	end opset.Operation,
) bool {
	elements := s.sequenceAll(object)
	start := 0

	if begin.Key.Element != nil {
		for i, element := range elements {
			if element.ID == *begin.Key.Element {
				start = i + 1

				break
			}
		}
	}

	stop := len(elements)
	if end.Key.Element != nil {
		for i, element := range elements {
			if element.ID == *end.Key.Element {
				stop = i + 1

				break
			}
		}
	}

	if start > stop {
		return false
	}

	for _, element := range elements[start:stop] {
		if element.ID.Compare(begin.ID) < 0 &&
			!s.isSuperseded(element.ID) {
			return true
		}
	}

	return false
}

// markOpUTF16Range returns the literal UTF-16 range a mark operation pair spans,
// without boundary-expansion adjustment. It is used to report mark and unmark
// operations as Mark patches, matching the reference's operation-based diff.
func (s *State) markOpUTF16Range(object opset.OpID, begin, end opset.Operation) (uint32, uint32, bool) {
	elements := s.sequenceElements(object)

	positions := make(map[opset.OpID]int, len(elements))
	for index, element := range elements {
		positions[element.ID] = index
	}

	beginExpand := begin.MarkExpand != nil && *begin.MarkExpand
	endExpand := end.MarkExpand != nil && *end.MarkExpand

	startIndex, startOK := s.markAnchorPosition(
		object,
		begin.Key,
		begin.ID,
		true,
		beginExpand,
		positions,
		elements,
		false,
		make(map[opset.OpID]struct{}),
	)
	endIndex, endOK := s.markAnchorPosition(
		object,
		end.Key,
		end.ID,
		false,
		endExpand,
		positions,
		elements,
		false,
		make(map[opset.OpID]struct{}),
	)

	if !startOK || !endOK {
		return 0, 0, false
	}

	return utf16PrefixLength(elements, startIndex), utf16PrefixLength(elements, endIndex), true
}

// utf16PrefixLength sums the UTF-16 width of the first count elements, so a mark
// anchor expressed as an element index becomes a UTF-16 position.
func utf16PrefixLength(elements []opset.Operation, count int) uint32 {
	var position uint32

	for i := 0; i < count && i < len(elements); i++ {
		position += elementLength(elements[i])
	}

	return position
}

func (s *State) markAnchorPosition(
	object opset.OpID,
	key opset.Key,
	marker opset.OpID,
	start bool,
	expand bool,
	positions map[opset.OpID]int,
	elements []opset.Operation,
	adjustBoundary bool,
	visited map[opset.OpID]struct{},
) (int, bool) {
	if key.IsHead {
		position := 0
		if adjustBoundary && ((start && !expand) || (!start && expand)) {
			position = s.markBoundaryInsertionEnd(key, marker, position, elements)
		}

		return position, true
	}

	if key.Element == nil {
		return 0, false
	}

	if position, ok := positions[*key.Element]; ok {
		position++
		if adjustBoundary && ((start && !expand) || (!start && expand)) {
			position = s.markBoundaryInsertionEnd(key, marker, position, elements)
		}

		return position, true
	}

	if _, ok := visited[*key.Element]; ok {
		return 0, false
	}

	visited[*key.Element] = struct{}{}

	operation, ok := s.operation(*key.Element)
	if !ok {
		return 0, false
	}

	if operation.Action == opset.ActionMark {
		return s.markAnchorPosition(
			object,
			operation.Key,
			marker,
			start,
			expand,
			positions,
			elements,
			false,
			visited,
		)
	}

	if expand {
		position := 0

		for _, element := range s.sequenceAll(object) {
			if element.ID == operation.ID {
				return position, true
			}

			if !s.isSuperseded(element.ID) {
				position++
			}
		}

		return 0, false
	}

	// A non-expanding marker anchored to a deleted element stays before
	// insertions at that element's former position. Follow the deleted
	// element's own predecessor chain to find that position.
	return s.markAnchorPosition(
		object,
		operation.Key,
		marker,
		start,
		expand,
		positions,
		elements,
		false,
		visited,
	)
}

// markBoundaryInsertionEnd returns the position after insertion branches that
// were created at a mark boundary after the marker operation. Mark markers and
// ordinary sequence insertions share an anchor in the Automerge operation tree;
// their relative operation IDs determine which side of the marker a later
// insertion occupies. Expanding end markers and non-expanding begin markers sit
// after these branches, while the opposite expansion modes sit before them.
func (s *State) markBoundaryInsertionEnd(
	key opset.Key,
	marker opset.OpID,
	position int,
	elements []opset.Operation,
) int {
	for position < len(elements) {
		child, ok := s.boundaryChild(elements[position], key, make(map[opset.OpID]struct{}))
		if !ok || child.Compare(marker) <= 0 {
			break
		}

		position++
	}

	return position
}

// boundaryChild returns the direct insertion child of a boundary anchor for an
// element, following insertion ancestry through the sequence tree.
func (s *State) boundaryChild(
	element opset.Operation,
	boundary opset.Key,
	visited map[opset.OpID]struct{},
) (opset.OpID, bool) {
	current := element

	for {
		if _, ok := visited[current.ID]; ok {
			return opset.OpID{}, false
		}

		visited[current.ID] = struct{}{}

		if boundary.IsHead && current.Key.IsHead {
			return current.ID, true
		}

		if boundary.Element != nil &&
			current.Key.Element != nil &&
			*current.Key.Element == *boundary.Element {
			return current.ID, true
		}

		if current.Key.Element == nil {
			return opset.OpID{}, false
		}

		parent, ok := s.operation(*current.Key.Element)
		if !ok || parent.Action == opset.ActionMark {
			return opset.OpID{}, false
		}

		current = parent
	}
}
