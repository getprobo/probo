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
	"math"
	"slices"
	"sort"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

// changeDependencies computes the dependency set for a new change authored by
// this backend's actor at the given sequence number. The dependencies are the
// current heads plus, matching upstream Rust, the actor's own previous change
// hash when it is not already a head (so that direct causal succession from the
// author's prior change is always recorded explicitly).
func (b *Engine) changeDependencies(sequence uint64) []opset.ChangeHash {
	dependencies := b.currentHeads()

	if sequence > 1 {
		last, ok := b.state.hashForActorSequence(b.actor, sequence-1)
		if ok && !containsHash(dependencies, last) {
			dependencies = append(dependencies, last)
			sort.Slice(
				dependencies,
				func(i, j int) bool {
					return bytes.Compare(dependencies[i][:], dependencies[j][:]) < 0
				},
			)
		}
	}

	return dependencies
}

func containsHash(hashes []opset.ChangeHash, target opset.ChangeHash) bool {
	return slices.Contains(hashes, target)
}

// Isolate pins the document to the given heads: subsequent reads reflect that
// frontier plus isolated writes, and new changes branch from it using a derived
// isolation actor so they never collide with the base actor's later history. It
// mirrors Rust's AutoCommit::isolate. Repeated calls re-pin to fresh heads.
func (b *Engine) Isolate(heads [][32]byte) error {
	if len(b.pending) > 0 {
		if _, err := b.Commit("", time.Time{}); err != nil {
			return err
		}
	}

	full := b.fullState
	if !b.isolationActive {
		full = b.state
	}

	nativeHeads := nativeHashes(heads)

	pinned, ok := newIsolationView(full, b.columns, nativeHeads)
	if !ok {
		return fmt.Errorf("isolation heads are unknown")
	}

	baseActor := b.baseActor
	if !b.isolationActive {
		baseActor = b.actor
	}

	b.isolationActive = true
	b.fullState = full
	b.baseActor = baseActor
	b.state = pinned
	b.actor = isolationActor(full, pinned, baseActor)
	b.nextOp = full.maxOpGlobal() + 1
	b.revision++

	b.isolationDiffTargets = append(
		b.isolationDiffTargets,
		append([][32]byte(nil), nativeToArrayHeads(nativeHeads)...),
	)

	return nil
}

// nativeToArrayHeads converts change hashes to the [32]byte head form used by
// the incremental diff cursor.
func nativeToArrayHeads(heads []opset.ChangeHash) [][32]byte {
	result := make([][32]byte, len(heads))
	for i, hash := range heads {
		result[i] = [32]byte(hash)
	}

	return result
}

// Integrate ends isolation, returning reads and writes to the full history that
// accumulated every isolated and merged change. It mirrors AutoCommit::integrate.
func (b *Engine) Integrate() error {
	if !b.isolationActive {
		return nil
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit("", time.Time{}); err != nil {
			return err
		}
	}

	b.state = b.fullState
	b.actor = b.baseActor
	b.fullState = nil
	b.isolationActive = false
	b.nextOp = b.state.maxOpGlobal() + 1
	b.revision++

	return nil
}

// isolationActor selects the actor for isolated writes: the base actor when all
// of its operations are already covered by the isolation heads, otherwise the
// lowest-level derived concurrency actor whose operations are covered, matching
// Rust's isolate_actor.
func isolationActor(full, pinned *State, base opset.ActorID) opset.ActorID {
	for level := uint64(0); ; level++ {
		candidate := base.WithConcurrency(level)
		if full.maxOpForActor(candidate) == pinned.maxOpForActor(candidate) {
			return candidate
		}
	}
}

