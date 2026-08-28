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
	"errors"
	"fmt"
	"slices"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

var errDirectColumnsUnsupported = errors.New(
	"canonical columns do not support direct mutation",
)

const directTextWireRunGap = 32

type (
	// columnMutationBatch is a transaction prepared from the pending semantic
	// overlay. It owns every mutable value needed to advance canonical columns,
	// so applying it cannot partially publish a new history.
	columnMutationBatch struct {
		appendedChanges  []opset.Change
		operationSplices []storage.SnapshotOperationSplice
		successorPatches []columnSuccessorPatch
		actors           []opset.ActorID
		heads            []opset.ChangeHash
		headIndexes      []uint64
		dependencyRows   map[opset.ChangeHash]uint64
		touchedObjects   map[opset.ObjectID][]opset.Operation
		reuseOperations  bool
	}

	columnSuccessorPatch struct {
		identifier opset.OpID
		successors []opset.OpID
	}
)

func newColumnMutationBatch(
	columns *columnarState,
	state *State,
	changes []*opset.Change,
	allowDivergedSequence bool,
) (*columnMutationBatch, error) {
	if columns == nil || columns.snapshot == nil || columns.columnsDirty {
		return nil, errDirectColumnsUnsupported
	}

	batch := &columnMutationBatch{
		appendedChanges: make([]opset.Change, 0, len(changes)),
		dependencyRows:  make(map[opset.ChangeHash]uint64),
		touchedObjects:  make(map[opset.ObjectID][]opset.Operation),
	}

	nextRows := make(map[opset.ChangeHash]int, len(columns.changeRows)+len(changes))
	for hash, row := range columns.changeRows {
		nextRows[hash] = row
	}

	for _, change := range changes {
		if change.Hash == nil {
			return nil, fmt.Errorf("direct column change has no hash")
		}

		if _, exists := nextRows[*change.Hash]; exists {
			return nil, fmt.Errorf("direct column change already exists")
		}

		for _, dependency := range change.Dependencies {
			row, ok := nextRows[dependency]
			if !ok {
				return nil, fmt.Errorf(
					"direct column dependency %s is absent",
					dependency,
				)
			}

			batch.dependencyRows[dependency] = uint64(row)
		}

		nextRows[*change.Hash] = len(columns.changes) + len(batch.appendedChanges)
		batch.appendedChanges = append(
			batch.appendedChanges,
			cloneChange(*change),
		)
	}

	touched := make(map[opset.ObjectID]struct{})

	for _, change := range changes {
		for _, operation := range change.Operations {
			touched[operation.Object] = struct{}{}
		}
	}

	objects := make([]opset.ObjectID, 0, len(touched))
	for object := range touched {
		objects = append(objects, object)
	}

	slices.SortFunc(objects, compareDocumentObjects)

	tailPlanned := false

	if len(objects) == 1 {
		var (
			splice   storage.SnapshotOperationSplice
			boundary []opset.Operation
		)

		splice, boundary, tailPlanned = directTailInsertion(
			columns,
			state,
			changes,
		)
		if tailPlanned {
			batch.operationSplices = append(batch.operationSplices, splice)
			batch.touchedObjects[objects[0]] = boundary
			// A strict end append cannot overwrite rows visible through the
			// current column view. Reuse its spare capacity when the view is
			// unshared; apply still copies forked snapshots below.
			batch.reuseOperations = true
		}

		if !tailPlanned && allowDivergedSequence {
			splice, boundary, tailPlanned = directDivergedSequenceInsertion(
				columns,
				state,
				changes,
			)
			if tailPlanned {
				batch.operationSplices = append(batch.operationSplices, splice)
				batch.touchedObjects[objects[0]] = boundary
			}
		}
	}

	textPlanned := false

	if len(objects) == 1 && !tailPlanned {
		var (
			splices  []storage.SnapshotOperationSplice
			boundary []opset.Operation
		)

		splices, boundary, textPlanned = directTextOverlay(
			columns,
			state,
			changes,
			objects[0],
		)
		if textPlanned {
			batch.operationSplices = append(batch.operationSplices, splices...)
			batch.touchedObjects[objects[0]] = boundary
		}
	}

	for _, object := range objects {
		if textPlanned || tailPlanned {
			break
		}

		order, ok := state.indexedObjectOrder(object)
		if !ok {
			return nil, fmt.Errorf("cannot order directly mutated object %v", object)
		}

		replacement := make([]opset.Operation, len(order))
		for i, identifier := range order {
			operation, ok := state.operation(identifier)
			if !ok {
				return nil, fmt.Errorf("direct operation %v is absent", identifier)
			}

			operation.Successors = append(
				[]opset.OpID(nil),
				state.successorIndex[identifier]...,
			)
			slices.SortFunc(operation.Successors, compareOperationIDs)
			replacement[i] = cloneOperation(operation)
		}

		batch.touchedObjects[object] = replacement

		current := []opset.Operation(nil)

		index := directObjectInsertionRow(columns, object)
		if span, exists := columns.objectSpans[object]; exists {
			index = span.start
			current = columns.operations[span.start:span.end]
		}

		prefix, suffix := commonOperationEdges(current, replacement)

		splice := storage.SnapshotOperationSplice{
			Index:       index + prefix,
			DeleteCount: len(current) - prefix - suffix,
			Operations:  replacement[prefix : len(replacement)-suffix],
		}
		if splice.DeleteCount > 0 || len(splice.Operations) > 0 {
			batch.operationSplices = append(batch.operationSplices, splice)
		}

		for i := prefix; i < len(current)-suffix; i++ {
			if i < len(replacement) && current[i].ID == replacement[i].ID {
				batch.successorPatches = append(
					batch.successorPatches,
					columnSuccessorPatch{
						identifier: replacement[i].ID,
						successors: append(
							[]opset.OpID(nil),
							replacement[i].Successors...,
						),
					},
				)
			}
		}
	}

	slices.Reverse(batch.operationSplices)
	batch.operationSplices = coalesceDirectInsertions(batch.operationSplices)

	batch.actors = directActorTable(columns.actors, changes)
	batch.heads = directBatchHeads(columns.heads, changes)

	batch.headIndexes = make([]uint64, len(batch.heads))
	for i, head := range batch.heads {
		row, ok := nextRows[head]
		if !ok {
			return nil, fmt.Errorf("direct column head %s is absent", head)
		}

		batch.headIndexes[i] = uint64(row)
	}

	return batch, nil
}

