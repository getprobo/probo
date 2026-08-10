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
	"context"
	"fmt"
	"slices"
	"sort"
	"time"
)

// changeDependencies computes the dependency set for a new change authored by
// this backend's actor at the given sequence number. The dependencies are the
// current heads plus, matching upstream Rust, the actor's own previous change
// hash when it is not already a head (so that direct causal succession from the
// author's prior change is always recorded explicitly).
func (b *Engine) changeDependencies(sequence uint64) []ChangeHash {
	dependencies := b.state.Heads()

	if sequence > 1 {
		last, ok := b.state.hashForActorSequence(b.actor, sequence-1)
		if ok && !containsHash(dependencies, last) {
			dependencies = append(dependencies, last)
			sort.Slice(dependencies, func(i, j int) bool {
				return bytes.Compare(dependencies[i][:], dependencies[j][:]) < 0
			})
		}
	}

	return dependencies
}

func containsHash(hashes []ChangeHash, target ChangeHash) bool {
	return slices.Contains(hashes, target)
}

// Isolate pins the document to the given heads: subsequent reads reflect that
// frontier plus isolated writes, and new changes branch from it using a derived
// isolation actor so they never collide with the base actor's later history. It
// mirrors Rust's AutoCommit::isolate. Repeated calls re-pin to fresh heads.
func (b *Engine) Isolate(ctx context.Context, heads [][32]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return err
		}
	}

	full := b.fullState
	if !b.isolationActive {
		full = b.state
	}

	nativeHeads := nativeHashes(heads)

	pinned, ok := full.at(nativeHeads)
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

	b.isolationDiffTargets = append(
		b.isolationDiffTargets,
		append([][32]byte(nil), nativeToArrayHeads(nativeHeads)...),
	)

	return nil
}

// nativeToArrayHeads converts change hashes to the [32]byte head form used by
// the incremental diff cursor.
func nativeToArrayHeads(heads []ChangeHash) [][32]byte {
	result := make([][32]byte, len(heads))
	for i, hash := range heads {
		result[i] = [32]byte(hash)
	}

	return result
}

// Integrate ends isolation, returning reads and writes to the full history that
// accumulated every isolated and merged change. It mirrors AutoCommit::integrate.
func (b *Engine) Integrate(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if !b.isolationActive {
		return nil
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return err
		}
	}

	b.state = b.fullState
	b.actor = b.baseActor
	b.fullState = nil
	b.isolationActive = false
	b.nextOp = b.state.maxOpGlobal() + 1

	return nil
}

// isolationActor selects the actor for isolated writes: the base actor when all
// of its operations are already covered by the isolation heads, otherwise the
// lowest-level derived concurrency actor whose operations are covered, matching
// Rust's isolate_actor.
func isolationActor(full, pinned *State, base ActorID) ActorID {
	for level := uint64(0); ; level++ {
		candidate := base.WithConcurrency(level)
		if full.maxOpForActor(candidate) == pinned.maxOpForActor(candidate) {
			return candidate
		}
	}
}

func (b *Engine) Commit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if err := ctx.Err(); err != nil {
		return [32]byte{}, err
	}

	if len(b.pending) == 0 {
		return [32]byte{}, fmt.Errorf("change contains no operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1
	dependencies := b.changeDependencies(sequence)

	change := &Change{
		Actor:        b.actor,
		Sequence:     sequence,
		StartOp:      b.pending[0].ID.Counter,
		MaxOp:        b.pending[len(b.pending)-1].ID.Counter,
		Time:         timestamp.Unix(),
		Message:      message,
		Dependencies: dependencies,
		Operations:   append([]Operation(nil), b.pending...),
	}
	if timestamp.IsZero() {
		change.Time = 0
	}

	raw, err := EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native change: %w", err)
	}

	if err := b.state.recordAppliedChange(change); err != nil {
		return [32]byte{}, err
	}

	// While isolated, the pinned view holds the change for subsequent reads, but
	// the full history must also record it so integration sees every isolated
	// change alongside merges. Decode a fresh copy from the encoded bytes so the
	// two states never share mutable operation state.
	if b.isolationActive && b.fullState != nil {
		document, err := DecodePartial(raw)
		if err != nil || len(document.Changes) == 0 {
			return [32]byte{}, fmt.Errorf("cannot decode isolated change for full history: %w", err)
		}

		fullChange := document.Changes[0]

		fullChange.Raw = append([]byte(nil), raw...)

		if err := b.fullState.ApplyChange(&fullChange); err != nil {
			return [32]byte{}, err
		}

		if next := b.fullState.maxOpGlobal() + 1; next > b.nextOp {
			b.nextOp = next
		}
	}

	b.appended = append(b.appended, raw)
	b.pending = nil

	return [32]byte(*change.Hash), nil
}

func (b *Engine) EmptyCommit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if err := ctx.Err(); err != nil {
		return [32]byte{}, err
	}

	if len(b.pending) != 0 {
		return [32]byte{}, fmt.Errorf("cannot create empty change with pending operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1

	change := &Change{
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

	raw, err := EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native empty change: %w", err)
	}

	if err := b.state.recordAppliedChange(change); err != nil {
		return [32]byte{}, err
	}

	b.appended = append(b.appended, raw)

	return [32]byte(*change.Hash), nil
}

func (b *Engine) Rollback(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if len(b.pending) == 0 {
		return 0, nil
	}

	data := append([]byte(nil), b.base...)
	for _, change := range b.appended {
		data = append(data, change...)
	}

	document, err := Decode(data)
	if err != nil {
		return 0, fmt.Errorf("cannot decode committed state during rollback: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return 0, fmt.Errorf("cannot restore committed state during rollback: %w", err)
	}

	cancelled := uint64(len(b.pending))
	b.state = state
	b.nextOp = state.maxOpGlobal() + 1
	b.pending = nil
	b.objects = map[uint32]ObjectID{0: RootObject()}
	b.nextHandle = 1

	return cancelled, nil
}
