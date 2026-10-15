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
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"sort"
)

type sequenceTail struct {
	index uint32
	last  opset.OpID
	valid bool
	safe  bool
}

func (s *State) observeLoadedSequenceOperation(operation opset.Operation) {
	if operation.Object.IsRoot {
		return
	}

	object := operation.Object.OpID

	tail := s.sequenceTailCache[object]
	if !tail.valid {
		tail.valid = true
		tail.safe = true
	}

	if operation.Action == opset.ActionMark || !operation.Insert {
		tail.safe = false
	} else {
		tail.index += elementLength(operation)
	}

	s.sequenceTailCache[object] = tail
}

// finalizeSequenceTails takes only the final insertion identifier from the
// canonical order; widths and safety were accumulated during normal load.
func (s *State) finalizeSequenceTails(order []opset.OpID) {
	remaining := make(map[opset.OpID]struct{})

	for object, tail := range s.sequenceTailCache {
		if tail.safe {
			remaining[object] = struct{}{}
		}
	}

	for i := len(order) - 1; i >= 0 && len(remaining) > 0; i-- {
		identifier := order[i]

		operation, ok := s.operation(identifier)
		if !ok || operation.Object.IsRoot || !operation.Insert {
			continue
		}

		object := operation.Object.OpID
		if _, needed := remaining[object]; !needed {
			continue
		}

		tail := s.sequenceTailCache[object]
		tail.last = operation.ID
		s.sequenceTailCache[object] = tail
		delete(remaining, object)
	}

	for object := range remaining {
		tail := s.sequenceTailCache[object]
		tail.safe = false
		s.sequenceTailCache[object] = tail
	}
}

func (s *State) sequence(object opset.OpID) []opset.Operation {
	values := s.sequenceIndex(object).values()
	operations := make([]opset.Operation, 0, len(values))
	for _, value := range values {
		if value.Operation.Action == opset.ActionSet {
			operations = append(operations, value.Operation)
		}
	}

	return operations
}

func (s *State) sequenceIndex(object opset.OpID) *sequenceIndex {
	if index, ok := s.sequenceIndexes[object]; ok {
		return index
	}

	index := s.buildSequenceIndex(object)
	s.sequenceIndexes[object] = index

	return index
}

func (s *State) sequenceAll(object opset.OpID) []opset.Operation {
	index := s.sequenceIndex(object)
	operations := make([]opset.Operation, 0, len(index.positions)-index.markCount)
	index.eachRaw(0, func(entry sequenceIndexEntry) bool {
		if operation, ok := s.operation(entry.insertion); ok && !entry.mark {
			operations = append(operations, operation)
		}

		return true
	})

	return operations
}

// insertOrder returns the RGA-ordered insertion operation IDs for a sequence
// object (including tombstones and marks), using the incremental cache
// when present and rebuilding from the operation set otherwise.
func (s *State) insertOrder(object opset.OpID) []opset.OpID {
	if cached, ok := s.insertOrderCache[object]; ok {
		return cached
	}

	if index, ok := s.sequenceIndexes[object]; ok {
		order := index.order()
		s.insertOrderCache[object] = order

		return order
	}

	children := make(map[opset.OpID][]opset.Operation)

	var head []opset.Operation

	s.eachOperation(func(operation opset.Operation) bool {
		// Mark begin and end operations occupy positions in the sequence so
		// insertions can anchor relative to them; element views filter them out.
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert {
			return true
		}

		if operation.Key.IsHead {
			head = append(head, operation)
		} else if operation.Key.Element != nil {
			children[*operation.Key.Element] = append(
				children[*operation.Key.Element],
				operation,
			)
		}

		return true
	})

	operations := make([]opset.Operation, 0)
	s.appendSequence(
		&operations,
		head,
		children,
		make(map[opset.OpID]struct{}),
		true,
	)

	order := make([]opset.OpID, len(operations))
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
func (s *State) spliceInsertOrder(operation opset.Operation) {
	if !operation.Insert || operation.Object.IsRoot {
		return
	}

	object := operation.Object.OpID

	order, ok := s.insertOrderCache[object]
	if !ok {
		return
	}

	if operation.Key.IsHead {
		s.insertOrderCache[object] = append([]opset.OpID{operation.ID}, order...)
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

		updated := append(order, opset.OpID{})
		copy(updated[position+1:], updated[position:])
		updated[position] = operation.ID
		s.insertOrderCache[object] = updated
		// An insertion in the middle shifts later positions; rebuild on demand.
		delete(s.insertOrderPositionCache, object)

		return
	}

	// The anchor is not present in the cached order; rebuild lazily.
	delete(s.insertOrderCache, object)
	delete(s.insertOrderPositionCache, object)
}

func (s *State) sequenceValues(object opset.OpID) []sequenceValue {
	return s.sequenceIndex(object).values()
}

// insertOrderPositions returns each insert-order operation's index. Insert order
// only ever grows, so a length mismatch is a sufficient staleness check.
func (s *State) insertOrderPositions(object opset.OpID) map[opset.OpID]int {
	order := s.insertOrder(object)

	if cached, ok := s.insertOrderPositionCache[object]; ok && len(cached) == len(order) {
		return cached
	}

	positions := make(map[opset.OpID]int, len(order))
	for index, id := range order {
		positions[id] = index
	}

	s.insertOrderPositionCache[object] = positions

	return positions
}

// visibleSequenceElementOperations returns every visible value operation whose
// list element is the given insertion, in ascending operation-ID order. This is
// the conflict set that a subsequent put, delete, or increment must reference as
// its predecessors, matching upstream Rust which references all visible ops.
func (s *State) visibleSequenceElementOperations(element opset.OpID) []opset.Operation {
	var result []opset.Operation

	insertion, insertionExists := s.operation(element)
	if insertionExists && !s.isSuperseded(insertion.ID) {
		result = append(result, insertion)
	}

	for _, identifier := range s.sequenceElementIndex[element] {
		operation, ok := s.operation(identifier)
		if !ok ||
			operation.Action == opset.ActionDelete ||
			operation.Action == opset.ActionIncrement ||
			s.isSuperseded(operation.ID) {
			continue
		}

		result = append(result, operation)
	}

	sort.Slice(
		result,
		func(i, j int) bool {
			return result[i].ID.Compare(result[j].ID) < 0
		},
	)

	return result
}

// sequenceConflicts returns every visible value operation at the given visible
// list index, i.e. the conflict set that get_all(index) exposes. The boolean is
// false when the index is out of range.
func (s *State) sequenceConflicts(object opset.OpID, index uint64) ([]opset.Operation, bool) {
	if index > uint64(^uint(0)>>1) {
		return nil, false
	}

	value, ok := s.sequenceIndex(object).valueAt(int(index))
	if !ok {
		return nil, false
	}

	return s.visibleSequenceElementOperations(value.Element), true
}

func (s *State) appendSequence(
	output *[]opset.Operation,
	operations []opset.Operation,
	children map[opset.OpID][]opset.Operation,
	visited map[opset.OpID]struct{},
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
