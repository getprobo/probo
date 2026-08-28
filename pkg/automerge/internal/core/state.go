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
	"bytes"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

type (
	State struct {
		// columns is authoritative for committed operations, changes, and heads
		// in Engine-managed states. The maps below are a pending/applied overlay
		// while a transaction is open or columns are being reconciled. Standalone
		// State values retain the maps as their complete fallback authority.
		columns       *columnarState
		changes       map[opset.ChangeHash]*opset.Change
		actorSequence map[opset.ActorID]uint64
		operations    map[opset.OpID]opset.Operation
		operationIDs  map[opset.OpID]struct{}
		superseded    map[opset.OpID]struct{}
		heads         map[opset.ChangeHash]struct{}
		removedHeads  map[opset.ChangeHash]struct{}
		sequenceCache map[opset.OpID][]opset.Operation

		// sequenceIndexes are built lazily per object. Their immutable chunks
		// carry RGA positions, visible winners, and UTF-16 prefix widths; local
		// edits replace only touched chunks, while unsafe merged changes and
		// rollbacks discard the object entry for a full rebuild.
		sequenceIndexes map[opset.OpID]*sequenceIndex
		sequenceRescue  map[opset.OpID]bool

		// insertOrderCache holds, per sequence object, the RGA-ordered list of
		// insertion operation IDs (including tombstones and marks). The
		// order depends only on insertion operations and their anchors, so it is
		// maintained incrementally: local inserts (whose IDs are always the
		// current maximum) splice in next to their anchor, while merged changes
		// and rollbacks invalidate the entry so it is rebuilt lazily.
		insertOrderCache map[opset.OpID][]opset.OpID

		// sequenceValuesCache and sequenceElementsCache are compatibility
		// projections for callers that still need flat slices. Mutations discard
		// them while preserving the localized chunked index.
		sequenceValuesCache   map[opset.OpID][]sequenceValue
		sequenceElementsCache map[opset.OpID][]opset.Operation

		// sequenceOffsetCache holds the cumulative UTF-16 width before each
		// element of sequenceElementsCache (with a trailing total), so a text
		// index resolves to an element by binary search instead of a linear
		// walk. It is kept in step with the elements cache and guarded by length.
		sequenceOffsetCache map[opset.OpID][]uint32
		sequenceTailCache   map[opset.OpID]sequenceTail

		// insertOrderPositionCache maps each insert-order operation to its index,
		// so an insertion anchor resolves in constant time instead of scanning
		// the order. Insert order only ever grows, so a length guard is enough to
		// detect staleness.
		insertOrderPositionCache map[opset.OpID]map[opset.OpID]int

		// mapKeyIndex groups operation IDs by the map property they address so
		// reading a key does not scan the whole operation set. It is built on
		// first use and then maintained as operations are applied.
		mapKeyIndex      map[opset.ObjectID]map[string][]opset.OpID
		mapKeyIndexBuilt bool

		sequenceElementIndex map[opset.OpID][]opset.OpID
		actorChangeIndex     map[opset.ActorID]map[uint64]opset.ChangeHash
		successorIndex       map[opset.OpID][]opset.OpID

		directRemoteSequence bool
	}

	RichSpan struct {
		Type  string         `json:"type"`
		Value any            `json:"value"`
		Marks map[string]any `json:"marks,omitempty"`
	}

	sequenceValue struct {
		Element   opset.OpID
		Operation opset.Operation
	}
)

