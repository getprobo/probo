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
	"fmt"
	"reflect"
	"sort"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

// columnarState is the engine's canonical wire-oriented operation and change
// view. State's maps remain query indexes during migration; they do not own the
// snapshot order used for export.
type columnarState struct {
	actors        []opset.ActorID
	changes       []opset.Change
	changeRows    map[opset.ChangeHash]int
	heads         []opset.ChangeHash
	operations    []opset.Operation
	operationRows *operationRowIndex
	objectSpans   map[opset.ObjectID]wireObjectSpan

	changeColumns    []opset.RawColumn
	operationColumns []opset.RawColumn
	columnsDirty     bool
	snapshot         *storage.SnapshotColumns
	shared           bool

	globalOrderFallbacks uint64
	snapshotReplacements uint64
}

type operationRowIndex struct {
	parent      *operationRowIndex
	rows        map[opset.OpID]int
	spliceStart int
	spliceEnd   int
	spliceDelta int
	hasSplice   bool
}

func newOperationRowIndex(capacity int) *operationRowIndex {
	return &operationRowIndex{rows: make(map[opset.OpID]int, capacity)}
}

func (i *operationRowIndex) with(
	operations []opset.Operation,
	offset int,
) *operationRowIndex {
	rows := make(map[opset.OpID]int, len(operations))
	for index, operation := range operations {
		rows[operation.ID] = offset + index
	}

	return &operationRowIndex{parent: i, rows: rows}
}

func (i *operationRowIndex) withSplice(
	operations []opset.Operation,
	index int,
	deleteCount int,
) *operationRowIndex {
	rows := make(map[opset.OpID]int, len(operations))
	for offset, operation := range operations {
		rows[operation.ID] = index + offset
	}

	return &operationRowIndex{
		parent:      i,
		rows:        rows,
		spliceStart: index,
		spliceEnd:   index + deleteCount,
		spliceDelta: len(operations) - deleteCount,
		hasSplice:   true,
	}
}

func (c *columnarState) clone() *columnarState {
	cloned := *c
	return &cloned
}

func (c *columnarState) prepareMutationCapacity() {
	if c == nil || c.shared || cap(c.operations)-len(c.operations) >= 64 {
		return
	}

	operations := make([]opset.Operation, len(c.operations), len(c.operations)+64)
	copy(operations, c.operations)
	c.operations = operations
}

func newColumnarState(document *opset.Document) (*columnarState, error) {
	if decoded, ok := document.Canonical.(*storage.DecodedDocument); ok &&
		decoded.Snapshot != nil {
		columns := &columnarState{
			actors:           document.Actors,
			changes:          document.Changes,
			changeRows:       make(map[opset.ChangeHash]int, len(document.Changes)),
			heads:            document.Heads,
			operations:       decoded.Operations,
			operationRows:    newOperationRowIndex(len(decoded.Operations)),
			changeColumns:    document.ChangeColumns,
			operationColumns: document.OperationColumns,
			snapshot:         decoded.Snapshot,
		}
		for i := range columns.changes {
			if columns.changes[i].Hash != nil {
				columns.changeRows[*columns.changes[i].Hash] = i
			}
		}

		for i := range columns.operations {
			identifier := columns.operations[i].ID
			if _, exists := columns.operationRows.rows[identifier]; exists {
				return nil, fmt.Errorf("duplicate snapshot operation ID %v", identifier)
			}

			columns.operationRows.rows[identifier] = i
		}

		columns.rebuildWireObjectSpans()

		return columns, nil
	}

	columns := &columnarState{
		actors:           append([]opset.ActorID(nil), document.Actors...),
		changes:          cloneChanges(document.Changes),
		changeRows:       make(map[opset.ChangeHash]int, len(document.Changes)),
		heads:            append([]opset.ChangeHash(nil), document.Heads...),
		operationRows:    newOperationRowIndex(len(document.OperationOrder)),
		changeColumns:    cloneRawColumns(document.ChangeColumns),
		operationColumns: cloneRawColumns(document.OperationColumns),
	}

	byID := make(map[opset.OpID]opset.Operation)

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash != nil {
			columns.changeRows[*change.Hash] = i
		}

		for _, operation := range change.Operations {
			if operation.Action != opset.ActionDelete {
				byID[operation.ID] = operation
			}
		}
	}

	if len(document.OperationOrder) == 0 && len(byID) > 0 {
		columns.columnsDirty = true
		return columns, nil
	}

	operationCapacity := len(document.OperationOrder)
	if operationCapacity > 0 {
		operationCapacity += 64
	}

	columns.operations = make([]opset.Operation, 0, operationCapacity)

	for _, identifier := range document.OperationOrder {
		operation, ok := byID[identifier]
		if !ok {
			return nil, fmt.Errorf(
				"canonical operation %v is absent from its change",
				identifier,
			)
		}

		columns.operationRows.rows[identifier] = len(columns.operations)
		columns.operations = append(
			columns.operations,
			cloneOperation(operation),
		)
	}

	snapshot, err := storage.NewSnapshotColumns(document, columns.operations)
	if err != nil {
		return nil, fmt.Errorf("cannot build canonical snapshot columns: %w", err)
	}

	columns.snapshot = snapshot
	columns.rebuildWireObjectSpans()

	return columns, nil
}