type directTextRowMutation struct {
	insertions  []opset.Operation
	replacement *opset.Operation
}

// directTextOverlay turns a transaction's final chunked RGA order into one
// ordered set of wire-row insertions and successor replacements. Existing text
// rows are never materialized or copied into a replacement object span.
func directTextOverlay(
	columns *columnarState,
	state *State,
	changes []*opset.Change,
	object opset.ObjectID,
) ([]storage.SnapshotOperationSplice, []opset.Operation, bool) {
	if object.IsRoot {
		return nil, nil, false
	}

	creator, ok := state.operation(object.OpID)
	if !ok || creator.Action != opset.ActionMakeText {
		return nil, nil, false
	}

	span, ok := columns.objectSpans[object]
	if !ok {
		return nil, nil, false
	}

	inserted := make(map[opset.OpID]struct{})
	changedSuccessors := make(map[opset.OpID]struct{})

	for _, change := range changes {
		for _, operation := range change.Operations {
			if operation.Object != object {
				return nil, nil, false
			}

			if operation.Insert {
				inserted[operation.ID] = struct{}{}
				continue
			}

			if operation.Action != opset.ActionDelete {
				return nil, nil, false
			}

			for _, predecessor := range operation.Predecessors {
				changedSuccessors[predecessor] = struct{}{}
			}
		}
	}

	if len(inserted) == 0 && len(changedSuccessors) == 0 {
		return nil, nil, false
	}

	index := state.sequenceIndex(object.OpID)
	mutations := make(map[int]*directTextRowMutation)
	pending := make([]opset.Operation, 0)

	var (
		first    opset.Operation
		last     opset.Operation
		hasEntry bool
	)

	flush := func(row int) {
		if len(pending) == 0 {
			return
		}

		mutation := mutations[row]
		if mutation == nil {
			mutation = &directTextRowMutation{}
			mutations[row] = mutation
		}

		mutation.insertions = append(mutation.insertions, pending...)
		pending = nil
	}

	for _, chunk := range index.chunks {
		for _, entry := range chunk.entries {
			operation, exists := state.operation(entry.insertion)
			if !exists {
				return nil, nil, false
			}

			if !hasEntry {
				first = operation
				hasEntry = true
			}

			last = operation
			if _, exists := inserted[entry.insertion]; exists {
				operation.Successors = append(
					[]opset.OpID(nil),
					state.successorIndex[operation.ID]...,
				)
				slices.SortFunc(operation.Successors, compareOperationIDs)
				pending = append(pending, cloneOperation(operation))

				continue
			}

			row, exists := columns.operationRows.lookup(entry.insertion)
			if !exists || row < span.start || row >= span.end {
				return nil, nil, false
			}

			flush(row)

			if _, changed := changedSuccessors[entry.insertion]; changed {
				operation.Successors = append(
					[]opset.OpID(nil),
					state.successorIndex[operation.ID]...,
				)
				slices.SortFunc(operation.Successors, compareOperationIDs)
				replacement := cloneOperation(operation)

				mutation := mutations[row]
				if mutation == nil {
					mutation = &directTextRowMutation{}
					mutations[row] = mutation
				}

				mutation.replacement = &replacement
			}
		}
	}

	flush(span.end)

	if !hasEntry {
		return nil, nil, false
	}

	boundary := []opset.Operation{first, last}

	rows := make([]int, 0, len(mutations))
	for row := range mutations {
		rows = append(rows, row)
	}

	slices.Sort(rows)

	splices := make([]storage.SnapshotOperationSplice, 0, len(rows))
	for first := 0; first < len(rows); {
		last := first
		for last+1 < len(rows) &&
			rows[last+1]-rows[last] <= directTextWireRunGap {
			last++
		}

		start := rows[first]
		cursor := start
		operations := make([]opset.Operation, 0)

		for _, row := range rows[first : last+1] {
			for cursor < row {
				operations = append(operations, columns.operations[cursor])
				cursor++
			}

			mutation := mutations[row]

			operations = append(operations, mutation.insertions...)
			if mutation.replacement != nil {
				operations = append(operations, *mutation.replacement)
				cursor++
			}
		}

		splices = append(splices, storage.SnapshotOperationSplice{
			Index:       start,
			DeleteCount: cursor - start,
			Operations:  operations,
		})
		first = last + 1
	}

	return splices, boundary, true
}

