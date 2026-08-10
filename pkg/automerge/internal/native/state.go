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

		// insertOrderCache holds, per sequence object, the RGA-ordered list of
		// insertion operation IDs (including tombstones, excluding marks). The
		// order depends only on insertion operations and their anchors, so it is
		// maintained incrementally: local inserts (whose IDs are always the
		// current maximum) splice in next to their anchor, while merged changes
		// and rollbacks invalidate the entry so it is rebuilt lazily.
		insertOrderCache map[OpID][]OpID

		// sequenceValuesCache holds the materialized visible values of a
		// sequence object, and sequenceElementsCache the materialized visible
		// elements. Appending a new element at the end extends them in place;
		// every other mutation drops the entry so it is recomputed.
		sequenceValuesCache   map[OpID][]sequenceValue
		sequenceElementsCache map[OpID][]Operation

		// mapKeyIndex groups operation IDs by the map property they address so
		// reading a key does not scan the whole operation set. It is built on
		// first use and then maintained as operations are applied.
		mapKeyIndex      map[ObjectID]map[string][]OpID
		mapKeyIndexBuilt bool
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
		changes:               make(map[ChangeHash]*Change),
		actorSequence:         make(map[ActorID]uint64),
		operations:            make(map[OpID]Operation),
		superseded:            make(map[OpID]struct{}),
		heads:                 make(map[ChangeHash]struct{}),
		sequenceCache:         make(map[OpID][]Operation),
		insertOrderCache:      make(map[OpID][]OpID),
		sequenceValuesCache:   make(map[OpID][]sequenceValue),
		sequenceElementsCache: make(map[OpID][]Operation),
		mapKeyIndex:           make(map[ObjectID]map[string][]OpID),
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
			delete(s.insertOrderCache, operation.Object.OpID)
			delete(s.sequenceValuesCache, operation.Object.OpID)
			delete(s.sequenceElementsCache, operation.Object.OpID)
		}

		s.indexMapKeyOperation(operation)
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

		matches := slices.Contains(increment.Predecessors, operation.ID)

		if !matches {
			if slices.Contains(operation.Successors, increment.ID) {
				matches = true
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

// mapKeyOperationIDs returns every operation addressing a map property,
// building the property index on first use.
func (s *State) mapKeyOperationIDs(object ObjectID, property string) []OpID {
	if !s.mapKeyIndexBuilt {
		s.mapKeyIndexBuilt = true

		for _, operation := range s.operations {
			s.indexMapKeyOperation(operation)
		}
	}

	properties, ok := s.mapKeyIndex[object]
	if !ok {
		return nil
	}

	return properties[property]
}

// indexMapKeyOperation records an operation under the map property it
// addresses. It is a no-op until the index has been built, because the pending
// build will pick the operation up from the operation set.
func (s *State) indexMapKeyOperation(operation Operation) {
	if !s.mapKeyIndexBuilt || operation.Key.Property == nil {
		return
	}

	properties, ok := s.mapKeyIndex[operation.Object]
	if !ok {
		properties = make(map[string][]OpID)
		s.mapKeyIndex[operation.Object] = properties
	}

	properties[*operation.Key.Property] = append(
		properties[*operation.Key.Property],
		operation.ID,
	)
}

func (s *State) visibleMapObjectOperations(
	object ObjectID,
	property string,
) []Operation {
	operations := make([]Operation, 0)

	for _, id := range s.mapKeyOperationIDs(object, property) {
		operation, ok := s.operations[id]
		if !ok ||
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

		return
	}

	if operation.Key.Element == nil {
		delete(s.insertOrderCache, object)

		return
	}

	anchor := *operation.Key.Element

	if len(order) > 0 && order[len(order)-1] == anchor {
		s.insertOrderCache[object] = append(order, operation.ID)

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

		return
	}

	// The anchor is not present in the cached order; rebuild lazily.
	delete(s.insertOrderCache, object)
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
	}
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
	// id is the mark's begin operation, which orders precedence: a later mark
	// (an unmark, or a new value) overrides an earlier one where they overlap.
	id OpID
}

// MarkRange is one active mark over a UTF-16 range of a text object.
type MarkRange struct {
	Start uint32
	End   uint32
	Name  string
	Value *Scalar
}

// insertAnchorKey adjusts an insertion anchor so a new element lands on the
// correct side of the mark boundaries that follow it, mirroring the reference's
// insert query. Scanning forward from the anchor, an expanding mark begin and a
// non-expanding mark end each offer a position after themselves, so the new
// element joins the expanding range. Reaching the end of a mark whose begin
// offered a position withdraws that offer, because a begin/end pair with no
// visible content between them must not capture the insertion. The scan stops at
// the first visible element; tombstones are stepped over.
func (s *State) insertAnchorKey(object OpID, base Key) Key {
	order := s.insertOrder(object)

	start := 0

	if !base.IsHead {
		if base.Element == nil {
			return base
		}

		found := false

		for i, id := range order {
			if id == *base.Element {
				start = i + 1
				found = true

				break
			}
		}

		if !found {
			return base
		}
	}

	type candidate struct {
		key Key
		id  *OpID
	}

	candidates := []candidate{{key: base}}

	for i := start; i < len(order); i++ {
		operation, ok := s.operations[order[i]]
		if !ok {
			continue
		}

		if operation.Action == ActionMark {
			expand := operation.MarkExpand != nil && *operation.MarkExpand
			isEnd := operation.MarkName == nil
			withdrawn := false

			if isEnd {
				begin := OpID{Actor: operation.ID.Actor, Counter: operation.ID.Counter - 1}

				for index := range candidates {
					if candidates[index].id != nil && *candidates[index].id == begin {
						candidates = candidates[:index]
						withdrawn = true

						break
					}
				}
			}

			if !withdrawn && ((!isEnd && expand) || (isEnd && !expand)) {
				candidates = append(candidates, candidate{
					key: Key{Element: new(operation.ID)},
					id:  new(operation.ID),
				})
			}

			continue
		}

		if !s.isSuperseded(operation.ID) && len(candidates) > 0 {
			break
		}
	}

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
func (s *State) richTextMarks(object OpID, elements []Operation) []richTextMark {
	order := s.insertOrder(object)

	type openMark struct {
		start     int
		operation Operation
	}

	open := make(map[OpID]openMark)
	marks := make([]richTextMark, 0)
	elementIndex := make(map[OpID]int)
	index := 0

	closeMark := func(begin openMark, end int) {
		if end <= begin.start || begin.operation.MarkName == nil {
			return
		}

		marks = append(marks, richTextMark{
			start:  begin.start,
			end:    end,
			name:   *begin.operation.MarkName,
			value:  scalarMaterializedValue(begin.operation.Value),
			scalar: begin.operation.Value,
			id:     begin.operation.ID,
		})
	}

	for _, id := range order {
		operation, ok := s.operations[id]
		if !ok || s.isSuperseded(id) {
			continue
		}

		if operation.Action != ActionMark {
			elementIndex[id] = index
			index++

			continue
		}

		if operation.MarkName != nil {
			open[operation.ID] = openMark{start: index, operation: operation}

			continue
		}

		begin := OpID{Actor: operation.ID.Actor, Counter: operation.ID.Counter - 1}
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
		endID := OpID{Actor: opened.operation.ID.Actor, Counter: opened.operation.ID.Counter + 1}

		// The end operation exists only when the following operation is actually
		// a mark end. A begin whose end insert failed leaves that counter free
		// for a later operation (a delete, say), so checking the action avoids
		// mistaking such an operation for the missing end.
		if end, ok := s.operations[endID]; ok &&
			end.Action == ActionMark && end.MarkName == nil {
			continue
		}

		remaining = append(remaining, opened)
	}

	sort.Slice(remaining, func(i, j int) bool {
		return remaining[i].operation.ID.Compare(remaining[j].operation.ID) < 0
	})

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
	sort.SliceStable(marks, func(i, j int) bool {
		return marks[i].id.Compare(marks[j].id) < 0
	})

	_ = elements

	return marks
}

// danglingBeginStart returns the visible index a leftward-expanding dangling
// begin should start from: the document start for a head anchor, the position
// immediately after the anchor element otherwise, and the walk index as a
// fallback when the anchor is no longer visible.
func danglingBeginStart(anchor Key, elementIndex map[OpID]int, fallback int) int {
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

// markOpUTF16Range returns the literal UTF-16 range a mark operation pair spans,
// without boundary-expansion adjustment. It is used to report mark and unmark
// operations as Mark patches, matching the reference's operation-based diff.
func (s *State) markOpUTF16Range(object OpID, begin, end Operation) (uint32, uint32, bool) {
	elements := s.sequenceElements(object)

	positions := make(map[OpID]int, len(elements))
	for index, element := range elements {
		positions[element.ID] = index
	}

	beginExpand := begin.MarkExpand != nil && *begin.MarkExpand
	endExpand := end.MarkExpand != nil && *end.MarkExpand

	startIndex, startOK := s.markAnchorPosition(
		object, begin.Key, begin.ID, true, beginExpand, positions, elements, false, make(map[OpID]struct{}),
	)
	endIndex, endOK := s.markAnchorPosition(
		object, end.Key, end.ID, false, endExpand, positions, elements, false, make(map[OpID]struct{}),
	)

	if !startOK || !endOK {
		return 0, 0, false
	}

	return utf16PrefixLength(elements, startIndex), utf16PrefixLength(elements, endIndex), true
}

// utf16PrefixLength sums the UTF-16 width of the first count elements, so a mark
// anchor expressed as an element index becomes a UTF-16 position.
func utf16PrefixLength(elements []Operation, count int) uint32 {
	var position uint32

	for i := 0; i < count && i < len(elements); i++ {
		position += elementLength(elements[i])
	}

	return position
}

func (s *State) markAnchorPosition(
	object OpID,
	key Key,
	marker OpID,
	start bool,
	expand bool,
	positions map[OpID]int,
	elements []Operation,
	adjustBoundary bool,
	visited map[OpID]struct{},
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

	operation, ok := s.operations[*key.Element]
	if !ok {
		return 0, false
	}

	if operation.Action == ActionMark {
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
	key Key,
	marker OpID,
	position int,
	elements []Operation,
) int {
	for position < len(elements) {
		child, ok := s.boundaryChild(elements[position], key, make(map[OpID]struct{}))
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
	element Operation,
	boundary Key,
	visited map[OpID]struct{},
) (OpID, bool) {
	current := element

	for {
		if _, ok := visited[current.ID]; ok {
			return OpID{}, false
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
			return OpID{}, false
		}

		parent, ok := s.operations[*current.Key.Element]
		if !ok || parent.Action == ActionMark {
			return OpID{}, false
		}

		current = parent
	}
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

// maxOpForActor returns the highest operation counter authored by the actor in
// this state, or zero if the actor has no operations. It is used to decide
// whether an actor is fully covered by a set of heads when isolating writes.
func (s *State) maxOpForActor(actor ActorID) uint64 {
	var maximum uint64

	for id := range s.operations {
		if id.Actor == actor && id.Counter > maximum {
			maximum = id.Counter
		}
	}

	return maximum
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

		s.spliceInsertOrder(operation)
		s.updateSequenceValues(operation)
		s.indexMapKeyOperation(operation)
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
