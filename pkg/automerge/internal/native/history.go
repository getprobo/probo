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
	"sort"
)

func (b *Engine) Heads(ctx context.Context) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	heads := b.state.Heads()

	result := make([][32]byte, len(heads))
	for i := range heads {
		result[i] = [32]byte(heads[i])
	}

	return result, nil
}

func (b *Engine) HasHeads(
	ctx context.Context,
	heads [][32]byte,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	for _, head := range heads {
		if !b.state.hasChange(ChangeHash(head)) {
			return false, nil
		}
	}

	return true, nil
}

func (b *Engine) MissingDependencies(
	ctx context.Context,
	heads [][32]byte,
) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	missing := make(map[[32]byte]struct{})

	for _, head := range heads {
		_, queued := b.queuedChanges[ChangeHash(head)]
		if !b.state.hasChange(ChangeHash(head)) && !queued {
			missing[head] = struct{}{}
		}
	}

	for _, change := range b.queuedChanges {
		for _, dependency := range change.Dependencies {
			if !b.state.hasChange(dependency) {
				missing[[32]byte(dependency)] = struct{}{}
			}
		}
	}

	result := make([][32]byte, 0, len(missing))
	for dependency := range missing {
		result = append(result, dependency)
	}

	sort.Slice(result, func(i, j int) bool {
		return bytes.Compare(result[i][:], result[j][:]) < 0
	})

	return result, nil
}

func (b *Engine) ChangesSince(
	ctx context.Context,
	heads [][32]byte,
) ([][]byte, [][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	knownHeads := make([]ChangeHash, len(heads))
	for i, head := range heads {
		knownHeads[i] = ChangeHash(head)
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
		if change.Hash == nil {
			return nil, nil, fmt.Errorf("change %d has no hash", i)
		}

		raw[i] = append([]byte(nil), change.Raw...)
		hashes[i] = [32]byte(*change.Hash)
	}

	return raw, hashes, nil
}

func (b *Engine) ApplyChanges(
	ctx context.Context,
	changes [][]byte,
) error {
	for i, change := range changes {
		if _, err := b.Merge(ctx, change); err != nil {
			return fmt.Errorf("cannot apply native change %d: %w", i, err)
		}
	}

	return nil
}

func (b *Engine) Merge(ctx context.Context, data []byte) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// While isolated, merged changes belong to the full history rather than the
	// pinned view, so operate on the full state and keep the pinned view intact.
	if b.isolationActive && b.fullState != nil {
		b.state, b.fullState = b.fullState, b.state

		defer func() {
			b.state, b.fullState = b.fullState, b.state
			b.nextOp = b.fullState.maxOpGlobal() + 1
		}()
	}

	document, err := Decode(data)
	if err != nil {
		document, err = DecodePartial(data)
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

		b.base = append([]byte(nil), data...)
		b.appended = nil
		b.saveCursor = 0

		return b.Heads(ctx)
	}

	if b.requiresSnapshotMerge(document) {
		if err := b.mergeDocumentSnapshot(data, document); err != nil {
			return nil, err
		}

		return b.Heads(ctx)
	}

	if err := b.applyMergedChanges(document.Changes); err != nil {
		return nil, err
	}

	if next := b.state.maxOpGlobal() + 1; next > b.nextOp {
		b.nextOp = next
	}

	return b.Heads(ctx)
}

func (b *Engine) requiresSnapshotMerge(document *Document) bool {
	if len(document.ChunkTypes) == 0 ||
		document.ChunkTypes[0] != ChunkDocument {
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
	document *Document,
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
	b.queuedChanges = make(map[ChangeHash]*Change)
	b.queuedBytes = 0
	b.nextOp = state.maxOpGlobal() + 1

	return nil
}

func documentChangeByActorSequence(
	document *Document,
	actor ActorID,
	sequence uint64,
) *Change {
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Actor == actor && change.Sequence == sequence {
			return change
		}
	}

	return nil
}

func (b *Engine) applyMergedChanges(changes []Change) error {
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

	return nil
}
