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
	"sort"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

type MergeBatch struct {
	changes [][]byte
}

func (b *Engine) CanDirectMerge() bool {
	return !b.isolationActive
}

func (b *Engine) Heads() ([][32]byte, error) {
	heads := b.currentHeads()

	result := make([][32]byte, len(heads))
	for i := range heads {
		result[i] = [32]byte(heads[i])
	}

	return result, nil
}

func (b *Engine) HasHeads(
	heads [][32]byte,
) (bool, error) {
	for _, head := range heads {
		if _, ok := b.lookupChange(opset.ChangeHash(head)); !ok {
			return false, nil
		}
	}

	return true, nil
}

func (b *Engine) MissingDependencies(
	heads [][32]byte,
) ([][32]byte, error) {
	missing := make(map[[32]byte]struct{})

	for _, head := range heads {
		_, queued := b.queuedChanges[opset.ChangeHash(head)]
		if _, ok := b.lookupChange(opset.ChangeHash(head)); !ok && !queued {
			missing[head] = struct{}{}
		}
	}

	for _, change := range b.queuedChanges {
		for _, dependency := range change.Dependencies {
			if _, ok := b.lookupChange(dependency); !ok {
				missing[[32]byte(dependency)] = struct{}{}
			}
		}
	}

	result := make([][32]byte, 0, len(missing))
	for dependency := range missing {
		result = append(result, dependency)
	}

	sort.Slice(
		result,
		func(i, j int) bool {
			return bytes.Compare(result[i][:], result[j][:]) < 0
		},
	)

	return result, nil
}

func (b *Engine) ChangesSince(
	heads [][32]byte,
) ([][]byte, [][32]byte, error) {
	knownHeads := make([]opset.ChangeHash, len(heads))
	for i, head := range heads {
		knownHeads[i] = opset.ChangeHash(head)
	}

	// A change in the frontier's ancestry may occasionally be unreachable, for
	// example after a merge that rebuilt the graph. changesSince then returns the
	// consistent, replayable prefix it can produce rather than nothing: reporting
	// every change that can be emitted keeps collaboration alive, where failing
	// the whole read would wedge the document on every request. A change that is
	// dropped has no bytes to return in any case.
	changes, _ := b.state.changesSince(knownHeads)

	raw := make([][]byte, len(changes))

	hashes := make([][32]byte, len(changes))
	for i, change := range changes {
		raw[i] = append([]byte(nil), change.Raw...)
		hashes[i] = [32]byte(*change.Hash)
	}

	return raw, hashes, nil
}

func (b *Engine) ApplyChanges(
	changes [][]byte,
) error {
	// Bulk changes received while isolated advance the retained full history,
	// exactly like Merge, without changing the pinned query view.
	if b.isolationActive && b.fullState != nil {
		b.state, b.fullState = b.fullState, b.state

		defer func() {
			b.state, b.fullState = b.fullState, b.state
			b.nextOp = b.fullState.maxOpGlobal() + 1
		}()
	}

	decoded := make([]opset.Change, 0, len(changes))
	for i, raw := range changes {
		document, _, err := storage.DecodeIncremental(raw)
		if err != nil {
			return fmt.Errorf("cannot decode native change %d: %w", i, err)
		}

		decoded = append(decoded, document.Changes...)
	}

	beforeChanges := b.state.changeCount()

	beforeQueued := len(b.queuedChanges)
	if err := b.applyMergedChanges(decoded); err != nil {
		return err
	}

	if b.state.changeCount() != beforeChanges || len(b.queuedChanges) != beforeQueued {
		b.revision++
	}

	if next := b.state.maxOpGlobal() + 1; next > b.nextOp {
		b.nextOp = next
	}

	return nil
}

func (b *Engine) ChangeHashes() [][32]byte {
	hashes := make([][32]byte, 0, b.state.changeCount())
	b.state.eachChange(func(hash opset.ChangeHash, _ *opset.Change) bool {
		hashes = append(hashes, [32]byte(hash))
		return true
	})

	return hashes
}

