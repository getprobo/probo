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

package native

import (
	"sort"
)

func (s *State) sequence(object OpID) []Operation {
	if cached, ok := s.sequenceCache[object]; ok {
		return cached
	}

	operations := s.sequenceElements(object)

	result := operations[:0]
	for _, operation := range operations {
		if operation.Action == ActionSet {
			result = append(result, operation)
		}
	}

	s.sequenceCache[object] = result

	return result
}

func (s *State) setSequenceCache(object OpID, operations []Operation) {
	s.sequenceCache[object] = operations
}

func (s *State) sequenceElements(object OpID) []Operation {
	if cached, ok := s.sequenceElementsCache[object]; ok {
		return cached
	}

	order := s.insertOrder(object)

	operations := make([]Operation, 0, len(order))

	for _, id := range order {
		if s.isSuperseded(id) {
			continue
		}

		if operation, ok := s.operations[id]; ok && operation.Action != ActionMark {
			operations = append(operations, operation)
		}
	}

	s.sequenceElementsCache[object] = operations

	return operations
}

func (s *State) sequenceAll(object OpID) []Operation {
	order := s.insertOrder(object)

	operations := make([]Operation, 0, len(order))

	for _, id := range order {
		if operation, ok := s.operations[id]; ok && operation.Action != ActionMark {
			operations = append(operations, operation)
		}
	}

	return operations
}

// insertOrder returns the RGA-ordered insertion operation IDs for a sequence
// object (including tombstones, excluding marks), using the incremental cache
// when present and rebuilding from the operation set otherwise.
func (s *State) insertOrder(object OpID) []OpID {
	if cached, ok := s.insertOrderCache[object]; ok {
		return cached
	}

	children := make(map[OpID][]Operation)

	var head []Operation

	for _, operation := range s.operations {
		// Mark begin and end operations occupy positions in the sequence so
		// insertions can anchor relative to them; element views filter them out.
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert {
			continue
		}

		if operation.Key.IsHead {
			head = append(head, operation)
		} else if operation.Key.Element != nil {
			children[*operation.Key.Element] = append(
				children[*operation.Key.Element],
				operation,
			)
		}
	}

	operations := make([]Operation, 0)
	s.appendSequence(
		&operations,
		head,
		children,
		make(map[OpID]struct{}),
		true,
	)

	order := make([]OpID, len(operations))
	for i, operation := range operations {
		order[i] = operation.ID
	}

	s.insertOrderCache[object] = order

	return order
}

// spliceInsertOrder inserts a locally created insertion operation into the
// cached RGA order. Local operations always carry the current maximum operation
// ID, so they sort ahead of every existing sibling: a head-anchored insertion
// goes to the front and an element-anchored one goes immediately after its
// anchor. If the object's order has not been cached yet the splice is skipped
// and the order is rebuilt on the next read.
func (s *State) spliceInsertOrder(operation Operation) {
	if !operation.Insert || operation.Object.IsRoot {
		return
	}

	object := operation.Object.OpID

	order, ok := s.insertOrderCache[object]
	if !ok {
		return
	}

	if operation.Key.IsHead {
		s.insertOrderCache[object] = append([]OpID{operation.ID}, order...)
		// A prepend shifts every position, so the index is rebuilt on demand.
		delete(s.insertOrderPositionCache, object)

		return
	}

	if operation.Key.Element == nil {
		delete(s.insertOrderCache, object)
		delete(s.insertOrderPositionCache, object)

		return
	}

	anchor := *operation.Key.Element

	if len(order) > 0 && order[len(order)-1] == anchor {
		s.insertOrderCache[object] = append(order, operation.ID)

		// Appending at the end keeps every existing index, so extend the position
		// index in step to keep sequential typing constant time.
		if positions, ok := s.insertOrderPositionCache[object]; ok &&
			len(positions) == len(order) {
			positions[operation.ID] = len(order)
		}

		return
	}

	for i, id := range order {
		if id != anchor {
			continue
		}

		position := i + 1
		updated := make([]OpID, 0, len(order)+1)
		updated = append(updated, order[:position]...)
		updated = append(updated, operation.ID)
		updated = append(updated, order[position:]...)
		s.insertOrderCache[object] = updated
		// An insertion in the middle shifts later positions; rebuild on demand.
		delete(s.insertOrderPositionCache, object)

		return
	}

	// The anchor is not present in the cached order; rebuild lazily.
	delete(s.insertOrderCache, object)
	delete(s.insertOrderPositionCache, object)
}

func (s *State) sequenceValues(object OpID) []sequenceValue {
	if cached, ok := s.sequenceValuesCache[object]; ok {
		return cached
	}

	insertions := s.sequenceAll(object)
	values := make([]sequenceValue, 0, len(insertions))

	// Collect the winning replacement value per element in a single pass. Doing
	// this per insertion would rescan every operation for every element, which
	// is quadratic in the size of the document.
	winners := s.elementValueWinners()

	for _, insertion := range insertions {
		var (
			value Operation
			found bool
		)

		if !s.isSuperseded(insertion.ID) {
			value = insertion
			found = true
		}

		if replacement, ok := winners[insertion.ID]; ok {
			if !found || replacement.ID.Compare(value.ID) > 0 {
				value = replacement
				found = true
			}
		}

		if found {
			values = append(
				values,
				sequenceValue{
					Element:   insertion.ID,
					Operation: value,
				},
			)
		}
	}

	s.sequenceValuesCache[object] = values

	return values
}