func newColumnarStateFromState(state *State) (*columnarState, error) {
	columns := &columnarState{
		changeRows:    make(map[opset.ChangeHash]int),
		operationRows: newOperationRowIndex(0),
		columnsDirty:  true,
	}
	if err := columns.reconcile(state); err != nil {
		return nil, err
	}

	return columns, nil
}

// stateFromCanonicalColumns rebuilds only the semantic query indexes from the
// canonical rows, then drops duplicate committed operation/change maps.
func stateFromCanonicalColumns(columns *columnarState) (*State, error) {
	state := NewState()
	state.columns = columns
	state.mapKeyIndexBuilt = true
	state.changes = make(
		map[opset.ChangeHash]*opset.Change,
		len(columns.changes),
	)

	operationCount := 0
	for i := range columns.changes {
		operationCount += len(columns.changes[i].Operations)
	}

	state.operationIDs = make(map[opset.OpID]struct{}, operationCount)

	for i := range columns.changes {
		change := &columns.changes[i]
		if change.Sequence > state.actorSequence[change.Actor] {
			state.actorSequence[change.Actor] = change.Sequence
		}

		if change.Hash != nil {
			metadata := *change
			metadata.Operations = nil
			metadata.Raw = nil
			state.changes[*change.Hash] = &metadata
			state.indexActorChange(change.Actor, change.Sequence, *change.Hash)
		}

		for _, changeOperation := range change.Operations {
			if _, exists := state.operationIDs[changeOperation.ID]; exists {
				return nil, fmt.Errorf(
					"duplicate snapshot operation ID %v",
					changeOperation.ID,
				)
			}

			operation := changeOperation
			if canonical, ok := columns.operation(operation.ID); ok {
				predecessors := operation.Predecessors
				operation = canonical
				operation.Predecessors = predecessors
			} else {
				state.operations[operation.ID] = operation
			}

			state.operationIDs[operation.ID] = struct{}{}
			state.observeLoadedSequenceOperation(operation)
			state.indexMapKeyOperation(operation)
			state.indexSequenceElementOperation(operation)
			state.indexSuccessors(operation)

			for _, successor := range operation.Successors {
				if successor.Counter == 0 {
					return nil, fmt.Errorf(
						"invalid zero successor for operation %v",
						operation.ID,
					)
				}
			}
		}
	}

	for i := range columns.changes {
		for _, identifier := range columns.changes[i].Operations {
			operation, ok := state.operation(identifier.ID)
			if !ok {
				continue
			}

			state.supersedePredecessors(operation)

			for _, successor := range operation.Successors {
				successorOperation, ok := state.operation(successor)
				if !ok ||
					successorOperation.Action != opset.ActionIncrement ||
					!isCounterOperation(operation) {
					state.superseded[operation.ID] = struct{}{}
				}
			}
		}
	}

	if err := state.validateMarkOrder(); err != nil {
		return nil, err
	}

	order := make([]opset.OpID, len(columns.operations))
	for i, operation := range columns.operations {
		order[i] = operation.ID
	}

	state.finalizeSequenceTails(order)

	return state, nil
}