func (b *Engine) PrepareMerge(known [][32]byte) (*MergeBatch, error) {
	if !b.CanDirectMerge() {
		return nil, fmt.Errorf("cannot prepare direct merge while isolated")
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit("", time.Time{}); err != nil {
			return nil, fmt.Errorf("cannot commit merge source: %w", err)
		}
	}

	knownSet := make(map[opset.ChangeHash]struct{}, len(known))
	for _, hash := range known {
		knownSet[opset.ChangeHash(hash)] = struct{}{}
	}

	changes, ok := b.state.allChanges()
	if !ok {
		return nil, fmt.Errorf("cannot enumerate merge source changes")
	}

	batch := &MergeBatch{changes: make([][]byte, 0)}
	for _, source := range changes {
		if source.Hash == nil {
			return nil, fmt.Errorf("cannot merge change without hash")
		}

		if _, ok := knownSet[*source.Hash]; ok {
			continue
		}

		change := *source

		change.Raw = append([]byte(nil), source.Raw...)
		if len(change.Raw) == 0 {
			expected := *change.Hash

			raw, err := storage.EncodeChange(&change)
			if err != nil {
				return nil, fmt.Errorf("cannot encode merge change: %w", err)
			}

			if change.Hash == nil || *change.Hash != expected {
				return nil, fmt.Errorf("cannot reproduce merge change hash")
			}

			document, err := storage.DecodePartial(raw)
			if err != nil {
				return nil, fmt.Errorf("cannot normalize merge change: %w", err)
			}

			if len(document.Changes) != 1 {
				return nil, fmt.Errorf(
					"cannot normalize merge change: decoded %d changes",
					len(document.Changes),
				)
			}

			change = document.Changes[0]
		}

		batch.changes = append(batch.changes, append([]byte(nil), change.Raw...))
	}

	for _, source := range orderedQueuedChanges(b.queuedChanges) {
		if source.Hash == nil {
			continue
		}

		if _, ok := knownSet[*source.Hash]; ok {
			continue
		}

		batch.changes = append(batch.changes, append([]byte(nil), source.Raw...))
	}

	return batch, nil
}

func (b *Engine) ApplyMerge(batch *MergeBatch) ([][32]byte, error) {
	if !b.CanDirectMerge() {
		return nil, fmt.Errorf("cannot apply direct merge while isolated")
	}

	if batch == nil {
		return b.Heads()
	}

	if err := b.ApplyChanges(batch.changes); err != nil {
		return nil, fmt.Errorf("cannot apply direct merge batch: %w", err)
	}

	return b.Heads()
}

func (b *Engine) Merge(data []byte) ([][32]byte, error) {
	// While isolated, merged changes belong to the full history rather than the
	// pinned view, so operate on the full state and keep the pinned view intact.
	if b.isolationActive && b.fullState != nil {
		b.state, b.fullState = b.fullState, b.state

		defer func() {
			b.state, b.fullState = b.fullState, b.state
			b.nextOp = b.fullState.maxOpGlobal() + 1
		}()
	}

	document, err := storage.Decode(data)
	if err != nil {
		document, err = storage.DecodePartial(data)
	}

	if err != nil {
		return nil, err
	}

	if len(b.state.Heads()) == 0 && len(b.pending) == 0 {
		state, err := NewStateFromDocument(document)
		if err != nil {
			return nil, fmt.Errorf("cannot initialize merged native state: %w", err)
		}

		b.state = state
		b.nextOp = state.maxOpGlobal() + 1

		columns, err := newColumnarState(document)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot initialize merged document columns: %w",
				err,
			)
		}

		b.columns = columns
		b.unknownColumns = cloneRawColumns(document.UnknownColumns)

		b.base = append([]byte(nil), data...)
		b.appended = nil
		b.saveCursor = 0

		b.revision++

		return b.Heads()
	}

	// A document chunk carries semantic change rows rather than standalone
	// change frames. Reproduce only missing frames here so the ordinary remote
	// batch can retain and forward them without rebuilding the local snapshot.
	// If an old/hash-only snapshot lacks enough metadata, preserve the
	// compatibility snapshot merge below.
	_ = b.restoreDocumentChangeFrames(document)
	if b.requiresSnapshotMerge(document) {
		beforeChanges := b.state.changeCount()

		beforeQueued := len(b.queuedChanges)
		if err := b.mergeDocumentSnapshot(data, document); err != nil {
			return nil, err
		}

		if b.state.changeCount() != beforeChanges || len(b.queuedChanges) != beforeQueued {
			b.revision++
		}

		return b.Heads()
	}

	beforeChanges := b.state.changeCount()

	beforeQueued := len(b.queuedChanges)
	if err := b.applyMergedChanges(document.Changes); err != nil {
		return nil, err
	}

	if b.state.changeCount() != beforeChanges || len(b.queuedChanges) != beforeQueued {
		b.revision++
	}

	if next := b.state.maxOpGlobal() + 1; next > b.nextOp {
		b.nextOp = next
	}

	return b.Heads()
}

