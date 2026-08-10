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

	consistent := true

	for _, head := range document.Heads {
		if _, ok := state.changes[head]; !ok {
			consistent = false

			break
		}
	}

	if consistent {
		for _, head := range document.Heads {
			state.heads[head] = struct{}{}
		}

		return state, nil
	}

	// The recorded frontier references a change the document does not carry, so
	// it cannot be trusted. Rebuild the frontier from the change graph instead:
	// a present change is a head when no other present change depends on it. This
	// keeps Heads() consistent with changes so incremental reads never break.
	dependedOn := make(map[ChangeHash]struct{}, len(state.changes))

	for _, change := range state.changes {
		for _, dependency := range change.Dependencies {
			dependedOn[dependency] = struct{}{}
		}
	}

	for hash := range state.changes {
		if _, ok := dependedOn[hash]; !ok {
			state.heads[hash] = struct{}{}
		}
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
	known := s.changeClosure(heads)

	ordered := make([]*Change, 0)
	visited := make(map[ChangeHash]struct{})

	var visit func(ChangeHash) bool

	visit = func(hash ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		visited[hash] = struct{}{}

		// The baseline closure is transitively closed, so everything below a change
		// the peer already holds is also held. Descending would add nothing and
		// would fail the whole read if any ancestor were ever unavailable.
		if _, ok := known[hash]; ok {
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

		if len(change.Raw) == 0 {
			return false
		}

		ordered = append(ordered, change)

		return true
	}

	for _, head := range s.Heads() {
		// A frontier head whose change is not retrievable contributes nothing and
		// must not abort the incremental computation for the whole document.
		if _, ok := s.changes[head]; !ok {
			continue
		}

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
		if _, ok := s.changes[head]; !ok {
			continue
		}

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

// changeClosure returns every change reachable from the given baseline heads.
// A head whose change is not present is skipped rather than failing: it excludes
// nothing from the result, so an incremental computation over-approximates (it
// may resend changes the peer already has, which is safe) instead of aborting.
// This keeps sync and persistence working even when a frontier references a
// change that is no longer retrievable, for example after a merge that rebuilt
// the change graph.
func (s *State) changeClosure(heads []ChangeHash) map[ChangeHash]struct{} {
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
			continue
		}

		closure[hash] = struct{}{}

		pending = append(pending, change.Dependencies...)
	}

	return closure
}