func (b *Engine) reconcileColumns() error {
	state := b.state
	if b.isolationActive && b.fullState != nil {
		state = b.fullState
	}

	next := b.columns.clone()
	if err := next.reconcile(state); err != nil {
		return err
	}

	b.columns = next
	state.attachCanonical(next)

	return nil
}

// reconcile transactionally advances the wire view to the accepted full
// history. It computes operation placement once for a batch and retains the
// previous view unchanged if the state cannot be exported.
func (c *columnarState) reconcile(state *State) error {
	runtimeMetrics.generalReconciles.Add(1)

	changes, ok := state.allChanges()
	if !ok {
		return fmt.Errorf("cannot enumerate canonical changes")
	}

	missing := make([]*opset.Change, 0)

	for _, change := range changes {
		if change.Hash == nil {
			return fmt.Errorf("canonical change has no hash")
		}

		if _, exists := c.changeRows[*change.Hash]; !exists {
			missing = append(missing, change)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	operations, operationRows, operationsAppended, err :=
		c.appendSequenceOperations(state, missing)
	if err != nil {
		return err
	}

	if !operationsAppended {
		var planned bool

		operations, operationRows, planned = c.planObjectBatch(state, missing)
		if !planned {
			c.globalOrderFallbacks++
			runtimeMetrics.globalOrderFallbacks.Add(1)

			order := state.documentOperationOrder()
			operations = make([]opset.Operation, 0, len(order)+64)

			operationRows = newOperationRowIndex(len(order))
			for _, identifier := range order {
				operation, ok := state.operation(identifier)
				if !ok {
					return fmt.Errorf(
						"canonical operation %v is absent",
						identifier,
					)
				}

				operation.Successors = append(
					[]opset.OpID(nil),
					state.successorIndex[identifier]...,
				)
				sort.Slice(operation.Successors, func(i, j int) bool {
					return operation.Successors[i].Compare(operation.Successors[j]) < 0
				})

				operationRows.rows[identifier] = len(operations)
				operations = append(operations, cloneOperation(operation))
			}
		}
	}

	nextChanges := make([]opset.Change, len(changes))

	nextRows := make(map[opset.ChangeHash]int, len(changes))
	for i, change := range changes {
		if row, exists := c.changeRows[*change.Hash]; exists {
			nextChanges[i] = c.changes[row]
		} else {
			nextChanges[i] = cloneChange(*change)
		}

		nextRows[*change.Hash] = i
	}

	actorSet := make(map[opset.ActorID]struct{}, len(c.actors)+len(missing))
	for _, actor := range c.actors {
		actorSet[actor] = struct{}{}
	}

	for i := range nextChanges {
		actorSet[nextChanges[i].Actor] = struct{}{}
	}

	for _, change := range missing {
		for _, operation := range change.Operations {
			actorSet[operation.ID.Actor] = struct{}{}
		}
	}

	actors := make([]opset.ActorID, 0, len(actorSet))
	for actor := range actorSet {
		actors = append(actors, actor)
	}

	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Compare(actors[j]) < 0
	})

	heads := state.Heads()

	headIndexes := make([]uint64, len(heads))
	for i, head := range heads {
		row, exists := nextRows[head]
		if !exists {
			return fmt.Errorf("canonical head %s is absent", head)
		}

		headIndexes[i] = uint64(row)
	}

	dependencyIndexes := make(map[opset.ChangeHash]uint64, len(nextRows))
	for hash, row := range nextRows {
		dependencyIndexes[hash] = uint64(row)
	}

	changeIndex, changeDeleteCount := changeSpliceBounds(
		c.changes,
		nextChanges,
	)
	operationIndex := len(c.operations)
	operationDeleteCount := 0

	var operationSplices []storage.SnapshotOperationSplice

	if !operationsAppended {
		batchOperations := 0
		for _, change := range missing {
			batchOperations += len(change.Operations)
		}

		if batchOperations <= 8 {
			splices, ok := operationSpliceRuns(c.operations, operations)
			if ok && len(splices) <= 8 {
				operationSplices = splices
			}
		}

		if len(operationSplices) == 0 {
			operationIndex, operationDeleteCount = operationSpliceBounds(
				c.operations,
				operations,
			)
		}
	}

	insertedChanges := make(
		[]*opset.Change,
		len(nextChanges)-changeIndex-changeSuffixLength(
			c.changes,
			nextChanges,
			changeIndex,
		),
	)
	for i := range insertedChanges {
		insertedChanges[i] = &nextChanges[changeIndex+i]
	}

	operationSuffix := 0
	if !operationsAppended {
		operationSuffix = operationSuffixLength(
			c.operations,
			operations,
			operationIndex,
		)
	}

	insertedOperations := operations[operationIndex : len(operations)-operationSuffix]

	if operationRows == nil {
		operationRows = c.operationRows

		if len(operationSplices) > 0 {
			for _, splice := range operationSplices {
				operationRows = operationRows.withSplice(
					splice.Operations,
					splice.Index,
					splice.DeleteCount,
				)
			}
		} else {
			operationRows = operationRows.withSplice(
				insertedOperations,
				operationIndex,
				operationDeleteCount,
			)
		}
	}

	document := &opset.Document{Heads: heads, Changes: nextChanges}
	snapshot := c.snapshot.Clone()

	changeAppendOnly := changeIndex == len(c.changes) &&
		changeDeleteCount == 0
	if snapshot == nil || !changeAppendOnly {
		c.snapshotReplacements++
		runtimeMetrics.snapshotReplacements.Add(1)

		var err error

		snapshot, err = storage.NewSnapshotColumns(document, operations)
		if err != nil {
			return fmt.Errorf("cannot replace canonical snapshot columns: %w", err)
		}
	} else if err := snapshot.Splice(storage.SnapshotSplice{
		Actors:               actors,
		Heads:                heads,
		HeadIndexes:          headIndexes,
		DependencyIndexes:    dependencyIndexes,
		ChangeIndex:          changeIndex,
		ChangeDeleteCount:    changeDeleteCount,
		Changes:              insertedChanges,
		OperationIndex:       operationIndex,
		OperationDeleteCount: operationDeleteCount,
		Operations:           insertedOperations,
		OperationSplices:     operationSplices,
	}); err != nil {
		c.snapshotReplacements++
		runtimeMetrics.snapshotReplacements.Add(1)

		if replaceErr := snapshot.Replace(document, operations); replaceErr != nil {
			return fmt.Errorf(
				"cannot splice canonical snapshot columns: %w",
				err,
			)
		}
	}

	c.changes = nextChanges
	c.changeRows = nextRows

	c.heads = append([]opset.ChangeHash(nil), heads...)
	c.operations = operations
	c.operationRows = operationRows
	c.rebuildWireObjectSpans()
	c.columnsDirty = false
	c.snapshot = snapshot
	c.actors = actors
	c.shared = false

	return nil
}