func (b *Engine) restoreDocumentChangeFrames(document *opset.Document) error {
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash == nil ||
			b.state.hasChange(*change.Hash) ||
			len(change.Raw) > 0 {
			continue
		}

		expected := *change.Hash

		raw, err := storage.EncodeChange(change)
		if err != nil {
			return fmt.Errorf("cannot encode snapshot change: %w", err)
		}

		if change.Hash == nil || *change.Hash != expected {
			return fmt.Errorf("cannot reproduce snapshot change hash")
		}

		decoded, err := storage.DecodePartial(raw)
		if err != nil {
			return fmt.Errorf("cannot decode reproduced snapshot change: %w", err)
		}

		if len(decoded.Changes) != 1 {
			return fmt.Errorf(
				"cannot reproduce snapshot change: decoded %d changes",
				len(decoded.Changes),
			)
		}

		document.Changes[i] = decoded.Changes[0]
	}

	return nil
}

func (b *Engine) requiresSnapshotMerge(document *opset.Document) bool {
	if len(document.ChunkTypes) == 0 ||
		document.ChunkTypes[0] != opset.ChunkDocument {
		return false
	}

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash != nil &&
			!b.state.hasChange(*change.Hash) &&
			len(change.Raw) == 0 {
			return true
		}
	}

	return false
}

func (b *Engine) mergeDocumentSnapshot(
	data []byte,
	document *opset.Document,
) error {
	localChanges, ok := b.state.allChanges()
	if !ok {
		return fmt.Errorf("cannot enumerate local changes for snapshot merge")
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return fmt.Errorf("cannot initialize merged snapshot state: %w", err)
	}

	for _, change := range localChanges {
		if change.Hash == nil || state.hasChange(*change.Hash) {
			continue
		}

		incoming := documentChangeByActorSequence(
			document,
			change.Actor,
			change.Sequence,
		)
		if incoming != nil {
			state.changes[*change.Hash] = incoming
		}
	}

	appended := make([][]byte, 0)

	for _, change := range localChanges {
		if change.Hash == nil ||
			state.hasChange(*change.Hash) ||
			documentChangeByActorSequence(
				document,
				change.Actor,
				change.Sequence,
			) != nil {
			continue
		}

		if len(change.Raw) == 0 {
			return fmt.Errorf(
				"cannot preserve local change %s during snapshot merge",
				change.Hash,
			)
		}

		if err := state.ApplyChange(change); err != nil {
			return fmt.Errorf("cannot apply local change to merged snapshot: %w", err)
		}

		appended = append(appended, append([]byte(nil), change.Raw...))
	}

	b.state = state

	b.base = append([]byte(nil), data...)
	b.appended = appended
	b.saveCursor = 0
	b.queuedChanges = make(map[opset.ChangeHash]*opset.Change)
	b.queuedBytes = 0
	b.nextOp = state.maxOpGlobal() + 1

	columns, err := newColumnarStateFromState(state)
	if err != nil {
		return fmt.Errorf("cannot rebuild merged document columns: %w", err)
	}

	b.columns = columns

	return nil
}

func documentChangeByActorSequence(
	document *opset.Document,
	actor opset.ActorID,
	sequence uint64,
) *opset.Change {
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Actor == actor && change.Sequence == sequence {
			return change
		}
	}

	return nil
}

func (b *Engine) applyMergedChanges(changes []opset.Change) error {
	direct, err := b.applyDirectRemoteChanges(changes)
	if err != nil {
		return err
	}

	if direct {
		return nil
	}

	for i := range changes {
		change := &changes[i]
		if change.Hash == nil || b.state.hasChange(*change.Hash) {
			continue
		}

		if _, queued := b.queuedChanges[*change.Hash]; queued {
			continue
		}

		if len(change.Raw) == 0 {
			return fmt.Errorf(
				"cannot preserve merged change %s: original bytes are unavailable",
				change.Hash,
			)
		}

		if len(b.queuedChanges) >= maxQueuedChanges ||
			b.queuedBytes+len(change.Raw) > maxQueuedChangeBytes {
			return fmt.Errorf("merged change queue exceeds its resource limit")
		}

		clone := *change
		clone.Raw = append([]byte(nil), change.Raw...)
		b.queuedChanges[*change.Hash] = &clone
		b.queuedBytes += len(clone.Raw)
	}

	for len(b.queuedChanges) > 0 {
		progressed := false

		for hash, change := range b.queuedChanges {
			if !b.state.hasDependencies(change) {
				continue
			}

			if err := b.state.ApplyChange(change); err != nil {
				return fmt.Errorf("cannot apply merged native change: %w", err)
			}

			b.appended = append(
				b.appended,
				append([]byte(nil), change.Raw...),
			)
			b.queuedBytes -= len(change.Raw)
			delete(b.queuedChanges, hash)

			progressed = true
		}

		if !progressed {
			break
		}
	}

	if err := b.reconcileColumns(); err != nil {
		return fmt.Errorf("cannot update canonical document columns: %w", err)
	}

	return nil
}

