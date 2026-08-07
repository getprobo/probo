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
	State struct {
		changes       map[ChangeHash]*Change
		actorSequence map[ActorID]uint64
		operations    map[OpID]Operation
		superseded    map[OpID]struct{}
		heads         map[ChangeHash]struct{}
	}
)

func NewState() *State {
	return &State{
		changes:       make(map[ChangeHash]*Change),
		actorSequence: make(map[ActorID]uint64),
		operations:    make(map[OpID]Operation),
		superseded:    make(map[OpID]struct{}),
		heads:         make(map[ChangeHash]struct{}),
	}
}

func NewStateFromDocument(document *Document) (*State, error) {
	state := NewState()
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Sequence > state.actorSequence[change.Actor] {
			state.actorSequence[change.Actor] = change.Sequence
		}
		if change.Hash != nil {
			state.changes[*change.Hash] = change
		}
		for _, operation := range change.Operations {
			if _, exists := state.operations[operation.ID]; exists {
				return nil, fmt.Errorf("duplicate snapshot operation ID %v", operation.ID)
			}
			state.operations[operation.ID] = operation
			for _, successor := range operation.Successors {
				state.superseded[operation.ID] = struct{}{}
				if successor.Counter == 0 {
					return nil, fmt.Errorf("invalid zero successor for operation %v", operation.ID)
				}
			}
		}
	}
	for _, head := range document.Heads {
		state.heads[head] = struct{}{}
	}
	return state, nil
}

func (s *State) ApplyChange(change *Change) error {
	if change.Hash == nil {
		return fmt.Errorf("change hash is required")
	}
	if _, ok := s.changes[*change.Hash]; ok {
		return nil
	}
	for _, dependency := range change.Dependencies {
		if _, ok := s.changes[dependency]; !ok {
			return fmt.Errorf("missing change dependency %s", dependency)
		}
	}

	expectedSequence := s.actorSequence[change.Actor] + 1
	if change.Sequence != expectedSequence {
		return fmt.Errorf(
			"actor sequence is %d, expected %d",
			change.Sequence,
			expectedSequence,
		)
	}

	for _, operation := range change.Operations {
		if _, exists := s.operations[operation.ID]; exists {
			return fmt.Errorf("duplicate operation ID %v", operation.ID)
		}
	}
	for _, operation := range change.Operations {
		s.operations[operation.ID] = operation
		for _, predecessor := range operation.Predecessors {
			s.superseded[predecessor] = struct{}{}
		}
	}

	s.changes[*change.Hash] = change
	s.actorSequence[change.Actor] = change.Sequence
	for _, dependency := range change.Dependencies {
		delete(s.heads, dependency)
	}
	s.heads[*change.Hash] = struct{}{}
	return nil
}

func (s *State) Heads() []ChangeHash {
	heads := make([]ChangeHash, 0, len(s.heads))
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
	objectOperation, ok := s.visibleMapOperation(property, ActionMakeText)
	if !ok {
		return "", fmt.Errorf("text property %q does not exist", property)
	}

	sequence := s.sequence(objectOperation.ID)
	var output strings.Builder
	for _, operation := range sequence {
		if operation.Value != nil && operation.Value.Type == ScalarString {
			output.WriteString(operation.Value.String)
		}
	}
	return output.String(), nil
}

func (s *State) visibleMapOperation(property string, action Action) (Operation, bool) {
	var (
		result Operation
		found  bool
	)
	for _, operation := range s.operations {
		if !operation.Object.IsRoot ||
			operation.Key.Property == nil ||
			*operation.Key.Property != property ||
			operation.Action != action ||
			s.isSuperseded(operation.ID) {
			continue
		}
		if !found || operation.ID.Compare(result.ID) > 0 {
			result = operation
			found = true
		}
	}
	return result, found
}

func (s *State) visibleMapOperations(property string) []Operation {
	operations := make([]Operation, 0)
	for _, operation := range s.operations {
		if !operation.Object.IsRoot ||
			operation.Key.Property == nil ||
			*operation.Key.Property != property ||
			s.isSuperseded(operation.ID) {
			continue
		}
		operations = append(operations, operation)
	}
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].ID.Compare(operations[j].ID) < 0
	})
	return operations
}

func (s *State) sequence(object OpID) []Operation {
	children := make(map[OpID][]Operation)
	var head []Operation
	for _, operation := range s.operations {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert ||
			operation.Action != ActionSet {
			continue
		}
		switch {
		case operation.Key.IsHead:
			head = append(head, operation)
		case operation.Key.Element != nil:
			children[*operation.Key.Element] = append(
				children[*operation.Key.Element],
				operation,
			)
		}
	}

	operations := make([]Operation, 0)
	s.appendSequence(&operations, head, children, make(map[OpID]struct{}))
	return operations
}

func (s *State) appendSequence(
	output *[]Operation,
	operations []Operation,
	children map[OpID][]Operation,
	visited map[OpID]struct{},
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
		if !s.isSuperseded(operation.ID) {
			*output = append(*output, operation)
		}
		s.appendSequence(output, children[operation.ID], children, visited)
	}
}

func (s *State) isSuperseded(id OpID) bool {
	_, ok := s.superseded[id]
	return ok
}

func (s *State) maxOpGlobal() uint64 {
	var maximum uint64
	for id := range s.operations {
		if id.Counter > maximum {
			maximum = id.Counter
		}
	}
	return maximum
}

func (s *State) sequenceForActor(actor ActorID) uint64 {
	return s.actorSequence[actor]
}

func (s *State) applyPending(operations []Operation) error {
	for _, operation := range operations {
		if _, exists := s.operations[operation.ID]; exists {
			return fmt.Errorf("duplicate pending operation ID %v", operation.ID)
		}
		s.operations[operation.ID] = operation
		for _, predecessor := range operation.Predecessors {
			s.superseded[predecessor] = struct{}{}
		}
	}
	return nil
}

func (s *State) recordAppliedChange(change *Change) error {
	if change.Hash == nil {
		return fmt.Errorf("change hash is required")
	}
	for _, dependency := range change.Dependencies {
		delete(s.heads, dependency)
	}
	s.heads[*change.Hash] = struct{}{}
	s.changes[*change.Hash] = change
	s.actorSequence[change.Actor] = change.Sequence
	return nil
}