func NewState() *State {
	return &State{
		changes:                  make(map[opset.ChangeHash]*opset.Change),
		actorSequence:            make(map[opset.ActorID]uint64),
		operations:               make(map[opset.OpID]opset.Operation),
		operationIDs:             make(map[opset.OpID]struct{}),
		superseded:               make(map[opset.OpID]struct{}),
		heads:                    make(map[opset.ChangeHash]struct{}),
		removedHeads:             make(map[opset.ChangeHash]struct{}),
		sequenceCache:            make(map[opset.OpID][]opset.Operation),
		sequenceIndexes:          make(map[opset.OpID]*sequenceIndex),
		sequenceRescue:           make(map[opset.OpID]bool),
		insertOrderCache:         make(map[opset.OpID][]opset.OpID),
		sequenceValuesCache:      make(map[opset.OpID][]sequenceValue),
		sequenceElementsCache:    make(map[opset.OpID][]opset.Operation),
		sequenceOffsetCache:      make(map[opset.OpID][]uint32),
		sequenceTailCache:        make(map[opset.OpID]sequenceTail),
		insertOrderPositionCache: make(map[opset.OpID]map[opset.OpID]int),
		mapKeyIndex:              make(map[opset.ObjectID]map[string][]opset.OpID),
		sequenceElementIndex:     make(map[opset.OpID][]opset.OpID),
		actorChangeIndex:         make(map[opset.ActorID]map[uint64]opset.ChangeHash),
		successorIndex:           make(map[opset.OpID][]opset.OpID),
	}
}

func NewStateFromDocument(document *opset.Document) (*State, error) {
	return newStateFromDocument(document, true)
}

func newRescueStateFromDocument(document *opset.Document) (*State, error) {
	return newStateFromDocument(document, false)
}

func newStateFromDocument(
	document *opset.Document,
	validateMarks bool,
) (*State, error) {
	state := NewState()

	// Presize the operation and change maps so loading a large document does not
	// rehash them repeatedly as it inserts every operation.
	operationCount := 0
	for i := range document.Changes {
		operationCount += len(document.Changes[i].Operations)
	}

	state.operations = make(map[opset.OpID]opset.Operation, operationCount)
	state.changes = make(map[opset.ChangeHash]*opset.Change, len(document.Changes))
	state.mapKeyIndexBuilt = true

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Sequence > state.actorSequence[change.Actor] {
			state.actorSequence[change.Actor] = change.Sequence
		}

		if change.Hash != nil {
			state.changes[*change.Hash] = change
			state.indexActorChange(change.Actor, change.Sequence, *change.Hash)
		}

		for _, operation := range change.Operations {
			if _, exists := state.operations[operation.ID]; exists {
				return nil, fmt.Errorf("duplicate snapshot operation ID %v", operation.ID)
			}

			state.operations[operation.ID] = operation
			state.operationIDs[operation.ID] = struct{}{}
			state.observeLoadedSequenceOperation(operation)
			state.indexMapKeyOperation(operation)
			state.indexSequenceElementOperation(operation)
			state.indexSuccessors(operation)
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
				successorOperation.Action != opset.ActionIncrement ||
				!isCounterOperation(operation) {
				state.superseded[operation.ID] = struct{}{}
			}
		}
		if !operation.Object.IsRoot &&
			operation.Insert &&
			state.isSuperseded(operation.ID) {
			tail := state.sequenceTailCache[operation.Object.OpID]
			tail.safe = false
			state.sequenceTailCache[operation.Object.OpID] = tail
		}
	}

	if validateMarks {
		if err := state.validateMarkOrder(); err != nil {
			return nil, err
		}
	}
	state.finalizeSequenceTails(document.OperationOrder)

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
	dependedOn := make(map[opset.ChangeHash]struct{}, len(state.changes))

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

func (s *State) validateMarkOrder() error {
	objects := make(map[opset.OpID]struct{})

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Action == opset.ActionMark && !operation.Object.IsRoot {
			objects[operation.Object.OpID] = struct{}{}
		}
		return true
	})

	for object := range objects {
		seen := make(map[opset.OpID]struct{})

		for _, id := range s.insertOrder(object) {
			operation, ok := s.operation(id)
			if !ok || operation.Action != opset.ActionMark {
				continue
			}

			if operation.MarkName != nil {
				seen[id] = struct{}{}

				continue
			}

			if id.Counter == 0 {
				continue
			}

			beginID := opset.OpID{Actor: id.Actor, Counter: id.Counter - 1}

			begin, ok := s.operation(beginID)
			if !ok ||
				begin.Action != opset.ActionMark ||
				begin.MarkName == nil ||
				begin.Object != operation.Object {
				continue
			}

			if _, ok := seen[beginID]; !ok {
				return fmt.Errorf("invalid mark operation order: end %v precedes begin %v", id, beginID)
			}
		}
	}

	return nil
}

