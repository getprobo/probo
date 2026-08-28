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
	"sort"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

const (
	sequenceIndexChunkTarget  = 128
	sequenceIndexChunkMaximum = sequenceIndexChunkTarget * 2
)

type (
	sequenceIndex struct {
		state        *State
		chunks       []*sequenceIndexChunk
		positions    map[opset.OpID]sequenceIndexPosition
		visibleCount int
		utf16Width   uint32
		markCount    int
	}

	sequenceIndexChunk struct {
		entries      []sequenceIndexEntry
		prefixWidths []uint32
		prefixValues []uint16
		visibleCount int
		utf16Width   uint32
		markCount    int
	}

	sequenceIndexEntry struct {
		insertion opset.OpID
		winner    opset.OpID
		width     uint32
		visible   bool
		mark      bool
	}

	sequenceIndexPosition struct {
		chunk  *sequenceIndexChunk
		offset int
	}

	sequenceIndexRange struct {
		start    int
		end      int
		previous *opset.OpID
		targets  []opset.Operation
	}
)

// buildSequenceIndex constructs an object's RGA order without materializing a
// document-wide winner map. Replacement candidates are already grouped by
// element in sequenceElementIndex, so winner selection is local to each entry.
func (s *State) buildSequenceIndex(object opset.OpID) *sequenceIndex {
	children := make(map[opset.OpID][]opset.OpID)
	var head []opset.OpID

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert {
			return true
		}

		if operation.Key.IsHead {
			head = append(head, operation.ID)
		} else if operation.Key.Element != nil {
			children[*operation.Key.Element] = append(
				children[*operation.Key.Element],
				operation.ID,
			)
		}
		return true
	})

	ordered := make([]opset.OpID, 0)
	visited := make(map[opset.OpID]struct{})
	var appendBranch func([]opset.OpID)
	appendBranch = func(identifiers []opset.OpID) {
		sort.Slice(identifiers, func(i, j int) bool {
			return identifiers[i].Compare(identifiers[j]) > 0
		})
		for _, identifier := range identifiers {
			if _, ok := visited[identifier]; ok {
				continue
			}
			visited[identifier] = struct{}{}
			ordered = append(ordered, identifier)
			appendBranch(children[identifier])
		}
	}
	appendBranch(head)

	index := &sequenceIndex{
		state:     s,
		positions: make(map[opset.OpID]sequenceIndexPosition, len(ordered)),
	}
	for start := 0; start < len(ordered); start += sequenceIndexChunkTarget {
		end := min(start+sequenceIndexChunkTarget, len(ordered))
		entries := make([]sequenceIndexEntry, 0, end-start)

		for _, identifier := range ordered[start:end] {
			operation, ok := s.operation(identifier)
			if !ok {
				continue
			}
			entries = append(
				entries,
				s.newSequenceIndexEntry(operation),
			)
		}

		index.appendChunk(newSequenceIndexChunk(entries))
	}

	return index
}

func newSequenceIndexChunk(entries []sequenceIndexEntry) *sequenceIndexChunk {
	chunk := &sequenceIndexChunk{
		entries:      entries,
		prefixWidths: make([]uint32, len(entries)+1),
		prefixValues: make([]uint16, len(entries)+1),
	}
	for index, entry := range entries {
		if entry.mark {
			chunk.markCount++
		}
		if entry.visible {
			chunk.visibleCount++
			chunk.utf16Width += entry.width
		}

		chunk.prefixWidths[index+1] = chunk.utf16Width
		chunk.prefixValues[index+1] = uint16(chunk.visibleCount)
	}

	return chunk
}

func (i *sequenceIndex) appendChunk(chunk *sequenceIndexChunk) {
	i.chunks = append(i.chunks, chunk)
	i.visibleCount += chunk.visibleCount
	i.utf16Width += chunk.utf16Width
	i.markCount += chunk.markCount
	i.indexChunk(chunk)
}

func (i *sequenceIndex) indexChunk(chunk *sequenceIndexChunk) {
	i.indexChunkFrom(chunk, 0)
}

func (i *sequenceIndex) indexChunkFrom(chunk *sequenceIndexChunk, start int) {
	for offset, entry := range chunk.entries[start:] {
		i.positions[entry.insertion] = sequenceIndexPosition{
			chunk:  chunk,
			offset: start + offset,
		}
	}
}