func (b *Engine) Commit(
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if len(b.pending) == 0 {
		return [32]byte{}, fmt.Errorf("change contains no operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1
	if sequence > math.MaxUint32 ||
		b.pending[len(b.pending)-1].ID.Counter > math.MaxUint32 {
		return [32]byte{}, fmt.Errorf(
			"change exceeds snapshot uint32 domain",
		)
	}
	dependencies := b.changeDependencies(sequence)

	change := &opset.Change{
		Actor:        b.actor,
		Sequence:     sequence,
		StartOp:      b.pending[0].ID.Counter,
		MaxOp:        b.pending[len(b.pending)-1].ID.Counter,
		Time:         timestamp.Unix(),
		Message:      message,
		Dependencies: dependencies,
		Operations:   append([]opset.Operation(nil), b.pending...),
	}
	if timestamp.IsZero() {
		change.Time = 0
	}

	raw, err := storage.EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native change: %w", err)
	}

	nextColumns, nextFullState, direct, err := b.applyDirectColumnCommit(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf(
			"cannot update canonical document columns: %w",
			err,
		)
	}

	// While isolated, the pinned view holds the change for subsequent reads, but
	// the full history must also record it so integration sees every isolated
	// change alongside merges. Changes and raw bytes are immutable after encode,
	// while State copies operation values into its own mutable query indexes.
	if b.isolationActive && b.fullState != nil {
		if direct {
			b.fullState = nextFullState
		} else if err := b.fullState.ApplyChange(change); err != nil {
			return [32]byte{}, fmt.Errorf(
				"cannot apply isolated change to full history: %w",
				err,
			)
		}

		if next := b.fullState.maxOpGlobal() + 1; next > b.nextOp {
			b.nextOp = next
		}
	}
	b.state.recordAppliedChange(change)
	if direct {
		b.columns = nextColumns
		if b.isolationActive && b.fullState != nil {
			b.fullState.promoteDirectCommit(change, nextColumns)
		} else {
			b.state.promoteDirectCommit(change, nextColumns)
		}
	} else {
		if err := b.reconcileColumns(); err != nil {
			return [32]byte{}, fmt.Errorf(
				"cannot update canonical document columns: %w",
				err,
			)
		}
	}

	b.appended = append(b.appended, raw)
	b.pending = nil
	b.revision++

	return [32]byte(*change.Hash), nil
}

func (b *Engine) EmptyCommit(
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if len(b.pending) != 0 {
		return [32]byte{}, fmt.Errorf("cannot create empty change with pending operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1
	if sequence > math.MaxUint32 || b.nextOp-1 > math.MaxUint32 {
		return [32]byte{}, fmt.Errorf(
			"change exceeds snapshot uint32 domain",
		)
	}

	change := &opset.Change{
		Actor:        b.actor,
		Sequence:     sequence,
		StartOp:      b.nextOp,
		MaxOp:        b.nextOp - 1,
		Time:         timestamp.Unix(),
		Message:      message,
		Dependencies: b.changeDependencies(sequence),
	}
	if timestamp.IsZero() {
		change.Time = 0
	}

	raw, err := storage.EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native empty change: %w", err)
	}

	nextColumns, nextFullState, direct, err := b.applyDirectColumnCommit(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf(
			"cannot update canonical document columns: %w",
			err,
		)
	}
	if b.isolationActive && b.fullState != nil {
		if direct {
			b.fullState = nextFullState
		} else if err := b.fullState.ApplyChange(change); err != nil {
			return [32]byte{}, fmt.Errorf(
				"cannot apply isolated empty change to full history: %w",
				err,
			)
		}
	}
	b.state.recordAppliedChange(change)
	if direct {
		b.columns = nextColumns
		if b.isolationActive && b.fullState != nil {
			b.fullState.promoteDirectCommit(change, nextColumns)
		} else {
			b.state.promoteDirectCommit(change, nextColumns)
		}
	} else {
		if err := b.reconcileColumns(); err != nil {
			return [32]byte{}, fmt.Errorf(
				"cannot update canonical document columns: %w",
				err,
			)
		}
	}

	b.appended = append(b.appended, raw)
	b.revision++

	return [32]byte(*change.Hash), nil
}

func (b *Engine) applyDirectColumnCommit(
	change *opset.Change,
) (*columnarState, *State, bool, error) {
	planningState := b.state
	var nextFullState *State
	changes := []*opset.Change{change}
	if b.isolationActive && b.fullState != nil {
		var err error
		nextFullState, changes, err = directIsolationState(
			b.columns,
			b.fullState,
			change,
		)
		if err != nil {
			return nil, nil, false, fmt.Errorf(
				"cannot prepare isolated direct state: %w",
				err,
			)
		}
		planningState = nextFullState
	}
	batch, err := newColumnMutationBatch(
		b.columns,
		planningState,
		changes,
		false,
	)
	if errors.Is(err, errDirectColumnsUnsupported) {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	if b.directColumnFailure != nil {
		if err := b.directColumnFailure(); err != nil {
			return nil, nil, false, fmt.Errorf(
				"cannot pass direct column failpoint: %w",
				err,
			)
		}
	}
	columns, err := batch.apply(b.columns)
	if err != nil {
		return nil, nil, false, err
	}
	return columns, nextFullState, true, nil
}

func directIsolationState(
	columns *columnarState,
	full *State,
	change *opset.Change,
) (*State, []*opset.Change, error) {
	state, err := stateFromSharedColumns(columns)
	if err != nil {
		return nil, nil, err
	}
	remaining := make(map[opset.ChangeHash]*opset.Change)
	full.eachChange(func(hash opset.ChangeHash, source *opset.Change) bool {
		if _, canonical := columns.changeRows[hash]; canonical {
			return true
		}
		cloned := cloneChange(*source)
		remaining[hash] = &cloned
		return true
	})
	applied := make([]*opset.Change, 0, len(remaining)+1)
	for len(remaining) > 0 {
		progressed := false
		for hash, source := range remaining {
			if !state.hasDependencies(source) {
				continue
			}
			if err := state.ApplyChange(source); err != nil {
				return nil, nil, fmt.Errorf(
					"cannot replay isolated overlay change: %w",
					err,
				)
			}
			applied = append(applied, source)
			delete(remaining, hash)
			progressed = true
		}
		if !progressed {
			return nil, nil, fmt.Errorf("cannot order isolated overlay changes")
		}
	}
	if err := state.ApplyChange(change); err != nil {
		return nil, nil, fmt.Errorf(
			"cannot apply isolated direct change: %w",
			err,
		)
	}
	applied = append(applied, change)
	return state, applied, nil
}

func (b *Engine) Rollback() (uint64, error) {
	if len(b.pending) == 0 {
		return 0, nil
	}

	cancelled := uint64(len(b.pending))

	rolledBack := make(map[opset.OpID]struct{}, len(b.pending))
	for _, operation := range b.pending {
		rolledBack[operation.ID] = struct{}{}
	}

	b.state.undoPending(b.pending)

	for handle, object := range b.objects {
		if object.IsRoot {
			continue
		}

		if _, ok := rolledBack[object.OpID]; ok {
			delete(b.objects, handle)
		}
	}

	b.nextOp = b.state.maxOpGlobal() + 1
	if b.isolationActive && b.fullState != nil {
		b.nextOp = max(b.nextOp, b.fullState.maxOpGlobal()+1)
	}

	b.pending = nil
	b.revision++

	return cancelled, nil
}