func (s *State) ApplyChange(change *opset.Change) error {
	if change.Hash == nil {
		return fmt.Errorf("change hash is required")
	}
	if err := validateChangeSnapshotDomain(change); err != nil {
		return err
	}

	if _, ok := s.change(*change.Hash); ok {
		return nil
	}

	for _, dependency := range change.Dependencies {
		if _, ok := s.change(dependency); !ok {
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
		if s.hasOperationID(operation.ID) {
			return fmt.Errorf("duplicate operation ID %v", operation.ID)
		}
	}

	for _, operation := range change.Operations {
		s.operations[operation.ID] = operation
		s.operationIDs[operation.ID] = struct{}{}
		if !operation.Object.IsRoot {
			delete(s.sequenceCache, operation.Object.OpID)
			delete(s.sequenceValuesCache, operation.Object.OpID)
			delete(s.sequenceElementsCache, operation.Object.OpID)
			delete(s.sequenceOffsetCache, operation.Object.OpID)
			delete(s.sequenceTailCache, operation.Object.OpID)
			if operation.Insert {
				delete(s.insertOrderCache, operation.Object.OpID)
				delete(s.insertOrderPositionCache, operation.Object.OpID)
			}
		}

		s.indexMapKeyOperation(operation)
		s.indexSequenceElementOperation(operation)
		s.indexSuccessors(operation)
		s.supersedePredecessors(operation)
		s.updateSequenceIndex(operation, false)
	}

	s.changes[*change.Hash] = change
	s.indexActorChange(change.Actor, change.Sequence, *change.Hash)

	s.actorSequence[change.Actor] = change.Sequence
	for _, dependency := range change.Dependencies {
		delete(s.heads, dependency)
		if s.columns != nil {
			s.removedHeads[dependency] = struct{}{}
		}
	}

	delete(s.removedHeads, *change.Hash)
	s.heads[*change.Hash] = struct{}{}

	return nil
}

func validateChangeSnapshotDomain(change *opset.Change) error {
	if change.Sequence > math.MaxUint32 || change.MaxOp > math.MaxUint32 {
		return fmt.Errorf("change exceeds snapshot uint32 domain")
	}
	for i, operation := range change.Operations {
		if operation.ID.Counter == 0 ||
			operation.ID.Counter > math.MaxUint32 {
			return fmt.Errorf(
				"operation %d counter exceeds snapshot uint32 domain",
				i,
			)
		}
		if !operation.Object.IsRoot &&
			operation.Object.OpID.Counter > math.MaxUint32 {
			return fmt.Errorf(
				"operation %d object exceeds snapshot uint32 domain",
				i,
			)
		}
		if operation.Key.Element != nil &&
			operation.Key.Element.Counter > math.MaxUint32 {
			return fmt.Errorf(
				"operation %d key exceeds snapshot uint32 domain",
				i,
			)
		}
		for _, predecessor := range operation.Predecessors {
			if predecessor.Counter > math.MaxUint32 {
				return fmt.Errorf(
					"operation %d predecessor exceeds snapshot uint32 domain",
					i,
				)
			}
		}
	}

	return nil
}

func (s *State) Heads() []opset.ChangeHash {
	capacity := len(s.heads)
	if s.columns != nil {
		capacity += len(s.columns.heads)
	}
	heads := make([]opset.ChangeHash, 0, capacity)
	if s.columns != nil {
		for _, head := range s.columns.heads {
			if _, removed := s.removedHeads[head]; !removed {
				heads = append(heads, head)
			}
		}
	}
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
	objectOperation, ok := s.visibleMapOperation(property, opset.ActionMakeText)
	if !ok {
		return "", fmt.Errorf("text property %q does not exist", property)
	}

	sequence := s.sequence(objectOperation.ID)

	var output strings.Builder

	for _, operation := range sequence {
		if operation.Value != nil && operation.Value.Type == opset.ScalarString {
			output.WriteString(operation.Value.String)
		}
	}

	return output.String(), nil
}

func (s *State) visibleMapOperation(property string, action opset.Action) (opset.Operation, bool) {
	return s.visibleMapObjectOperation(opset.RootObject(), property, action)
}

func (s *State) visibleMapObjectOperation(
	object opset.ObjectID,
	property string,
	action opset.Action,
) (opset.Operation, bool) {
	var (
		result opset.Operation
		found  bool
	)

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object != object ||
			operation.Key.Property == nil ||
			*operation.Key.Property != property ||
			operation.Action != action ||
			s.isSuperseded(operation.ID) {
			return true
		}

		if !found || operation.ID.Compare(result.ID) > 0 {
			result = operation
			found = true
		}
		return true
	})

	return result, found
}