func (s *State) newSequenceIndexEntry(insertion opset.Operation) sequenceIndexEntry {
	entry := sequenceIndexEntry{
		insertion: insertion.ID,
		mark:      insertion.Action == opset.ActionMark,
	}
	if entry.mark {
		return entry
	}

	winner, ok := s.sequenceElementWinner(insertion)
	if !ok {
		return entry
	}

	entry.winner = winner.ID
	entry.visible = true
	entry.width = uint32(sequenceValueUTF16Width(sequenceValue{
		Element:   insertion.ID,
		Operation: winner,
	}))

	return entry
}

func (s *State) sequenceElementWinner(
	insertion opset.Operation,
) (opset.Operation, bool) {
	var (
		winner opset.Operation
		found  bool
	)

	if !s.isSuperseded(insertion.ID) {
		winner = insertion
		found = true
	}

	for _, identifier := range s.sequenceElementIndex[insertion.ID] {
		operation, ok := s.operation(identifier)
		if !ok ||
			operation.Insert ||
			operation.Action == opset.ActionDelete ||
			operation.Action == opset.ActionIncrement ||
			s.isSuperseded(operation.ID) {
			continue
		}

		if !found || operation.ID.Compare(winner.ID) > 0 {
			winner = operation
			found = true
		}
	}

	return winner, found
}

func (i *sequenceIndex) order() []opset.OpID {
	order := make([]opset.OpID, 0, len(i.positions))
	for _, chunk := range i.chunks {
		for _, entry := range chunk.entries {
			order = append(order, entry.insertion)
		}
	}

	return order
}

func (i *sequenceIndex) elements() []opset.Operation {
	elements := make([]opset.Operation, 0, i.visibleCount)
	for _, chunk := range i.chunks {
		for _, entry := range chunk.entries {
			if entry.visible {
				operation, ok := i.state.operation(entry.insertion)
				if ok {
					elements = append(elements, operation)
				}
			}
		}
	}

	return elements
}

func (i *sequenceIndex) values() []sequenceValue {
	values := make([]sequenceValue, 0, i.visibleCount)
	for _, chunk := range i.chunks {
		for _, entry := range chunk.entries {
			if entry.visible {
				operation, ok := i.state.operation(entry.winner)
				if !ok {
					continue
				}
				values = append(values, sequenceValue{
					Element:   entry.insertion,
					Operation: operation,
				})
			}
		}
	}

	return values
}

// rangeAt resolves a UTF-16 edit directly over chunk summaries. Only the
// boundary chunks are inspected element-by-element.
func (i *sequenceIndex) rangeAt(index, deleteCount uint32) (sequenceIndexRange, error) {
	if index > i.utf16Width {
		return sequenceIndexRange{}, fmt.Errorf("text index %d is out of bounds", index)
	}

	start, startWidth := i.visibleBoundary(index)
	target := startWidth + deleteCount
	end, _ := i.visibleBoundary(min(target, i.utf16Width))

	result := sequenceIndexRange{start: start, end: end}
	if start > 0 {
		previous := i.visibleInsertion(start - 1)
		result.previous = new(previous.ID)
	}
	if end > start {
		result.targets = make([]opset.Operation, 0, end-start)
		i.eachVisible(start, end, func(entry sequenceIndexEntry) {
			if operation, ok := i.state.operation(entry.insertion); ok {
				result.targets = append(result.targets, operation)
			}
		})
	}

	return result, nil
}

// richPosition resolves an exact UTF-16 boundary for mark and block APIs.
// Unlike splice ranges, positions inside surrogate pairs are rejected.
func (i *sequenceIndex) richPosition(
	index uint32,
) (*opset.Operation, *opset.OpID, error) {
	if index > i.utf16Width {
		return nil, nil, fmt.Errorf("rich-text index %d is out of bounds", index)
	}

	visible, width := i.visibleBoundary(index)
	if width != index {
		return nil, nil, fmt.Errorf(
			"rich-text index %d splits a Unicode character or block at %d",
			index,
			width,
		)
	}

	var previous *opset.OpID
	if visible > 0 {
		operation := i.visibleInsertion(visible - 1)
		previous = new(operation.ID)
	}
	if visible == i.visibleCount {
		return nil, previous, nil
	}

	operation := i.visibleInsertion(visible)
	return &operation, previous, nil
}

// visibleBoundary returns the first element boundary at or after a UTF-16
// position and the width at that boundary. A position inside a surrogate pair
// advances to the following boundary.
func (i *sequenceIndex) visibleBoundary(target uint32) (int, uint32) {
	var (
		position uint32
		visible  int
	)

	for _, chunk := range i.chunks {
		if position+chunk.utf16Width < target {
			position += chunk.utf16Width
			visible += chunk.visibleCount
			continue
		}

		local := target - position
		boundary := sort.Search(
			len(chunk.prefixWidths),
			func(index int) bool {
				return chunk.prefixWidths[index] >= local
			},
		)

		return visible + int(chunk.prefixValues[boundary]),
			position + chunk.prefixWidths[boundary]
	}

	return visible, position
}