func directTailInsertion(
	columns *columnarState,
	state *State,
	changes []*opset.Change,
) (storage.SnapshotOperationSplice, []opset.Operation, bool) {
	if len(columns.operations) == 0 {
		return storage.SnapshotOperationSplice{}, nil, false
	}

	last := columns.operations[len(columns.operations)-1]
	anchor := last.ID
	object := last.Object
	added := make([]opset.Operation, 0)

	for _, change := range changes {
		for _, changed := range change.Operations {
			if changed.Action == opset.ActionDelete ||
				!changed.Insert ||
				changed.Key.Element == nil ||
				*changed.Key.Element != anchor ||
				changed.Object != object ||
				len(changed.Predecessors) != 0 {
				return storage.SnapshotOperationSplice{}, nil, false
			}

			operation, ok := state.operation(changed.ID)
			if !ok {
				operation = changed
			}

			operation.Successors = append(
				[]opset.OpID(nil),
				state.successorIndex[changed.ID]...,
			)
			slices.SortFunc(operation.Successors, compareOperationIDs)
			added = append(added, cloneOperation(operation))
			anchor = changed.ID
		}
	}

	if len(added) == 0 {
		return storage.SnapshotOperationSplice{}, nil, false
	}

	span, ok := columns.objectSpans[object]
	if !ok {
		return storage.SnapshotOperationSplice{}, nil, false
	}

	boundary := []opset.Operation{
		columns.operations[span.start],
		added[len(added)-1],
	}

	return storage.SnapshotOperationSplice{
		Index:      len(columns.operations),
		Operations: added,
	}, boundary, true
}

