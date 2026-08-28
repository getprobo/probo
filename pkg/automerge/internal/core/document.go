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
)

// compact serializes the whole history as one document chunk, the form save()
// produces in the other implementations, followed by any retained orphan changes
// as trailing change chunks (which is how a snapshot carries changes it cannot
// place in the operation set because their dependencies are missing).
//
// It reports ok=false, rather than an error, when the history cannot be
// compacted: while isolated, where the pinned view is not the whole history, or
// when the change graph is not internally consistent. The caller then falls back
// to the faithful change stream, which preserves every byte that was loaded.
func (b *Engine) compact(retainOrphans, deflate bool) ([]byte, bool, error) {
	if b.isolationActive {
		return nil, false, nil
	}

	if _, ok := b.state.allChanges(); !ok {
		return nil, false, nil
	}

	if b.columns == nil {
		return nil, false, nil
	}

	data, err := b.columns.snapshot.Encode(
		b.unknownColumns,
		deflate,
	)
	if err != nil {
		return nil, false, err
	}

	if retainOrphans {
		for _, change := range orderedQueuedChanges(b.queuedChanges) {
			data = append(data, maybeCompressChangeChunk(change.Raw, deflate)...)
		}
	}

	return data, true, nil
}

// documentOperationOrder returns the operation-set order a document chunk is
// written in: the root map first, then every object in identifier order, with a
// map's operations grouped by property and a sequence's following the order a
// reader sees. Deletes are left out because a snapshot records them only as
// successors of what they removed.
func (s *State) documentOperationOrder() []opset.OpID {
	order := make([]opset.OpID, 0, s.operationCount())

	for _, object := range s.documentObjects() {
		operation, _ := s.operation(object.OpID)
		if object.IsRoot || isMapObject(operation.Action) {
			order = append(order, s.mapObjectOrder(object)...)

			continue
		}

		order = append(order, s.sequenceObjectOrder(object)...)
	}

	return order
}

// documentObjects lists the root map followed by every object the history
// creates, ordered by the identifier of the operation that made it.
func (s *State) documentObjects() []opset.ObjectID {
	objects := make([]opset.ObjectID, 0)

	s.eachOperation(func(operation opset.Operation) bool {
		if isObjectAction(operation.Action) {
			objects = append(objects, opset.ObjectID{OpID: operation.ID})
		}

		return true
	})

	slices.SortFunc(
		objects,
		func(left, right opset.ObjectID) int {
			return left.OpID.Compare(right.OpID)
		},
	)

	return append([]opset.ObjectID{opset.RootObject()}, objects...)
}

func (s *State) mapObjectOrder(object opset.ObjectID) []opset.OpID {
	byProperty := make(map[string][]opset.OpID)

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object != object ||
			operation.Key.Property == nil ||
			operation.Action == opset.ActionDelete {
			return true
		}

		property := *operation.Key.Property
		byProperty[property] = append(byProperty[property], operation.ID)

		return true
	})

	properties := make([]string, 0, len(byProperty))
	for property := range byProperty {
		properties = append(properties, property)
	}

	slices.Sort(properties)

	order := make([]opset.OpID, 0, s.operationCount())

	for _, property := range properties {
		identifiers := byProperty[property]

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

func (s *State) sequenceObjectOrder(object opset.ObjectID) []opset.OpID {
	// Operations that address an element rather than create it, such as an
	// overwrite, follow the element they target.
	byElement := make(map[opset.OpID][]opset.OpID)

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object != object ||
			operation.Insert ||
			operation.Key.Element == nil ||
			operation.Action == opset.ActionDelete {
			return true
		}

		element := *operation.Key.Element
		byElement[element] = append(byElement[element], operation.ID)

		return true
	})

	for element := range byElement {
		slices.SortFunc(
			byElement[element],
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
	}

	elements := s.insertOrder(object.OpID)
	order := make([]opset.OpID, 0, len(elements))

	for _, element := range elements {
		if operation, ok := s.operation(element); ok && operation.Action != opset.ActionDelete {
			order = append(order, element)
		}

		order = append(order, byElement[element]...)
	}

	return order
}

func isObjectAction(action opset.Action) bool {
	switch action {
	case opset.ActionMakeMap, opset.ActionMakeList, opset.ActionMakeText, opset.ActionMakeTable:
		return true
	default:
		return false
	}
}

func isMapObject(action opset.Action) bool {
	return action == opset.ActionMakeMap || action == opset.ActionMakeTable
}