func (i *sequenceIndex) visibleInsertion(target int) opset.Operation {
	var visible int
	for _, chunk := range i.chunks {
		if visible+chunk.visibleCount <= target {
			visible += chunk.visibleCount
			continue
		}
		for _, entry := range chunk.entries {
			if !entry.visible {
				continue
			}
			if visible == target {
				operation, _ := i.state.operation(entry.insertion)
				return operation
			}
			visible++
		}
	}

	return opset.Operation{}
}

// rawPosition returns an insertion's position in the complete RGA order,
// including tombstones and zero-width mark boundaries. Chunk summaries keep
// this proportional to the number of chunks rather than the sequence length.
func (i *sequenceIndex) rawPosition(target opset.OpID) (int, bool) {
	position, ok := i.positions[target]
	if !ok {
		return 0, false
	}

	offset := 0
	for _, chunk := range i.chunks {
		if chunk == position.chunk {
			return offset + position.offset, true
		}
		offset += len(chunk.entries)
	}

	return 0, false
}

// eachRaw walks a localized raw-order range. Marks and tombstones are included
// because both influence insertion anchoring despite occupying no visible
// UTF-16 width.
func (i *sequenceIndex) eachRaw(start int, yield func(sequenceIndexEntry) bool) {
	position := 0
	for _, chunk := range i.chunks {
		if position+len(chunk.entries) <= start {
			position += len(chunk.entries)
			continue
		}
		for offset := max(start-position, 0); offset < len(chunk.entries); offset++ {
			if !yield(chunk.entries[offset]) {
				return
			}
		}
		position += len(chunk.entries)
	}
}

func (i *sequenceIndex) eachVisible(
	start int,
	end int,
	yield func(sequenceIndexEntry),
) {
	visible := 0
	for _, chunk := range i.chunks {
		if visible+chunk.visibleCount <= start {
			visible += chunk.visibleCount
			continue
		}
		for _, entry := range chunk.entries {
			if !entry.visible {
				continue
			}
			if visible >= end {
				return
			}
			if visible >= start {
				yield(entry)
			}
			visible++
		}
	}
}

func (i *sequenceIndex) replaceEntry(
	position sequenceIndexPosition,
	entry sequenceIndexEntry,
) {
	if i.chunkIndex(position.chunk) < 0 {
		return
	}

	chunk := position.chunk
	i.removeChunkSummary(chunk)
	chunk.entries[position.offset] = entry
	refreshSequenceIndexChunk(chunk)
	i.addChunkSummary(chunk)
	i.indexChunk(chunk)
}

func (i *sequenceIndex) insertEntry(
	position int,
	entry sequenceIndexEntry,
) bool {
	if len(i.chunks) == 0 {
		i.appendChunk(newSequenceIndexChunk([]sequenceIndexEntry{entry}))
		return true
	}

	var seen int
	for chunkIndex, chunk := range i.chunks {
		if position > seen+len(chunk.entries) {
			seen += len(chunk.entries)
			continue
		}

		offset := position - seen
		if len(chunk.entries)+1 <= sequenceIndexChunkMaximum {
			i.removeChunkSummary(chunk)
			chunk.entries = append(chunk.entries, sequenceIndexEntry{})
			copy(chunk.entries[offset+1:], chunk.entries[offset:])
			chunk.entries[offset] = entry
			refreshSequenceIndexChunk(chunk)
			i.addChunkSummary(chunk)
			i.indexChunk(chunk)
			return true
		}

		entries := make([]sequenceIndexEntry, 0, len(chunk.entries)+1)
		entries = append(entries, chunk.entries[:offset]...)
		entries = append(entries, entry)
		entries = append(entries, chunk.entries[offset:]...)
		middle := len(entries) / 2
		left := newSequenceIndexChunk(append([]sequenceIndexEntry(nil), entries[:middle]...))
		right := newSequenceIndexChunk(append([]sequenceIndexEntry(nil), entries[middle:]...))
		i.removeChunkSummary(chunk)
		i.chunks = append(i.chunks, nil)
		copy(i.chunks[chunkIndex+2:], i.chunks[chunkIndex+1:])
		i.chunks[chunkIndex] = left
		i.chunks[chunkIndex+1] = right
		i.addChunkSummary(left)
		i.addChunkSummary(right)
		i.indexChunk(left)
		i.indexChunk(right)

		return true
	}

	return false
}