func (c *columnarState) rebuildWireObjectSpans() {
	spans := make(map[opset.ObjectID]wireObjectSpan)
	for row, operation := range c.operations {
		span, ok := spans[operation.Object]
		if !ok {
			spans[operation.Object] = wireObjectSpan{
				start: row,
				end:   row + 1,
			}

			continue
		}

		span.end = row + 1
		spans[operation.Object] = span
	}

	c.objectSpans = spans
}

func (c *columnarState) appendSequenceOperations(
	state *State,
	changes []*opset.Change,
) ([]opset.Operation, *operationRowIndex, bool, error) {
	if len(c.operations) == 0 {
		return nil, nil, false, nil
	}

	added := make([]opset.Operation, 0)
	anchor := c.operations[len(c.operations)-1].ID
	object := c.operations[len(c.operations)-1].Object

	for _, change := range changes {
		for _, changed := range change.Operations {
			if changed.Action == opset.ActionDelete {
				return nil, nil, false, nil
			}

			if !changed.Insert ||
				changed.Key.Element == nil ||
				*changed.Key.Element != anchor ||
				changed.Object != object ||
				len(changed.Predecessors) != 0 {
				return nil, nil, false, nil
			}

			operation, ok := state.operation(changed.ID)
			if !ok {
				return nil, nil, false, fmt.Errorf(
					"canonical operation %v is absent",
					changed.ID,
				)
			}

			operation.Successors = append(
				operation.Successors[:0],
				state.successorIndex[changed.ID]...,
			)
			sort.Slice(operation.Successors, func(i, j int) bool {
				return operation.Successors[i].Compare(operation.Successors[j]) < 0
			})
			added = append(added, cloneOperation(operation))
			anchor = changed.ID
		}
	}

	if len(added) == 0 {
		return c.operations, c.operationRows, true, nil
	}

	operations := c.operations
	if c.shared || cap(operations)-len(operations) < len(added) {
		operations = make(
			[]opset.Operation,
			len(c.operations),
			len(c.operations)+max(64, len(added)),
		)
		copy(operations, c.operations)
	}

	operations = append(operations, added...)
	rows := c.operationRows.with(added, len(c.operations))

	return operations, rows, true, nil
}

