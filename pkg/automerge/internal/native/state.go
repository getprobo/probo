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
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type (
	ResolvedOpID struct {
		Counter uint64
		Actor   string
	}

	resolvedKey struct {
		Map      bool
		Property string
		Element  ResolvedOpID
		Head     bool
	}

	resolvedOperation struct {
		ID           ResolvedOpID
		ObjectRoot   bool
		Object       ResolvedOpID
		Key          resolvedKey
		Insert       bool
		Action       Action
		Value        Scalar
		Predecessors []ResolvedOpID
	}

	State struct {
		changes       map[[32]byte]*Change
		actorSequence map[string]uint64
		operations    map[ResolvedOpID]resolvedOperation
		superseded    map[ResolvedOpID]struct{}
		heads         map[[32]byte]struct{}
	}
)

func NewState() *State {
	return &State{
		changes:       make(map[[32]byte]*Change),
		actorSequence: make(map[string]uint64),
		operations:    make(map[ResolvedOpID]resolvedOperation),
		superseded:    make(map[ResolvedOpID]struct{}),
		heads:         make(map[[32]byte]struct{}),
	}
}

func (s *State) ApplyChange(change *Change) error {
	if _, ok := s.changes[change.Hash]; ok {
		return nil
	}
	for _, dependency := range change.Dependencies {
		if _, ok := s.changes[dependency]; !ok {
			return fmt.Errorf("missing change dependency %x", dependency)
		}
	}

	actor := string(change.Actor)
	expectedSequence := s.actorSequence[actor] + 1
	if change.Sequence != expectedSequence {
		return fmt.Errorf(
			"actor sequence is %d, expected %d",
			change.Sequence,
			expectedSequence,
		)
	}

	operations, err := change.Operations()
	if err != nil {
		return fmt.Errorf("cannot decode change operations: %w", err)
	}
	resolved := make([]resolvedOperation, len(operations))
	for i, operation := range operations {
		resolved[i], err = resolveOperation(change, operation)
		if err != nil {
			return fmt.Errorf("cannot resolve operation %d: %w", i, err)
		}
		if _, exists := s.operations[resolved[i].ID]; exists {
			return fmt.Errorf("duplicate operation ID %v", resolved[i].ID)
		}
	}

	for _, operation := range resolved {
		s.operations[operation.ID] = operation
		for _, predecessor := range operation.Predecessors {
			s.superseded[predecessor] = struct{}{}
		}
	}
	s.changes[change.Hash] = change
	s.actorSequence[actor] = change.Sequence
	for _, dependency := range change.Dependencies {
		delete(s.heads, dependency)
	}
	s.heads[change.Hash] = struct{}{}
	return nil
}

func (s *State) Heads() [][32]byte {
	heads := make([][32]byte, 0, len(s.heads))
	for head := range s.heads {
		heads = append(heads, head)
	}
	sort.Slice(
		heads,
		func(i, j int) bool {
			return bytes.Compare(heads[i][:], heads[j][:]) < 0
		},
	)
	return heads
}

func (s *State) Text(property string) (string, error) {
	var (
		objectOperation resolvedOperation
		found           bool
	)
	for _, operation := range s.operations {
		if !operation.ObjectRoot ||
			!operation.Key.Map ||
			operation.Key.Property != property ||
			operation.Action != ActionMakeText ||
			s.isSuperseded(operation.ID) {
			continue
		}
		if !found || compareOpID(operation.ID, objectOperation.ID) > 0 {
			objectOperation = operation
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("text property %q does not exist", property)
	}

	children := make(map[ResolvedOpID][]resolvedOperation)
	var head []resolvedOperation
	for _, operation := range s.operations {
		if operation.ObjectRoot ||
			operation.Object != objectOperation.ID ||
			!operation.Insert ||
			operation.Action != ActionSet {
			continue
		}
		if operation.Key.Head {
			head = append(head, operation)
		} else {
			children[operation.Key.Element] = append(
				children[operation.Key.Element],
				operation,
			)
		}
	}

	var output strings.Builder
	s.appendSequence(&output, head, children, make(map[ResolvedOpID]struct{}))
	return output.String(), nil
}

func (s *State) appendSequence(
	output *strings.Builder,
	operations []resolvedOperation,
	children map[ResolvedOpID][]resolvedOperation,
	visited map[ResolvedOpID]struct{},
) {
	sort.Slice(
		operations,
		func(i, j int) bool {
			return compareOpID(operations[i].ID, operations[j].ID) > 0
		},
	)
	for _, operation := range operations {
		if _, ok := visited[operation.ID]; ok {
			continue
		}
		visited[operation.ID] = struct{}{}
		if !s.isSuperseded(operation.ID) {
			if value, ok := operation.Value.Value.(string); ok {
				output.WriteString(value)
			}
		}
		s.appendSequence(output, children[operation.ID], children, visited)
	}
}

func (s *State) isSuperseded(id ResolvedOpID) bool {
	_, ok := s.superseded[id]
	return ok
}

func resolveOperation(change *Change, operation Operation) (resolvedOperation, error) {
	id, err := resolveOpID(change, operation.ID)
	if err != nil {
		return resolvedOperation{}, fmt.Errorf("cannot resolve ID: %w", err)
	}

	resolved := resolvedOperation{
		ID:         id,
		ObjectRoot: operation.Object.Root,
		Key: resolvedKey{
			Map:      operation.Key.Map,
			Property: operation.Key.Property,
			Head:     operation.Key.Head,
		},
		Insert: operation.Insert,
		Action: operation.Action,
		Value:  operation.Value,
	}
	if !operation.Object.Root {
		resolved.Object, err = resolveOpID(change, operation.Object.OpID)
		if err != nil {
			return resolvedOperation{}, fmt.Errorf("cannot resolve object: %w", err)
		}
	}
	if !operation.Key.Map && !operation.Key.Head {
		resolved.Key.Element, err = resolveOpID(change, operation.Key.Element)
		if err != nil {
			return resolvedOperation{}, fmt.Errorf("cannot resolve element: %w", err)
		}
	}
	resolved.Predecessors = make([]ResolvedOpID, len(operation.Predecessors))
	for i, predecessor := range operation.Predecessors {
		resolved.Predecessors[i], err = resolveOpID(change, predecessor)
		if err != nil {
			return resolvedOperation{}, fmt.Errorf("cannot resolve predecessor %d: %w", i, err)
		}
	}
	return resolved, nil
}

func resolveOpID(change *Change, id OpID) (ResolvedOpID, error) {
	var actor []byte
	if id.ActorIndex == 0 {
		actor = change.Actor
	} else {
		index := id.ActorIndex - 1
		if index >= uint64(len(change.OtherActors)) {
			return ResolvedOpID{}, fmt.Errorf("actor index %d is out of bounds", id.ActorIndex)
		}
		actor = change.OtherActors[index]
	}
	return ResolvedOpID{Counter: id.Counter, Actor: string(actor)}, nil
}

func compareOpID(left, right ResolvedOpID) int {
	if left.Counter < right.Counter {
		return -1
	}
	if left.Counter > right.Counter {
		return 1
	}
	return bytes.Compare([]byte(left.Actor), []byte(right.Actor))
}