func (i *sequenceIndex) insertEntryDeferred(
	position int,
	entry sequenceIndexEntry,
	dirty map[*sequenceIndexChunk]struct{},
) bool {
	if len(i.chunks) == 0 {
		chunk := newSequenceIndexChunk([]sequenceIndexEntry{entry})
		i.chunks = append(i.chunks, chunk)
		i.indexChunk(chunk)
		dirty[chunk] = struct{}{}
		return true
	}

	seen := 0
	for chunkIndex, chunk := range i.chunks {
		if position > seen+len(chunk.entries) {
			seen += len(chunk.entries)
			continue
		}

		offset := position - seen
		chunk.entries = append(chunk.entries, sequenceIndexEntry{})
		copy(chunk.entries[offset+1:], chunk.entries[offset:])
		chunk.entries[offset] = entry
		if len(chunk.entries) <= sequenceIndexChunkMaximum {
			i.indexChunkFrom(chunk, offset)
			dirty[chunk] = struct{}{}
			return true
		}

		middle := len(chunk.entries) / 2
		left := newSequenceIndexChunk(
			append([]sequenceIndexEntry(nil), chunk.entries[:middle]...),
		)
		right := newSequenceIndexChunk(
			append([]sequenceIndexEntry(nil), chunk.entries[middle:]...),
		)
		i.chunks = append(i.chunks, nil)
		copy(i.chunks[chunkIndex+2:], i.chunks[chunkIndex+1:])
		i.chunks[chunkIndex] = left
		i.chunks[chunkIndex+1] = right
		delete(dirty, chunk)
		dirty[left] = struct{}{}
		dirty[right] = struct{}{}
		i.indexChunk(left)
		i.indexChunk(right)
		return true
	}

	return false
}

func (i *sequenceIndex) finishDeferredMutations(
	dirty map[*sequenceIndexChunk]struct{},
) {
	for chunk := range dirty {
		refreshSequenceIndexChunk(chunk)
	}
	i.visibleCount = 0
	i.utf16Width = 0
	i.markCount = 0
	for _, chunk := range i.chunks {
		i.addChunkSummary(chunk)
	}
}

func refreshSequenceIndexChunk(chunk *sequenceIndexChunk) {
	required := len(chunk.entries) + 1
	if cap(chunk.prefixWidths) >= required {
		chunk.prefixWidths = chunk.prefixWidths[:required]
		clear(chunk.prefixWidths)
	} else {
		chunk.prefixWidths = make([]uint32, required)
	}
	if cap(chunk.prefixValues) >= required {
		chunk.prefixValues = chunk.prefixValues[:required]
		clear(chunk.prefixValues)
	} else {
		chunk.prefixValues = make([]uint16, required)
	}
	chunk.visibleCount = 0
	chunk.utf16Width = 0
	chunk.markCount = 0
	for index, entry := range chunk.entries {
		if entry.mark {
			chunk.markCount++
		}
		if entry.visible {
			chunk.visibleCount++
			chunk.utf16Width += entry.width
		}
		chunk.prefixWidths[index+1] = chunk.utf16Width
		chunk.prefixValues[index+1] = uint16(chunk.visibleCount)
	}
}

func (i *sequenceIndex) replaceChunk(index int, replacement *sequenceIndexChunk) {
	current := i.chunks[index]
	i.removeChunkSummary(current)
	i.chunks[index] = replacement
	i.addChunkSummary(replacement)
	i.indexChunk(replacement)
}

func (i *sequenceIndex) removeChunkSummary(chunk *sequenceIndexChunk) {
	i.visibleCount -= chunk.visibleCount
	i.utf16Width -= chunk.utf16Width
	i.markCount -= chunk.markCount
}

func (i *sequenceIndex) addChunkSummary(chunk *sequenceIndexChunk) {
	i.visibleCount += chunk.visibleCount
	i.utf16Width += chunk.utf16Width
	i.markCount += chunk.markCount
}

func (i *sequenceIndex) chunkIndex(target *sequenceIndexChunk) int {
	for index, chunk := range i.chunks {
		if chunk == target {
			return index
		}
	}

	return -1
}