// applyDirectRemoteChanges stages a complete incoming set against private query
// indexes and publishes its columns, indexes, and orphan queue together. It is
// deliberately all-or-nothing: a malformed late change or a failed snapshot
// splice leaves every engine-owned pointer and byte count unchanged.
func (b *Engine) applyDirectRemoteChanges(changes []opset.Change) (bool, error) {
	if b.columns == nil ||
		b.columns.snapshot == nil ||
		b.columns.columnsDirty ||
		len(b.pending) != 0 {
		return false, nil
	}

	current := b.state
	if b.isolationActive && b.fullState != nil {
		current = b.fullState
	}

	if current.changeCount() != len(b.columns.changes) {
		return false, nil
	}

	nextQueue := make(
		map[opset.ChangeHash]*opset.Change,
		len(b.queuedChanges)+len(changes),
	)
	nextQueuedBytes := 0

	for hash, change := range b.queuedChanges {
		cloned := cloneChange(*change)
		nextQueue[hash] = &cloned
		nextQueuedBytes += len(cloned.Raw)
	}

	for i := range changes {
		change := &changes[i]
		if change.Hash == nil || current.hasChange(*change.Hash) {
			continue
		}

		if _, queued := nextQueue[*change.Hash]; queued {
			continue
		}

		if len(change.Raw) == 0 {
			return true, fmt.Errorf(
				"cannot preserve merged change %s: original bytes are unavailable",
				change.Hash,
			)
		}

		if len(nextQueue) >= maxQueuedChanges ||
			nextQueuedBytes+len(change.Raw) > maxQueuedChangeBytes {
			return true, fmt.Errorf("merged change queue exceeds its resource limit")
		}

		cloned := cloneChange(*change)
		nextQueue[*change.Hash] = &cloned
		nextQueuedBytes += len(cloned.Raw)
	}

	applied, err := validateDirectRemoteQueue(current, nextQueue)
	if err != nil {
		return true, err
	}

	for _, change := range applied {
		nextQueuedBytes -= len(change.Raw)
		delete(nextQueue, *change.Hash)
	}

	directSequence := len(changes) == 1 &&
		len(b.queuedChanges) == 0 &&
		len(nextQueue) == 0 &&
		isDirectSequenceRemoteBatch(applied)

	planning := current
	if len(applied) > 0 && !directSequence {
		planning, err = stateFromSharedColumns(b.columns)
		if err != nil {
			return true, fmt.Errorf("cannot clone remote query state: %w", err)
		}

		for _, change := range applied {
			if err := planning.ApplyChange(change); err != nil {
				return true, fmt.Errorf(
					"cannot stage merged native change: %w",
					err,
				)
			}
		}
	}

	preApplied := false

	if len(applied) > 0 && directSequence {
		current.directRemoteSequence = true
		for _, change := range applied {
			if err := current.ApplyChange(change); err != nil {
				current.directRemoteSequence = false

				return true, fmt.Errorf(
					"cannot stage direct sequence change: %w",
					err,
				)
			}
		}

		current.directRemoteSequence = false
		preApplied = true
	}

	nextColumns := b.columns
	if len(applied) > 0 {
		batch, err := newColumnMutationBatch(
			b.columns,
			planning,
			applied,
			directSequence,
		)
		if errors.Is(err, errDirectColumnsUnsupported) {
			if preApplied {
				b.restoreDirectRemoteState()
			}

			return false, nil
		}

		if err != nil {
			if preApplied {
				b.restoreDirectRemoteState()
			}

			return true, fmt.Errorf("cannot plan direct remote columns: %w", err)
		}

		batch.reuseOperations = directSequence

		if b.directColumnFailure != nil {
			if err := b.directColumnFailure(); err != nil {
				if preApplied {
					b.restoreDirectRemoteState()
				}

				return true, fmt.Errorf(
					"cannot pass direct column failpoint: %w",
					err,
				)
			}
		}

		nextColumns, err = batch.apply(b.columns)
		if err != nil {
			if preApplied {
				b.restoreDirectRemoteState()
			}

			return false, nil
		}

		if !preApplied {
			current.directRemoteSequence = directSequence
			for _, change := range applied {
				if err := current.ApplyChange(change); err != nil {
					panic(fmt.Sprintf(
						"validated direct remote change failed publication: %v",
						err,
					))
				}
			}

			current.directRemoteSequence = false
		}

		current.attachCanonical(nextColumns)
	}

	for _, change := range applied {
		b.appended = append(b.appended, append([]byte(nil), change.Raw...))
	}

	b.columns = nextColumns
	if b.isolationActive && b.fullState != nil {
		b.fullState = current
	} else {
		b.state = current
	}

	b.queuedChanges = nextQueue
	b.queuedBytes = nextQueuedBytes

	return true, nil
}

