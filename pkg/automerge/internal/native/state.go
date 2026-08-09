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
	"slices"
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
		sequenceCache map[OpID][]Operation
	}

	RichSpan struct {
		Type  string         `json:"type"`
		Value any            `json:"value"`
		Marks map[string]any `json:"marks,omitempty"`
	}

	sequenceValue struct {
		Element   OpID
		Operation Operation
	}
)

func NewState() *State {
	return &State{
		changes:       make(map[ChangeHash]*Change),
		actorSequence: make(map[ActorID]uint64),
		operations:    make(map[OpID]Operation),
		superseded:    make(map[OpID]struct{}),
		heads:         make(map[ChangeHash]struct{}),
		sequenceCache: make(map[OpID][]Operation),
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
				if successor.Counter == 0 {
					return nil, fmt.Errorf("invalid zero successor for operation %v", operation.ID)
				}
			}
		}
	}

	for _, operation := range state.operations {
		state.supersedePredecessors(operation)

		for _, successor := range operation.Successors {
			successorOperation, ok := state.operations[successor]
			if !ok ||
				successorOperation.Action != ActionIncrement ||
				!isCounterOperation(operation) {
				state.superseded[operation.ID] = struct{}{}
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
		if !operation.Object.IsRoot {
			delete(s.sequenceCache, operation.Object.OpID)
		}

		s.supersedePredecessors(operation)
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
	return s.visibleMapObjectOperation(RootObject(), property, action)
}

func (s *State) visibleMapObjectOperation(
	object ObjectID,
	property string,
	action Action,
) (Operation, bool) {
	var (
		result Operation
		found  bool
	)

	for _, operation := range s.operations {
		if operation.Object != object ||
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

func (s *State) visibleMapObjectValue(
	object ObjectID,
	property string,
) (Operation, bool) {
	var (
		result Operation
		found  bool
	)

	for _, operation := range s.visibleMapObjectOperations(object, property) {
		if operation.Action == ActionIncrement {
			continue
		}

		if !found || operation.ID.Compare(result.ID) > 0 {
			result = operation
			found = true
		}
	}

	return result, found
}

func isCounterOperation(operation Operation) bool {
	return operation.Action == ActionSet &&
		operation.Value != nil &&
		operation.Value.Type == ScalarCounter
}

// supersedePredecessors marks the predecessors overwritten by operation. A
// regular operation supersedes all of its predecessors. An increment supersedes
// only its non-counter predecessors: incrementing a counter keeps it visible,
// but an increment that also references a conflicting non-counter value deletes
// that value, matching upstream Rust.
func (s *State) supersedePredecessors(operation Operation) {
	for _, predecessor := range operation.Predecessors {
		if operation.Action == ActionIncrement {
			if pred, ok := s.operations[predecessor]; ok && isCounterOperation(pred) {
				continue
			}
		}

		s.superseded[predecessor] = struct{}{}
	}
}

func (s *State) scalarValue(operation Operation) (Scalar, bool) {
	if operation.Action != ActionSet || operation.Value == nil {
		return Scalar{}, false
	}

	value := *operation.Value

	value.Bytes = append([]byte(nil), operation.Value.Bytes...)
	if value.Type != ScalarCounter {
		return value, true
	}

	for _, increment := range s.operations {
		if increment.Action != ActionIncrement ||
			increment.Value == nil ||
			s.isSuperseded(increment.ID) {
			continue
		}

		matches := false

		for _, predecessor := range increment.Predecessors {
			if predecessor == operation.ID {
				matches = true
				break
			}
		}

		if !matches {
			for _, successor := range operation.Successors {
				if successor == increment.ID {
					matches = true
					break
				}
			}
		}

		if matches {
			value.Int += increment.Value.Int
		}
	}

	return value, true
}

func (s *State) visibleMapOperations(property string) []Operation {
	return s.visibleMapObjectOperations(RootObject(), property)
}

func (s *State) visibleMapObjectOperations(
	object ObjectID,
	property string,
) []Operation {
	operations := make([]Operation, 0)
	for _, operation := range s.operations {
		if operation.Object != object ||
			operation.Key.Property == nil ||
			*operation.Key.Property != property ||
			operation.Action == ActionDelete ||
			operation.Action == ActionIncrement ||
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

func (s *State) mapLength(object ObjectID) uint64 {
	return uint64(len(s.mapKeys(object)))
}

func (s *State) mapKeys(object ObjectID) []string {
	properties := make(map[string]struct{})

	for _, operation := range s.operations {
		if operation.Object == object &&
			operation.Key.Property != nil &&
			operation.Action != ActionDelete &&
			!s.isSuperseded(operation.ID) {
			properties[*operation.Key.Property] = struct{}{}
		}
	}

	keys := make([]string, 0, len(properties))
	for property := range properties {
		keys = append(keys, property)
	}

	slices.Sort(keys)

	return keys
}

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

func (s *State) sequenceValues(object OpID) []sequenceValue {
	insertions := s.sequenceAll(object)
	values := make([]sequenceValue, 0, len(insertions))

	for _, insertion := range insertions {
		var (
			value Operation
			found bool
		)
		if !s.isSuperseded(insertion.ID) {
			value = insertion
			found = true
		}

		for _, operation := range s.operations {
			if operation.Insert ||
				operation.Action == ActionDelete ||
				operation.Action == ActionIncrement ||
				operation.Key.Element == nil ||
				*operation.Key.Element != insertion.ID ||
				s.isSuperseded(operation.ID) {
				continue
			}

			if !found || operation.ID.Compare(value.ID) > 0 {
				value = operation
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

	return values
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
	scalar *Scalar
}

// MarkRange is one active mark over a UTF-16 range of a text object.
type MarkRange struct {
	Start uint32
	End   uint32
	Name  string
	Value *Scalar
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

	byID := make(map[OpID]Operation, len(markOperations))
	for _, operation := range markOperations {
		byID[operation.ID] = operation
	}

	marks := make([]richTextMark, 0, len(markOperations)/2)
	for _, begin := range markOperations {
		if begin.MarkName == nil {
			continue
		}

		end, ok := byID[OpID{
			Actor:   begin.ID.Actor,
			Counter: begin.ID.Counter + 1,
		}]
		if !ok || end.MarkName != nil {
			continue
		}

		startPosition, startOK := s.markAnchorPosition(
			object,
			begin.Key,
			begin.MarkExpand != nil && *begin.MarkExpand,
			positions,
			make(map[OpID]struct{}),
		)

		endPosition, endOK := s.markAnchorPosition(
			object,
			end.Key,
			end.MarkExpand != nil && *end.MarkExpand,
			positions,
			make(map[OpID]struct{}),
		)
		if !startOK || !endOK || startPosition >= endPosition {
			continue
		}

		// If a mark started at the document head and every element that
		// existed in its original range has since been deleted, replacement
		// text inserted at the head must not inherit the mark. A visible
		// left-hand anchor (or any surviving original marked element) keeps
		// boundary expansion active; this distinction matches Rust's splice
		// behavior when replacing some marked text versus all of it.
		if begin.Key.IsHead &&
			!s.markRangeHasSurvivingElement(object, begin, end) {
			continue
		}

		marks = append(
			marks,
			richTextMark{
				start:  startPosition,
				end:    endPosition,
				name:   *begin.MarkName,
				value:  scalarMaterializedValue(begin.Value),
				scalar: begin.Value,
			},
		)
	}

	return marks
}

// Marks returns the active marks over a text object as UTF-16 ranges, matching
// upstream Rust's marks(): contiguous runs of an identical (name, value) mark
// are merged, block markers occupy one position, and marks removed by a null
// value are excluded.
func (s *State) Marks(object OpID) []MarkRange {
	elements := s.sequenceElements(object)
	marks := s.richTextMarks(object, elements)

	type openMark struct {
		start uint32
		value *Scalar
	}

	open := make(map[string]openMark)

	result := make([]MarkRange, 0)

	var position uint32

	closeMark := func(name string, mark openMark, end uint32) {
		result = append(result, MarkRange{
			Start: mark.start,
			End:   end,
			Name:  name,
			Value: mark.value,
		})
	}

	for index, element := range elements {
		active := make(map[string]*Scalar)

		for _, mark := range marks {
			if index < mark.start || index >= mark.end {
				continue
			}

			if mark.scalar == nil || mark.scalar.Type == ScalarNull {
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

	sort.Slice(result, func(i, j int) bool {
		if result[i].Start != result[j].Start {
			return result[i].Start < result[j].Start
		}

		return result[i].Name < result[j].Name
	})

	return result
}

func (s *State) markRangeHasSurvivingElement(
	object OpID,
	begin Operation,
	end Operation,
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

func (s *State) markAnchorPosition(
	object OpID,
	key Key,
	expand bool,
	positions map[OpID]int,
	visited map[OpID]struct{},
) (int, bool) {
	if key.IsHead {
		return 0, true
	}

	if key.Element == nil {
		return 0, false
	}

	if position, ok := positions[*key.Element]; ok {
		return position + 1, true
	}

	if _, ok := visited[*key.Element]; ok {
		return 0, false
	}

	visited[*key.Element] = struct{}{}

	operation, ok := s.operations[*key.Element]
	if !ok {
		return 0, false
	}

	if operation.Action == ActionMark {
		return s.markAnchorPosition(
			object,
			operation.Key,
			expand,
			positions,
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
		expand,
		positions,
		visited,
	)
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
			value, err := s.listValue(operation.ID, visited)
			if err != nil {
				return nil, err
			}

			result[property] = value
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

func (s *State) listValue(
	object OpID,
	visited map[OpID]struct{},
) ([]any, error) {
	if _, ok := visited[object]; ok {
		return nil, fmt.Errorf("object cycle detected")
	}

	visited[object] = struct{}{}
	defer delete(visited, object)

	elements := s.sequenceElements(object)
	result := make([]any, 0, len(elements))

	for _, element := range elements {
		switch element.Action {
		case ActionMakeMap:
			value, err := s.mapValue(element.ID, visited)
			if err != nil {
				return nil, err
			}

			result = append(result, value)
		case ActionMakeList:
			value, err := s.listValue(element.ID, visited)
			if err != nil {
				return nil, err
			}

			result = append(result, value)
		case ActionMakeText:
			var value strings.Builder

			for _, textElement := range s.sequence(element.ID) {
				if textElement.Value != nil &&
					textElement.Value.Type == ScalarString {
					value.WriteString(textElement.Value.String)
				}
			}

			result = append(result, value.String())
		case ActionSet:
			result = append(result, scalarMaterializedValue(element.Value))
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

// hashForActorSequence returns the hash of the change authored by actor at the
// given sequence number, if it is known.
func (s *State) hashForActorSequence(
	actor ActorID,
	sequence uint64,
) (ChangeHash, bool) {
	for hash, change := range s.changes {
		if change.Actor == actor && change.Sequence == sequence {
			return hash, true
		}
	}

	return ChangeHash{}, false
}

func (s *State) applyPending(operations []Operation) error {
	for _, operation := range operations {
		if _, exists := s.operations[operation.ID]; exists {
			return fmt.Errorf("duplicate pending operation ID %v", operation.ID)
		}

		s.operations[operation.ID] = operation
		if !operation.Object.IsRoot {
			delete(s.sequenceCache, operation.Object.OpID)
		}

		s.supersedePredecessors(operation)
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

func (s *State) changesSince(heads []ChangeHash) ([]*Change, bool) {
	known, ok := s.changeClosure(heads)
	if !ok {
		return nil, false
	}

	ordered := make([]*Change, 0)
	visited := make(map[ChangeHash]struct{})

	var visit func(ChangeHash) bool

	visit = func(hash ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		visited[hash] = struct{}{}

		change, ok := s.changes[hash]
		if !ok {
			return false
		}

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				return false
			}
		}

		if _, ok := known[hash]; ok {
			return true
		}

		if len(change.Raw) == 0 {
			return false
		}

		ordered = append(ordered, change)

		return true
	}

	for _, head := range s.Heads() {
		if !visit(head) {
			return nil, false
		}
	}

	return ordered, true
}

func (s *State) allChanges() ([]*Change, bool) {
	ordered := make([]*Change, 0, len(s.changes))
	visited := make(map[ChangeHash]struct{}, len(s.changes))

	var visit func(ChangeHash) bool

	visit = func(hash ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := s.changes[hash]
		if !ok {
			return false
		}

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				return false
			}
		}

		visited[hash] = struct{}{}

		ordered = append(ordered, change)

		return true
	}

	for _, head := range s.Heads() {
		if !visit(head) {
			return nil, false
		}
	}

	return ordered, len(visited) == len(s.changes)
}

func (s *State) at(heads []ChangeHash) (*State, bool) {
	target := NewState()
	visited := make(map[ChangeHash]struct{})

	var visit func(ChangeHash) bool

	visit = func(hash ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := s.changes[hash]
		if !ok {
			return false
		}

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				return false
			}
		}

		clone := *change

		clone.Hash = new(hash)
		if err := target.ApplyChange(&clone); err != nil {
			return false
		}

		visited[hash] = struct{}{}

		return true
	}

	for _, head := range heads {
		if !visit(head) {
			return nil, false
		}
	}

	return target, true
}

func (s *State) changeClosure(heads []ChangeHash) (map[ChangeHash]struct{}, bool) {
	closure := make(map[ChangeHash]struct{})

	pending := append([]ChangeHash(nil), heads...)

	for len(pending) > 0 {
		index := len(pending) - 1
		hash := pending[index]
		pending = pending[:index]

		if _, ok := closure[hash]; ok {
			continue
		}

		change, ok := s.changes[hash]
		if !ok {
			return nil, false
		}

		closure[hash] = struct{}{}

		pending = append(pending, change.Dependencies...)
	}

	return closure, true
}