func (i *sequenceIndex) insertionPosition(operation opset.Operation, local bool) (int, bool) {
	if len(i.positions) == 0 {
		return 0, operation.Key.IsHead
	}

	position := 0
	if operation.Key.IsHead {
		if local {
			return 0, true
		}
	} else {
		if operation.Key.Element == nil {
			return 0, false
		}
		var ok bool
		position, ok = i.rawPosition(*operation.Key.Element)
		if !ok {
			return 0, false
		}
		position++
	}

	if local {
		return position, true
	}

	// Remote siblings sort by descending operation ID, with each sibling's
	// descendants immediately following it. Walk only the branches rooted at
	// this anchor until the incoming branch's ordered position is reached.
	i.eachRaw(position, func(entry sequenceIndexEntry) bool {
		candidate, exists := i.state.operation(entry.insertion)
		if !exists {
			return true
		}
		root, descendant := directSequenceBranchRoot(i.state, candidate, operation)
		if !descendant || root.Compare(operation.ID) < 0 {
			return false
		}
		position++
		return true
	})

	return position, true
}

func (s *State) updateSequenceIndex(operation opset.Operation, local bool) {
	if operation.Object.IsRoot {
		return
	}

	object := operation.Object.OpID
	delete(s.sequenceCache, object)
	delete(s.sequenceValuesCache, object)
	delete(s.sequenceElementsCache, object)
	delete(s.sequenceOffsetCache, object)

	index, ok := s.sequenceIndexes[object]
	if !ok {
		return
	}
	creator, ok := s.operation(object)
	if !ok || creator.Action != opset.ActionMakeText {
		s.invalidateSequenceIndex(object)
		return
	}
	if operation.Insert {
		position, safe := index.insertionPosition(operation, local)
		if !safe {
			s.invalidateSequenceIndex(object)
			return
		}

		entry := s.newSequenceIndexEntry(operation)
		if !index.insertEntry(position, entry) {
			s.invalidateSequenceIndex(object)
		}

		return
	}

	if operation.Key.Element == nil {
		return
	}
	position, ok := index.positions[*operation.Key.Element]
	if !ok {
		s.invalidateSequenceIndex(object)
		return
	}

	entry := position.chunk.entries[position.offset]
	insertion, ok := s.operation(entry.insertion)
	if !ok {
		s.invalidateSequenceIndex(object)
		return
	}
	entry = s.newSequenceIndexEntry(insertion)
	index.replaceEntry(position, entry)
}

// updateSequenceIndexes applies a pending operation batch to the transaction's
// chunk overlay. A chunk touched by both insertion and deletion is summarized
// once after the full splice, which avoids rebuilding its prefix arrays for
// every semantic operation in the edit.
func (s *State) updateSequenceIndexes(
	operations []opset.Operation,
	local bool,
) {
	dirtyByObject := make(map[opset.OpID]map[*sequenceIndexChunk]struct{})
	for _, operation := range operations {
		if operation.Object.IsRoot {
			continue
		}
		object := operation.Object.OpID
		delete(s.sequenceCache, object)
		delete(s.sequenceValuesCache, object)
		delete(s.sequenceElementsCache, object)
		delete(s.sequenceOffsetCache, object)

		index, ok := s.sequenceIndexes[object]
		if !ok {
			continue
		}
		creator, ok := s.operation(object)
		if !ok || creator.Action != opset.ActionMakeText {
			s.invalidateSequenceIndex(object)
			delete(dirtyByObject, object)
			continue
		}
		dirty := dirtyByObject[object]
		if dirty == nil {
			dirty = make(map[*sequenceIndexChunk]struct{})
			dirtyByObject[object] = dirty
		}

		if operation.Insert {
			position, safe := index.insertionPosition(operation, local)
			if !safe || !index.insertEntryDeferred(
				position,
				s.newSequenceIndexEntry(operation),
				dirty,
			) {
				s.invalidateSequenceIndex(object)
				delete(dirtyByObject, object)
			}
			continue
		}
		if operation.Key.Element == nil {
			continue
		}
		position, ok := index.positions[*operation.Key.Element]
		if !ok {
			s.invalidateSequenceIndex(object)
			delete(dirtyByObject, object)
			continue
		}
		insertion, ok := s.operation(position.chunk.entries[position.offset].insertion)
		if !ok {
			s.invalidateSequenceIndex(object)
			delete(dirtyByObject, object)
			continue
		}
		position.chunk.entries[position.offset] = s.newSequenceIndexEntry(insertion)
		dirty[position.chunk] = struct{}{}
	}
	for object, dirty := range dirtyByObject {
		if index, ok := s.sequenceIndexes[object]; ok {
			index.finishDeferredMutations(dirty)
		}
	}
}

func (s *State) invalidateSequenceIndex(object opset.OpID) {
	delete(s.sequenceIndexes, object)
	delete(s.sequenceCache, object)
	delete(s.insertOrderCache, object)
	delete(s.insertOrderPositionCache, object)
	delete(s.sequenceValuesCache, object)
	delete(s.sequenceElementsCache, object)
	delete(s.sequenceOffsetCache, object)
}