func (b *Engine) restoreDirectRemoteState() {
	restored, err := stateFromSharedColumns(b.columns)
	if err != nil {
		panic(fmt.Sprintf("cannot restore failed direct remote state: %v", err))
	}

	if b.isolationActive && b.fullState != nil {
		b.fullState = restored
		return
	}

	b.state = restored
}

func validateDirectRemoteQueue(
	current *State,
	queue map[opset.ChangeHash]*opset.Change,
) ([]*opset.Change, error) {
	remaining := make(map[opset.ChangeHash]*opset.Change, len(queue))
	for hash, change := range queue {
		remaining[hash] = change
	}

	known := make(map[opset.ChangeHash]struct{}, len(queue))
	sequences := make(map[opset.ActorID]uint64)
	operationIDs := make(map[opset.OpID]struct{})
	applied := make([]*opset.Change, 0, len(queue))

	for {
		progressed := false

		for _, change := range orderedQueuedChanges(remaining) {
			dependenciesPresent := true

			for _, dependency := range change.Dependencies {
				if current.hasChange(dependency) {
					continue
				}

				if _, ok := known[dependency]; !ok {
					dependenciesPresent = false
					break
				}
			}

			if !dependenciesPresent {
				continue
			}

			if err := validateChangeSnapshotDomain(change); err != nil {
				return nil, fmt.Errorf(
					"cannot validate merged snapshot domain: %w",
					err,
				)
			}

			sequence, ok := sequences[change.Actor]
			if !ok {
				sequence = current.sequenceForActor(change.Actor)
			}

			if change.Sequence != sequence+1 {
				return nil, fmt.Errorf(
					"cannot validate merged actor sequence: actor sequence is %d, expected %d",
					change.Sequence,
					sequence+1,
				)
			}

			for _, operation := range change.Operations {
				if current.hasOperationID(operation.ID) {
					return nil, fmt.Errorf(
						"cannot validate merged operation: duplicate operation ID %v",
						operation.ID,
					)
				}

				if _, exists := operationIDs[operation.ID]; exists {
					return nil, fmt.Errorf(
						"cannot validate merged operation set: duplicate operation ID %v",
						operation.ID,
					)
				}

				operationIDs[operation.ID] = struct{}{}
			}

			sequences[change.Actor] = change.Sequence
			known[*change.Hash] = struct{}{}
			applied = append(applied, change)
			delete(remaining, *change.Hash)

			progressed = true
		}

		if !progressed {
			return applied, nil
		}
	}
}

func isDirectSequenceRemoteBatch(changes []*opset.Change) bool {
	var (
		object   opset.ObjectID
		previous opset.OpID
		seen     bool
		count    int
	)

	for _, change := range changes {
		for _, operation := range change.Operations {
			count++
			if count > 1 {
				return false
			}

			if operation.Action == opset.ActionDelete ||
				operation.Action == opset.ActionMark ||
				isObjectAction(operation.Action) ||
				!operation.Insert ||
				operation.Object.IsRoot ||
				len(operation.Predecessors) != 0 {
				return false
			}

			if !seen {
				if !operation.Key.IsHead && operation.Key.Element == nil {
					return false
				}

				object = operation.Object
				seen = true
			} else if operation.Object != object ||
				operation.Key.Element == nil ||
				*operation.Key.Element != previous {
				return false
			}

			previous = operation.ID
		}
	}

	return seen && count == 1
}