// updateSequenceValues keeps the materialized sequence values coherent after an
// operation is applied. Appending a brand new element at the end of the
// sequence extends the cached slice, which keeps sequential editing linear;
// anything else (a replacement, a deletion, an insertion in the middle) can
// change which values win, so the entry is dropped and rebuilt on demand.
func (s *State) updateSequenceValues(operation Operation) {
	if operation.Object.IsRoot || operation.Action == ActionMark {
		return
	}

	object := operation.Object.OpID

	order := s.insertOrderCache[object]
	appended := operation.Insert &&
		len(operation.Predecessors) == 0 &&
		len(order) > 0 &&
		order[len(order)-1] == operation.ID

	if !appended {
		delete(s.sequenceValuesCache, object)
		delete(s.sequenceElementsCache, object)
		delete(s.sequenceOffsetCache, object)

		return
	}

	if cached, ok := s.sequenceValuesCache[object]; ok {
		s.sequenceValuesCache[object] = append(cached, sequenceValue{
			Element:   operation.ID,
			Operation: operation,
		})
	}

	if cached, ok := s.sequenceElementsCache[object]; ok {
		s.sequenceElementsCache[object] = append(cached, operation)

		// Extend the offset index in step so appending stays constant time. A
		// valid offset slice for N elements holds N+1 entries, so it lines up
		// with the pre-append element count; the new element starts at the
		// previous total width.
		if offsets, ok := s.sequenceOffsetCache[object]; ok &&
			len(offsets) == len(cached)+1 {
			s.sequenceOffsetCache[object] = append(
				offsets,
				offsets[len(offsets)-1]+elementLength(operation),
			)
		}
	}
}

// sequenceOffsets returns the cumulative UTF-16 width before each element of the
// given sequence, with a trailing entry holding the total width. It is cached
// and rebuilt whenever it does not line up with the elements, so a text index
// can be resolved by binary search rather than a linear walk.
func (s *State) sequenceOffsets(object OpID, elements []Operation) []uint32 {
	if cached, ok := s.sequenceOffsetCache[object]; ok && len(cached) == len(elements)+1 {
		return cached
	}

	offsets := make([]uint32, len(elements)+1)
	for i, operation := range elements {
		offsets[i+1] = offsets[i] + elementLength(operation)
	}

	s.sequenceOffsetCache[object] = offsets

	return offsets
}

// insertOrderPositions returns each insert-order operation's index. Insert order
// only ever grows, so a length mismatch is a sufficient staleness check.
func (s *State) insertOrderPositions(object OpID) map[OpID]int {
	order := s.insertOrder(object)

	if cached, ok := s.insertOrderPositionCache[object]; ok && len(cached) == len(order) {
		return cached
	}

	positions := make(map[OpID]int, len(order))
	for index, id := range order {
		positions[id] = index
	}

	s.insertOrderPositionCache[object] = positions

	return positions
}

// elementValueWinners returns, for every list element that has been assigned a
// replacement value, the visible operation with the highest ID. Element IDs are
// globally unique, so a single map covers every object.
func (s *State) elementValueWinners() map[OpID]Operation {
	winners := make(map[OpID]Operation)

	for _, operation := range s.operations {
		if operation.Insert ||
			operation.Action == ActionDelete ||
			operation.Action == ActionIncrement ||
			operation.Key.Element == nil ||
			s.isSuperseded(operation.ID) {
			continue
		}

		element := *operation.Key.Element
		if current, ok := winners[element]; !ok || operation.ID.Compare(current.ID) > 0 {
			winners[element] = operation
		}
	}

	return winners
}

// visibleSequenceElementOperations returns every visible value operation whose
// list element is the given insertion, in ascending operation-ID order. This is
// the conflict set that a subsequent put, delete, or increment must reference as
// its predecessors, matching upstream Rust which references all visible ops.
func (s *State) visibleSequenceElementOperations(element OpID) []Operation {
	var result []Operation

	if insertion, ok := s.operations[element]; ok && !s.isSuperseded(insertion.ID) {
		result = append(result, insertion)
	}

	for _, operation := range s.operations {
		if operation.Insert ||
			operation.Action == ActionDelete ||
			operation.Action == ActionIncrement ||
			operation.Key.Element == nil ||
			*operation.Key.Element != element ||
			s.isSuperseded(operation.ID) {
			continue
		}

		result = append(result, operation)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID.Compare(result[j].ID) < 0
	})

	return result
}

// sequenceConflicts returns every visible value operation at the given visible
// list index, i.e. the conflict set that get_all(index) exposes. The boolean is
// false when the index is out of range.
func (s *State) sequenceConflicts(object OpID, index uint64) ([]Operation, bool) {
	values := s.sequenceValues(object)
	if index >= uint64(len(values)) {
		return nil, false
	}

	return s.visibleSequenceElementOperations(values[index].Element), true
}

func (s *State) appendSequence(
	output *[]Operation,
	operations []Operation,
	children map[OpID][]Operation,
	visited map[OpID]struct{},
	includeSuperseded bool,
) {
	sort.Slice(
		operations,
		func(i, j int) bool {
			return operations[i].ID.Compare(operations[j].ID) > 0
		},
	)

	for _, operation := range operations {
		if _, ok := visited[operation.ID]; ok {
			continue
		}

		visited[operation.ID] = struct{}{}
		if includeSuperseded || !s.isSuperseded(operation.ID) {
			*output = append(*output, operation)
		}

		s.appendSequence(
			output,
			children[operation.ID],
			children,
			visited,
			includeSuperseded,
		)
	}
}