// directDivergedSequenceInsertion inserts one incoming RGA branch without
// materializing the containing object's order. Starting at the retained anchor
// row, it walks only sibling subtrees that can sort before the new branch.
func directDivergedSequenceInsertion(
	columns *columnarState,
	state *State,
	changes []*opset.Change,
) (storage.SnapshotOperationSplice, []opset.Operation, bool) {
	var (
		added  []opset.Operation
		first  opset.Operation
		object opset.ObjectID
	)

	for _, change := range changes {
		for _, changed := range change.Operations {
			if len(added) == 0 {
				first = changed
				object = changed.Object
			}

			if changed.Action == opset.ActionDelete ||
				!changed.Insert ||
				changed.Object != object ||
				len(changed.Predecessors) != 0 {
				return storage.SnapshotOperationSplice{}, nil, false
			}

			if len(added) == 0 {
				if !changed.Key.IsHead && changed.Key.Element == nil {
					return storage.SnapshotOperationSplice{}, nil, false
				}
			} else if changed.Key.Element == nil ||
				*changed.Key.Element != added[len(added)-1].ID {
				return storage.SnapshotOperationSplice{}, nil, false
			}

			operation, ok := state.operation(changed.ID)
			if !ok {
				operation = changed
			}

			operation.Successors = append(
				[]opset.OpID(nil),
				state.successorIndex[changed.ID]...,
			)
			slices.SortFunc(operation.Successors, compareOperationIDs)
			added = append(added, cloneOperation(operation))
		}
	}

	if len(added) == 0 {
		return storage.SnapshotOperationSplice{}, nil, false
	}

	span, ok := columns.objectSpans[object]
	if !ok || span.start == span.end {
		return storage.SnapshotOperationSplice{}, nil, false
	}

	index := span.start

	if !first.Key.IsHead {
		anchorRow, exists := columns.operationRows.lookup(*first.Key.Element)
		if !exists || anchorRow < span.start || anchorRow >= span.end {
			return storage.SnapshotOperationSplice{}, nil, false
		}

		index = anchorRow + 1
	}

	for index < span.end {
		runtimeMetrics.directPlanningRows.Add(1)

		candidate := columns.operations[index]
		if !candidate.Insert {
			index++
			continue
		}

		root, descendant := directSequenceBranchRoot(state, candidate, first)
		if !descendant || root.Compare(first.ID) < 0 {
			break
		}

		index++
	}

	last := columns.operations[span.end-1]
	if index == span.end {
		last = added[len(added)-1]
	}

	return storage.SnapshotOperationSplice{
			Index:      index,
			Operations: added,
		}, []opset.Operation{
			columns.operations[span.start],
			last,
		}, true
}

func directSequenceBranchRoot(
	state *State,
	candidate opset.Operation,
	incoming opset.Operation,
) (opset.OpID, bool) {
	root := candidate
	for {
		if incoming.Key.IsHead {
			if root.Key.IsHead {
				return root.ID, true
			}
		} else if root.Key.Element != nil &&
			*root.Key.Element == *incoming.Key.Element {
			return root.ID, true
		}

		if root.Key.Element == nil {
			return opset.OpID{}, false
		}

		parent, ok := state.operation(*root.Key.Element)
		if !ok || !parent.Insert || parent.Object != incoming.Object {
			return opset.OpID{}, false
		}

		root = parent
	}
}