func (s *State) visibleMapObjectValue(
	object opset.ObjectID,
	property string,
) (opset.Operation, bool) {
	var (
		result opset.Operation
		found  bool
	)

	for _, operation := range s.visibleMapObjectOperations(object, property) {
		if operation.Action == opset.ActionIncrement {
			continue
		}

		if !found || operation.ID.Compare(result.ID) > 0 {
			result = operation
			found = true
		}
	}

	return result, found
}

func isCounterOperation(operation opset.Operation) bool {
	return operation.Action == opset.ActionSet &&
		operation.Value != nil &&
		operation.Value.Type == opset.ScalarCounter
}

// supersedePredecessors marks the predecessors overwritten by operation. A
// regular operation supersedes all of its predecessors. An increment supersedes
// only its non-counter predecessors: incrementing a counter keeps it visible,
// but an increment that also references a conflicting non-counter value deletes
// that value, matching upstream Rust.
func (s *State) supersedePredecessors(operation opset.Operation) {
	for _, predecessor := range operation.Predecessors {
		if operation.Action == opset.ActionIncrement {
			if pred, ok := s.operation(predecessor); ok && isCounterOperation(pred) {
				continue
			}
		}

		s.superseded[predecessor] = struct{}{}
	}
}

func (s *State) scalarValue(operation opset.Operation) (opset.Scalar, bool) {
	if operation.Action != opset.ActionSet || operation.Value == nil {
		return opset.Scalar{}, false
	}

	value := *operation.Value

	value.Bytes = append([]byte(nil), operation.Value.Bytes...)
	if value.Type != opset.ScalarCounter {
		return value, true
	}

	s.eachOperation(func(increment opset.Operation) bool {
		if increment.Action != opset.ActionIncrement ||
			increment.Value == nil ||
			s.isSuperseded(increment.ID) {
			return true
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
		return true
	})

	return value, true
}

func (s *State) visibleMapOperations(property string) []opset.Operation {
	return s.visibleMapObjectOperations(opset.RootObject(), property)
}

// mapKeyOperationIDs returns every operation addressing a map property,
// building the property index on first use.
func (s *State) mapKeyOperationIDs(object opset.ObjectID, property string) []opset.OpID {
	if !s.mapKeyIndexBuilt {
		s.mapKeyIndexBuilt = true

		s.eachOperation(func(operation opset.Operation) bool {
			s.indexMapKeyOperation(operation)
			return true
		})
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
func (s *State) indexMapKeyOperation(operation opset.Operation) {
	if !s.mapKeyIndexBuilt || operation.Key.Property == nil {
		return
	}

	properties, ok := s.mapKeyIndex[operation.Object]
	if !ok {
		properties = make(map[string][]opset.OpID)
		s.mapKeyIndex[operation.Object] = properties
	}

	properties[*operation.Key.Property] = append(
		properties[*operation.Key.Property],
		operation.ID,
	)
}

func (s *State) visibleMapObjectOperations(
	object opset.ObjectID,
	property string,
) []opset.Operation {
	operations := make([]opset.Operation, 0)

	for _, id := range s.mapKeyOperationIDs(object, property) {
		operation, ok := s.operation(id)
		if !ok ||
			operation.Action == opset.ActionDelete ||
			operation.Action == opset.ActionIncrement ||
			s.isSuperseded(operation.ID) {
			continue
		}

		operations = append(operations, operation)
	}

	sort.Slice(
		operations,
		func(i, j int) bool {
			return operations[i].ID.Compare(operations[j].ID) < 0
		},
	)

	return operations
}

func (s *State) mapLength(object opset.ObjectID) uint64 {
	return uint64(len(s.mapKeys(object)))
}

func (s *State) mapKeys(object opset.ObjectID) []string {
	properties := make(map[string]struct{})

	s.eachOperation(func(operation opset.Operation) bool {
		if operation.Object == object &&
			operation.Key.Property != nil &&
			operation.Action != opset.ActionDelete &&
			!s.isSuperseded(operation.ID) {
			properties[*operation.Key.Property] = struct{}{}
		}
		return true
	})

	keys := make([]string, 0, len(properties))
	for property := range properties {
		keys = append(keys, property)
	}

	slices.Sort(keys)

	return keys
}

func (s *State) isSuperseded(id opset.OpID) bool {
	_, ok := s.superseded[id]
	return ok
}

func (s *State) maxOpGlobal() uint64 {
	var maximum uint64
	for id := range s.operationIDs {
		if id.Counter > maximum {
			maximum = id.Counter
		}
	}

	return maximum
}

func (s *State) sequenceForActor(actor opset.ActorID) uint64 {
	return s.actorSequence[actor]
}

// maxOpForActor returns the highest operation counter authored by the actor in
// this state, or zero if the actor has no operations. It is used to decide
// whether an actor is fully covered by a set of heads when isolating writes.
func (s *State) maxOpForActor(actor opset.ActorID) uint64 {
	var maximum uint64

	for id := range s.operationIDs {
		if id.Actor == actor && id.Counter > maximum {
			maximum = id.Counter
		}
	}

	return maximum
}

// hashForActorSequence returns the hash of the change authored by actor at the
// given sequence number, if it is known.
func (s *State) hashForActorSequence(
	actor opset.ActorID,
	sequence uint64,
) (opset.ChangeHash, bool) {
	sequences, ok := s.actorChangeIndex[actor]
	if !ok {
		return opset.ChangeHash{}, false
	}

	hash, ok := sequences[sequence]

	return hash, ok
}

func (s *State) applyPending(operations []opset.Operation) error {
	for _, operation := range operations {
		if s.hasOperationID(operation.ID) {
			return fmt.Errorf("duplicate pending operation ID %v", operation.ID)
		}

		s.operations[operation.ID] = operation
		s.operationIDs[operation.ID] = struct{}{}
		if !operation.Object.IsRoot {
			delete(s.sequenceCache, operation.Object.OpID)
			delete(s.sequenceTailCache, operation.Object.OpID)
		}

		s.spliceInsertOrder(operation)
		s.indexMapKeyOperation(operation)
		s.indexSequenceElementOperation(operation)
		s.indexSuccessors(operation)
		s.supersedePredecessors(operation)
	}
	s.updateSequenceIndexes(operations, true)

	return nil
}

// undoPending removes operations in reverse transaction order. Removing the
// newest operation first mirrors Rust's transaction rollback and lets each
// predecessor become visible again as soon as its last superseding operation is
// removed.
func (s *State) undoPending(operations []opset.Operation) {
	remainingSuperseders := make(map[opset.OpID]int)
	s.eachOperation(func(operation opset.Operation) bool {
		for _, predecessor := range operation.Predecessors {
			if s.operationSupersedesPredecessor(operation, predecessor) {
				remainingSuperseders[predecessor]++
			}
		}
		return true
	})

	for i := len(operations) - 1; i >= 0; i-- {
		operation := operations[i]

		for _, predecessor := range operation.Predecessors {
			if !s.operationSupersedesPredecessor(operation, predecessor) {
				continue
			}

			remainingSuperseders[predecessor]--
			if remainingSuperseders[predecessor] == 0 {
				delete(remainingSuperseders, predecessor)
				delete(s.superseded, predecessor)
			}
		}

		delete(s.operations, operation.ID)
		delete(s.operationIDs, operation.ID)
		s.removeMapKeyOperation(operation)
		s.removeSequenceElementOperation(operation)
		s.removeSuccessors(operation)
		s.invalidateObjectCaches(operation.Object)

		if operation.Action == opset.ActionMakeList ||
			operation.Action == opset.ActionMakeText {
			s.invalidateObjectCaches(opset.ObjectID{OpID: operation.ID})
		}
	}
}

func (s *State) removeMapKeyOperation(operation opset.Operation) {
	if !s.mapKeyIndexBuilt || operation.Key.Property == nil {
		return
	}

	properties, ok := s.mapKeyIndex[operation.Object]
	if !ok {
		return
	}

	property := *operation.Key.Property
	identifiers := properties[property]

	for i, identifier := range identifiers {
		if identifier == operation.ID {
			properties[property] = append(identifiers[:i], identifiers[i+1:]...)

			break
		}
	}

	if len(properties[property]) == 0 {
		delete(properties, property)
	}

	if len(properties) == 0 {
		delete(s.mapKeyIndex, operation.Object)
	}
}

func (s *State) invalidateObjectCaches(object opset.ObjectID) {
	if object.IsRoot {
		return
	}

	delete(s.sequenceCache, object.OpID)
	delete(s.sequenceIndexes, object.OpID)
	delete(s.insertOrderCache, object.OpID)
	delete(s.insertOrderPositionCache, object.OpID)
	delete(s.sequenceValuesCache, object.OpID)
	delete(s.sequenceElementsCache, object.OpID)
	delete(s.sequenceOffsetCache, object.OpID)
	delete(s.sequenceTailCache, object.OpID)
}

func (s *State) operationSupersedesPredecessor(
	operation opset.Operation,
	identifier opset.OpID,
) bool {
	if operation.Action != opset.ActionIncrement {
		return true
	}

	predecessor, ok := s.operation(identifier)

	return !ok || !isCounterOperation(predecessor)
}

func (s *State) recordAppliedChange(change *opset.Change) {
	for _, dependency := range change.Dependencies {
		delete(s.heads, dependency)
		if s.columns != nil {
			s.removedHeads[dependency] = struct{}{}
		}
	}

	delete(s.removedHeads, *change.Hash)
	s.heads[*change.Hash] = struct{}{}
	s.changes[*change.Hash] = change
	s.actorSequence[change.Actor] = change.Sequence
	s.indexActorChange(change.Actor, change.Sequence, *change.Hash)
}

func (s *State) indexActorChange(
	actor opset.ActorID,
	sequence uint64,
	hash opset.ChangeHash,
) {
	sequences := s.actorChangeIndex[actor]
	if sequences == nil {
		sequences = make(map[uint64]opset.ChangeHash)
		s.actorChangeIndex[actor] = sequences
	}
	sequences[sequence] = hash
}

func (s *State) indexSequenceElementOperation(operation opset.Operation) {
	if operation.Insert || operation.Key.Element == nil {
		return
	}

	element := *operation.Key.Element
	s.sequenceElementIndex[element] = append(
		s.sequenceElementIndex[element],
		operation.ID,
	)
}

func (s *State) removeSequenceElementOperation(operation opset.Operation) {
	if operation.Insert || operation.Key.Element == nil {
		return
	}

	element := *operation.Key.Element
	identifiers := s.sequenceElementIndex[element]
	for i, identifier := range identifiers {
		if identifier != operation.ID {
			continue
		}

		identifiers = append(identifiers[:i], identifiers[i+1:]...)
		break
	}

	if len(identifiers) == 0 {
		delete(s.sequenceElementIndex, element)
	} else {
		s.sequenceElementIndex[element] = identifiers
	}
}

func (s *State) indexSuccessors(operation opset.Operation) {
	for _, predecessor := range operation.Predecessors {
		s.successorIndex[predecessor] = append(
			s.successorIndex[predecessor],
			operation.ID,
		)
	}
}

func (s *State) removeSuccessors(operation opset.Operation) {
	for _, predecessor := range operation.Predecessors {
		successors := s.successorIndex[predecessor]
		for i, successor := range successors {
			if successor != operation.ID {
				continue
			}
			successors = append(successors[:i], successors[i+1:]...)
			break
		}
		if len(successors) == 0 {
			delete(s.successorIndex, predecessor)
		} else {
			s.successorIndex[predecessor] = successors
		}
	}
}

func (s *State) hasChange(hash opset.ChangeHash) bool {
	_, ok := s.change(hash)
	return ok
}

func (s *State) hasDependencies(change *opset.Change) bool {
	for _, dependency := range change.Dependencies {
		if !s.hasChange(dependency) {
			return false
		}
	}

	return true
}

// changesSince returns the changes reachable from the current frontier that the
// baseline heads do not already cover, in dependency order, ready to replay.
//
// The result is always a consistent prefix: a change is emitted only once every
// one of its ancestors has been emitted or is already known to the baseline, so
// a caller never receives a change whose dependency it was not also given. The
// second return reports whether that prefix is complete. It is false when some
// change in the frontier's ancestry could not be produced, either because it is
// absent from the graph or because its original bytes are unavailable.
//
// Completeness is a signal, not a gate. Sync uses it to fall back to sending a
// whole document, which reproduces changes even when the in-memory graph is
// inconsistent. Incremental reads use the prefix regardless, because returning
// every change that can be produced keeps a document usable where failing the
// whole read would wedge it: a change that cannot be emitted has no bytes to
// return anyway.
func (s *State) changesSince(heads []opset.ChangeHash) ([]*opset.Change, bool) {
	known := s.changeClosure(heads)

	const (
		visiting = iota
		reachable
		unreachable
	)

	ordered := make([]*opset.Change, 0)
	status := make(map[opset.ChangeHash]int)

	var visit func(opset.ChangeHash) bool

	visit = func(hash opset.ChangeHash) bool {
		if state, ok := status[hash]; ok {
			// A change still on the stack cannot be depended upon to be complete
			// yet, but treating the cycle edge as reachable avoids excluding the
			// whole branch over a graph that should never contain a cycle anyway.
			return state != unreachable
		}

		// The baseline closure is transitively closed, so everything below a change
		// the peer already holds is also held and need not be walked.
		if _, ok := known[hash]; ok {
			status[hash] = reachable

			return true
		}

		status[hash] = visiting

		change, ok := s.change(hash)
		if !ok {
			status[hash] = unreachable

			return false
		}

		complete := true

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				complete = false
			}
		}

		// A change is emittable only when every ancestor is, so the result stays a
		// replayable prefix, and only when its bytes exist to be returned.
		if !complete || len(change.Raw) == 0 {
			status[hash] = unreachable

			return false
		}

		ordered = append(ordered, change)
		status[hash] = reachable

		return true
	}

	complete := true

	for _, head := range s.Heads() {
		// A frontier head whose change is not retrievable contributes nothing.
		if _, ok := s.change(head); !ok {
			complete = false

			continue
		}

		if !visit(head) {
			complete = false
		}
	}

	return ordered, complete
}

func (s *State) allChanges() ([]*opset.Change, bool) {
	changeCount := s.changeCount()
	if s.columns != nil && len(s.columns.changes) == changeCount {
		ordered := make([]*opset.Change, 0, changeCount)
		for i := range s.columns.changes {
			change := &s.columns.changes[i]
			if change.Hash == nil {
				return nil, false
			}
			if _, retained := s.changes[*change.Hash]; !retained {
				return nil, false
			}
			ordered = append(ordered, change)
		}
		return ordered, true
	}
	ordered := make([]*opset.Change, 0, changeCount)
	visited := make(map[opset.ChangeHash]struct{}, changeCount)

	var visit func(opset.ChangeHash) bool

	visit = func(hash opset.ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := s.change(hash)
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
		if _, ok := s.change(head); !ok {
			continue
		}

		if !visit(head) {
			return nil, false
		}
	}

	return ordered, len(visited) == changeCount
}

func (s *State) at(heads []opset.ChangeHash) (*State, bool) {
	target := NewState()
	visited := make(map[opset.ChangeHash]struct{})

	var visit func(opset.ChangeHash) bool

	visit = func(hash opset.ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := s.change(hash)
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
func (s *State) changeClosure(heads []opset.ChangeHash) map[opset.ChangeHash]struct{} {
	closure := make(map[opset.ChangeHash]struct{})

	pending := append([]opset.ChangeHash(nil), heads...)

	for len(pending) > 0 {
		index := len(pending) - 1
		hash := pending[index]
		pending = pending[:index]

		if _, ok := closure[hash]; ok {
			continue
		}

		change, ok := s.change(hash)
		if !ok {
			continue
		}

		closure[hash] = struct{}{}

		pending = append(pending, change.Dependencies...)
	}

	return closure
}