func changeSpliceBounds(
	current []opset.Change,
	next []opset.Change,
) (int, int) {
	prefix := 0
	for prefix < len(current) &&
		prefix < len(next) &&
		changeHashesEqual(current[prefix].Hash, next[prefix].Hash) {
		prefix++
	}

	suffix := changeSuffixLength(current, next, prefix)

	return prefix, len(current) - prefix - suffix
}

func changeSuffixLength(
	current []opset.Change,
	next []opset.Change,
	prefix int,
) int {
	suffix := 0
	for suffix < len(current)-prefix &&
		suffix < len(next)-prefix &&
		changeHashesEqual(
			current[len(current)-1-suffix].Hash,
			next[len(next)-1-suffix].Hash,
		) {
		suffix++
	}

	return suffix
}

func changeHashesEqual(
	left *opset.ChangeHash,
	right *opset.ChangeHash,
) bool {
	return left != nil && right != nil && *left == *right
}

func operationSpliceBounds(
	current []opset.Operation,
	next []opset.Operation,
) (int, int) {
	prefix := 0
	for prefix < len(current) &&
		prefix < len(next) &&
		wireOperationsEqual(current[prefix], next[prefix]) {
		prefix++
	}

	suffix := operationSuffixLength(current, next, prefix)

	return prefix, len(current) - prefix - suffix
}

func operationSuffixLength(
	current []opset.Operation,
	next []opset.Operation,
	prefix int,
) int {
	suffix := 0
	for suffix < len(current)-prefix &&
		suffix < len(next)-prefix &&
		wireOperationsEqual(
			current[len(current)-1-suffix],
			next[len(next)-1-suffix],
		) {
		suffix++
	}

	return suffix
}

func wireOperationsEqual(left, right opset.Operation) bool {
	left.Predecessors = nil
	right.Predecessors = nil

	return reflect.DeepEqual(left, right)
}

func (c *columnarState) document(
	heads []opset.ChangeHash,
	unknown []opset.RawColumn,
) (*opset.Document, []opset.Operation) {
	document := &opset.Document{
		Actors:           append([]opset.ActorID(nil), c.actors...),
		Heads:            append([]opset.ChangeHash(nil), heads...),
		Changes:          c.changes,
		UnknownColumns:   cloneRawColumns(unknown),
		ChangeColumns:    cloneRawColumns(c.changeColumns),
		OperationColumns: cloneRawColumns(c.operationColumns),
	}

	return document, c.operations
}

func cloneChanges(changes []opset.Change) []opset.Change {
	cloned := make([]opset.Change, len(changes))
	for i := range changes {
		cloned[i] = cloneChange(changes[i])
	}

	return cloned
}

func cloneChange(change opset.Change) opset.Change {
	change.Dependencies = append([]opset.ChangeHash(nil), change.Dependencies...)
	change.DependencyIndexes = append([]uint64(nil), change.DependencyIndexes...)
	operations := change.Operations

	change.Operations = make([]opset.Operation, len(operations))
	for i := range operations {
		change.Operations[i] = cloneOperation(operations[i])
	}

	change.ExtraBytes = append([]byte(nil), change.ExtraBytes...)
	change.Raw = append([]byte(nil), change.Raw...)

	return change
}

func cloneOperation(operation opset.Operation) opset.Operation {
	operation.Predecessors = append([]opset.OpID(nil), operation.Predecessors...)

	operation.Successors = append([]opset.OpID(nil), operation.Successors...)
	if operation.Value != nil {
		value := *operation.Value
		value.Bytes = append([]byte(nil), value.Bytes...)
		operation.Value = &value
	}

	return operation
}