func coalesceDirectInsertions(
	splices []storage.SnapshotOperationSplice,
) []storage.SnapshotOperationSplice {
	if len(splices) < 2 {
		return splices
	}

	coalesced := make([]storage.SnapshotOperationSplice, 0, len(splices))
	for _, splice := range splices {
		last := len(coalesced) - 1
		if last >= 0 &&
			coalesced[last].Index == splice.Index &&
			coalesced[last].DeleteCount == 0 &&
			splice.DeleteCount == 0 {
			operations := make(
				[]opset.Operation,
				0,
				len(splice.Operations)+len(coalesced[last].Operations),
			)
			operations = append(operations, splice.Operations...)
			operations = append(operations, coalesced[last].Operations...)
			coalesced[last].Operations = operations

			continue
		}

		coalesced = append(coalesced, splice)
	}

	return coalesced
}

func compareDocumentObjects(left, right opset.ObjectID) int {
	if left.IsRoot {
		if right.IsRoot {
			return 0
		}

		return -1
	}

	if right.IsRoot {
		return 1
	}

	return left.OpID.Compare(right.OpID)
}

func compareOperationIDs(left, right opset.OpID) int {
	return left.Compare(right)
}

func directObjectInsertionRow(
	columns *columnarState,
	object opset.ObjectID,
) int {
	if object.IsRoot {
		return 0
	}

	for row, operation := range columns.operations {
		if operation.Object.IsRoot {
			continue
		}

		if operation.Object.OpID.Compare(object.OpID) > 0 {
			return row
		}
	}

	return len(columns.operations)
}

func commonOperationEdges(
	current []opset.Operation,
	replacement []opset.Operation,
) (int, int) {
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

	return prefix, suffix
}

func directActorTable(
	current []opset.ActorID,
	changes []*opset.Change,
) []opset.ActorID {
	seen := make(map[opset.ActorID]struct{}, len(current)+1)
	for _, actor := range current {
		seen[actor] = struct{}{}
	}

	add := func(actor opset.ActorID) {
		if actor != "" {
			seen[actor] = struct{}{}
		}
	}
	for _, change := range changes {
		add(change.Actor)

		for _, operation := range change.Operations {
			add(operation.ID.Actor)

			if !operation.Object.IsRoot {
				add(operation.Object.OpID.Actor)
			}

			if operation.Key.Element != nil {
				add(operation.Key.Element.Actor)
			}

			for _, successor := range operation.Successors {
				add(successor.Actor)
			}
		}
	}

	actors := make([]opset.ActorID, 0, len(seen))
	for actor := range seen {
		actors = append(actors, actor)
	}

	slices.SortFunc(
		actors,
		func(left, right opset.ActorID) int {
			return left.Compare(right)
		},
	)

	return actors
}

func directBatchHeads(
	current []opset.ChangeHash,
	changes []*opset.Change,
) []opset.ChangeHash {
	heads := make(map[opset.ChangeHash]struct{}, len(current)+len(changes))
	for _, head := range current {
		heads[head] = struct{}{}
	}

	for _, change := range changes {
		for _, dependency := range change.Dependencies {
			delete(heads, dependency)
		}

		heads[*change.Hash] = struct{}{}
	}

	ordered := make([]opset.ChangeHash, 0, len(heads))
	for head := range heads {
		ordered = append(ordered, head)
	}

	slices.SortFunc(
		ordered,
		func(left, right opset.ChangeHash) int {
			return bytes.Compare(left[:], right[:])
		},
	)

	return ordered
}

