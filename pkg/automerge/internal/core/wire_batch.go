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
	"slices"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

type wireObjectSpan struct {
	start int
	end   int
}

func operationSpliceRuns(
	current []opset.Operation,
	next []opset.Operation,
) ([]storage.SnapshotOperationSplice, bool) {
	runs := make([]storage.SnapshotOperationSplice, 0)
	currentIndex := 0
	nextIndex := 0
	for currentIndex < len(current) && nextIndex < len(next) {
		if wireOperationsEqual(current[currentIndex], next[nextIndex]) {
			currentIndex++
			nextIndex++
			continue
		}
		if current[currentIndex].ID == next[nextIndex].ID {
			startCurrent := currentIndex
			startNext := nextIndex
			for currentIndex < len(current) &&
				nextIndex < len(next) &&
				current[currentIndex].ID == next[nextIndex].ID &&
				!wireOperationsEqual(current[currentIndex], next[nextIndex]) {
				currentIndex++
				nextIndex++
			}
			runs = append(runs, storage.SnapshotOperationSplice{
				Index:       startCurrent,
				DeleteCount: currentIndex - startCurrent,
				Operations:  next[startNext:nextIndex],
			})
			continue
		}

		startNext := nextIndex
		for nextIndex < len(next) &&
			next[nextIndex].ID != current[currentIndex].ID {
			nextIndex++
		}
		if nextIndex == len(next) {
			return nil, false
		}
		runs = append(runs, storage.SnapshotOperationSplice{
			Index:      currentIndex,
			Operations: next[startNext:nextIndex],
		})
	}
	if currentIndex != len(current) {
		return nil, false
	}
	if nextIndex < len(next) {
		runs = append(runs, storage.SnapshotOperationSplice{
			Index:      currentIndex,
			Operations: next[nextIndex:],
		})
	}
	slices.Reverse(runs)

	return runs, true
}

// planObjectBatch computes canonical order only for objects touched by an
// incoming batch. Object-creation batches retain the checked global-order
// fallback until their new spans can be inserted atomically.
func (c *columnarState) planObjectBatch(
	state *State,
	changes []*opset.Change,
) ([]opset.Operation, *operationRowIndex, bool) {
	touchedSet := make(map[opset.ObjectID]struct{})
	incomingRows := make(map[opset.ObjectID]int)
	for _, change := range changes {
		for _, operation := range change.Operations {
			if isObjectAction(operation.Action) {
				return nil, nil, false
			}
			touchedSet[operation.Object] = struct{}{}
			if operation.Action != opset.ActionDelete {
				incomingRows[operation.Object]++
			}
		}
	}
	if len(touchedSet) == 0 {
		return c.operations, c.operationRows, true
	}

	touched := make([]opset.ObjectID, 0, len(touchedSet))
	for object := range touchedSet {
		if _, ok := c.objectSpans[object]; !ok {
			return nil, nil, false
		}
		touched = append(touched, object)
	}
	slices.SortFunc(
		touched,
		func(left, right opset.ObjectID) int {
			return c.objectSpans[left].start - c.objectSpans[right].start
		},
	)

	operations := make([]opset.Operation, 0, len(c.operations)+64)
	cursor := 0
	for _, object := range touched {
		span := c.objectSpans[object]
		order, ok := state.indexedObjectOrder(object)
		if !ok {
			return nil, nil, false
		}
		// Partial/orphan query indexes cannot account for every retained row.
		// Detect them without rebuilding global operation order.
		if len(order) != span.end-span.start+incomingRows[object] {
			return nil, nil, false
		}

		replacement := make([]opset.Operation, len(order))
		for i, identifier := range order {
			operation, ok := state.operation(identifier)
			if !ok {
				return nil, nil, false
			}
			operation.Successors = append(
				[]opset.OpID(nil),
				state.successorIndex[identifier]...,
			)
			slices.SortFunc(
				operation.Successors,
				func(left, right opset.OpID) int {
					return left.Compare(right)
				},
			)
			replacement[i] = operation
		}

		current := c.operations[span.start:span.end]
		prefix := 0
		for prefix < len(current) &&
			prefix < len(replacement) &&
			wireOperationsEqual(current[prefix], replacement[prefix]) {
			prefix++
		}
		suffix := 0
		for suffix < len(current)-prefix &&
			suffix < len(replacement)-prefix &&
			wireOperationsEqual(
				current[len(current)-1-suffix],
				replacement[len(replacement)-1-suffix],
			) {
			suffix++
		}

		operations = append(
			operations,
			c.operations[cursor:span.start+prefix]...,
		)
		for _, operation := range replacement[prefix : len(replacement)-suffix] {
			operations = append(operations, cloneOperation(operation))
		}
		cursor = span.end - suffix
	}
	operations = append(operations, c.operations[cursor:]...)

	return operations, nil, true
}

func (s *State) indexedObjectOrder(
	object opset.ObjectID,
) ([]opset.OpID, bool) {
	if object.IsRoot {
		return s.indexedMapObjectOrder(object), true
	}
	creator, ok := s.operation(object.OpID)
	if !ok {
		return nil, false
	}
	if isMapObject(creator.Action) {
		return s.indexedMapObjectOrder(object), true
	}
	return s.indexedSequenceObjectOrder(object), true
}

func (s *State) indexedMapObjectOrder(object opset.ObjectID) []opset.OpID {
	if !s.mapKeyIndexBuilt {
		s.mapKeyIndexBuilt = true
		s.eachOperation(func(operation opset.Operation) bool {
			s.indexMapKeyOperation(operation)
			return true
		})
	}
	properties := s.mapKeyIndex[object]
	keys := make([]string, 0, len(properties))
	for property := range properties {
		keys = append(keys, property)
	}
	slices.Sort(keys)

	order := make([]opset.OpID, 0)
	for _, property := range keys {
		identifiers := append([]opset.OpID(nil), properties[property]...)
		identifiers = slices.DeleteFunc(
			identifiers,
			func(identifier opset.OpID) bool {
				operation, ok := s.operation(identifier)
				return !ok || operation.Action == opset.ActionDelete
			},
		)
		slices.SortFunc(
			identifiers,
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
		order = append(order, identifiers...)
	}

	return order
}

func (s *State) indexedSequenceObjectOrder(
	object opset.ObjectID,
) []opset.OpID {
	elements := s.insertOrder(object.OpID)
	order := make([]opset.OpID, 0, len(elements))
	for _, element := range elements {
		if operation, ok := s.operation(element); ok &&
			operation.Action != opset.ActionDelete {
			order = append(order, element)
		}

		identifiers := append(
			[]opset.OpID(nil),
			s.sequenceElementIndex[element]...,
		)
		identifiers = slices.DeleteFunc(
			identifiers,
			func(identifier opset.OpID) bool {
				operation, ok := s.operation(identifier)
				return !ok ||
					operation.Object != object ||
					operation.Insert ||
					operation.Action == opset.ActionDelete
			},
		)
		slices.SortFunc(
			identifiers,
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
		order = append(order, identifiers...)
	}

	return order
}
