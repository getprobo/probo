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
	"reflect"
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

	RichSpan struct {
		Type  string         `json:"type"`
		Value any            `json:"value"`
		Marks map[string]any `json:"marks,omitempty"`
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
	operations := s.sequenceElements(object)

	result := operations[:0]
	for _, operation := range operations {
		if operation.Action == ActionSet {
			result = append(result, operation)
		}
	}

	return result
}

func (s *State) sequenceElements(object OpID) []Operation {
	children := make(map[OpID][]Operation)

	var head []Operation

	for _, operation := range s.operations {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert ||
			operation.Action == ActionMark {
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
	s.appendSequence(
		&operations,
		head,
		children,
		make(map[OpID]struct{}),
		false,
	)

	return operations
}

func (s *State) sequenceAll(object OpID) []Operation {
	children := make(map[OpID][]Operation)

	var head []Operation

	for _, operation := range s.operations {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			!operation.Insert ||
			operation.Action == ActionMark {
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

	return operations
}

func (s *State) RichTextSpans(object OpID) ([]RichSpan, error) {
	elements := s.sequenceElements(object)
	marks := s.richTextMarks(object, elements)
	spans := make([]RichSpan, 0)

	for i, operation := range elements {
		switch operation.Action {
		case ActionMakeMap:
			value, err := s.mapValue(operation.ID, make(map[OpID]struct{}))
			if err != nil {
				return nil, fmt.Errorf("cannot hydrate block %v: %w", operation.ID, err)
			}

			spans = append(spans, RichSpan{Type: "block", Value: value})
		case ActionSet:
			if operation.Value == nil || operation.Value.Type != ScalarString {
				continue
			}

			activeMarks := make(map[string]any)

			for _, mark := range marks {
				if i >= mark.start && i < mark.end {
					activeMarks[mark.name] = mark.value
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
	start int
	end   int
	name  string
	value any
}

func (s *State) richTextMarks(object OpID, elements []Operation) []richTextMark {
	positions := make(map[OpID]int, len(elements))
	for i, element := range elements {
		positions[element.ID] = i
	}

	markOperations := make([]Operation, 0)

	for _, operation := range s.operations {
		if !operation.Object.IsRoot &&
			operation.Object.OpID == object &&
			operation.Action == ActionMark &&
			!s.isSuperseded(operation.ID) {
			markOperations = append(markOperations, operation)
		}
	}

	sort.Slice(markOperations, func(i, j int) bool {
		return markOperations[i].ID.Compare(markOperations[j].ID) < 0
	})

	marks := make([]richTextMark, 0, len(markOperations)/2)
	for i := 0; i+1 < len(markOperations); i += 2 {
		begin := markOperations[i]

		end := markOperations[i+1]
		if begin.MarkName == nil || begin.Key.Element == nil || end.Key.Element == nil {
			continue
		}

		startPosition, startOK := positions[*begin.Key.Element]

		endPosition, endOK := positions[*end.Key.Element]
		if !startOK || !endOK || startPosition >= endPosition {
			continue
		}

		marks = append(
			marks,
			richTextMark{
				start: startPosition + 1,
				end:   endPosition + 1,
				name:  *begin.MarkName,
				value: scalarMaterializedValue(begin.Value),
			},
		)
	}

	return marks
}

func (s *State) mapValue(
	object OpID,
	visited map[OpID]struct{},
) (map[string]any, error) {
	if _, ok := visited[object]; ok {
		return nil, fmt.Errorf("object cycle detected")
	}

	visited[object] = struct{}{}
	defer delete(visited, object)

	properties := make(map[string][]Operation)

	for _, operation := range s.operations {
		if operation.Object.IsRoot ||
			operation.Object.OpID != object ||
			operation.Key.Property == nil ||
			s.isSuperseded(operation.ID) {
			continue
		}

		property := *operation.Key.Property
		properties[property] = append(properties[property], operation)
	}

	result := make(map[string]any, len(properties))
	for property, operations := range properties {
		sort.Slice(operations, func(i, j int) bool {
			return operations[i].ID.Compare(operations[j].ID) > 0
		})

		operation := operations[0]
		switch operation.Action {
		case ActionMakeMap:
			value, err := s.mapValue(operation.ID, visited)
			if err != nil {
				return nil, err
			}

			result[property] = value
		case ActionMakeList:
			result[property] = []any{}
		case ActionMakeText:
			var value strings.Builder

			for _, element := range s.sequence(operation.ID) {
				if element.Value != nil && element.Value.Type == ScalarString {
					value.WriteString(element.Value.String)
				}
			}

			result[property] = value.String()
		case ActionSet:
			result[property] = scalarMaterializedValue(operation.Value)
		}
	}

	return result, nil
}

func scalarMaterializedValue(value *Scalar) any {
	if value == nil {
		return nil
	}

	switch value.Type {
	case ScalarNull:
		return nil
	case ScalarFalse:
		return false
	case ScalarTrue:
		return true
	case ScalarUint:
		return value.Uint
	case ScalarInt, ScalarCounter, ScalarTimestamp:
		return value.Int
	case ScalarFloat64:
		return value.Float
	case ScalarString:
		return value.String
	case ScalarBytes:
		return append([]byte(nil), value.Bytes...)
	default:
		return append([]byte(nil), value.Raw...)
	}
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

func (s *State) hasChange(hash ChangeHash) bool {
	_, ok := s.changes[hash]
	return ok
}

func (s *State) hasDependencies(change *Change) bool {
	for _, dependency := range change.Dependencies {
		if !s.hasChange(dependency) {
			return false
		}
	}

	return true
}