func (b *columnMutationBatch) apply(columns *columnarState) (*columnarState, error) {
	next := columns.clone()
	snapshot := columns.snapshot.Clone()

	insertedChanges := make([]*opset.Change, len(b.appendedChanges))
	for i := range b.appendedChanges {
		insertedChanges[i] = &b.appendedChanges[i]
	}

	err := snapshot.Splice(
		storage.SnapshotSplice{
			Actors:            b.actors,
			Heads:             b.heads,
			HeadIndexes:       b.headIndexes,
			DependencyIndexes: b.dependencyRows,
			ChangeIndex:       len(columns.changes),
			Changes:           insertedChanges,
			OperationSplices:  b.operationSplices,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot apply direct snapshot splice: %w", err)
	}

	operationDelta := 0
	for _, splice := range b.operationSplices {
		operationDelta += len(splice.Operations) - splice.DeleteCount
	}

	operations := columns.operations
	if !b.reuseOperations ||
		columns.shared ||
		cap(operations) < len(operations)+operationDelta {
		runtimeMetrics.directOperationCopies.Add(uint64(len(operations)))
		operations = append([]opset.Operation(nil), operations...)
	}

	rows := columns.operationRows

	for _, splice := range b.operationSplices {
		inserted := make([]opset.Operation, len(splice.Operations))
		for i := range splice.Operations {
			inserted[i] = cloneOperation(splice.Operations[i])
		}

		operations = slices.Replace(
			operations,
			splice.Index,
			splice.Index+splice.DeleteCount,
			inserted...,
		)
		rows = rows.withSplice(inserted, splice.Index, splice.DeleteCount)
	}

	changes := append([]opset.Change(nil), columns.changes...)

	changeRows := make(map[opset.ChangeHash]int, len(columns.changeRows)+1)
	for hash, row := range columns.changeRows {
		changeRows[hash] = row
	}

	for i := range b.appendedChanges {
		change := cloneChange(b.appendedChanges[i])
		changeRows[*change.Hash] = len(changes)
		changes = append(changes, change)
	}

	next.actors = append([]opset.ActorID(nil), b.actors...)
	next.changes = changes
	next.changeRows = changeRows

	next.heads = append([]opset.ChangeHash(nil), b.heads...)
	next.operations = operations
	next.operationRows = rows
	next.objectSpans = updateDirectObjectSpans(columns, rows, b.touchedObjects)
	next.columnsDirty = false
	next.snapshot = snapshot
	next.shared = false

	runtimeMetrics.directColumnBatches.Add(1)

	return next, nil
}

func updateDirectObjectSpans(
	columns *columnarState,
	rows *operationRowIndex,
	touched map[opset.ObjectID][]opset.Operation,
) map[opset.ObjectID]wireObjectSpan {
	spans := make(map[opset.ObjectID]wireObjectSpan, len(columns.objectSpans))
	for object, span := range columns.objectSpans {
		if _, changed := touched[object]; changed || span.start == span.end {
			continue
		}

		first, firstOK := rows.lookup(columns.operations[span.start].ID)

		last, lastOK := rows.lookup(columns.operations[span.end-1].ID)
		if firstOK && lastOK {
			spans[object] = wireObjectSpan{start: first, end: last + 1}
		}
	}

	for object, operations := range touched {
		if len(operations) == 0 {
			continue
		}

		first, firstOK := rows.lookup(operations[0].ID)

		last, lastOK := rows.lookup(operations[len(operations)-1].ID)
		if firstOK && lastOK {
			spans[object] = wireObjectSpan{start: first, end: last + 1}
		}
	}

	return spans
}

func (s *State) promoteDirectCommit(
	change *opset.Change,
	columns *columnarState,
) {
	for _, operation := range change.Operations {
		delete(s.operations, operation.ID)
	}

	metadata := *change
	metadata.Operations = nil
	metadata.Raw = nil
	s.changes[*change.Hash] = &metadata
	s.columns = columns
	clear(s.heads)
	clear(s.removedHeads)
}
