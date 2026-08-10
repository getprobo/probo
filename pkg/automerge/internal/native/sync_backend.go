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
	"encoding/json"
	"fmt"
	"sort"
)

func (b *Backend) NewSyncState(ctx context.Context) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	handle := b.nextSyncState
	b.nextSyncState++
	b.syncStates[handle] = &nativeSyncState{}

	return handle, nil
}

func (b *Backend) CloseSyncState(ctx context.Context, handle uint32) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, ok := b.syncStates[handle]; !ok {
		return fmt.Errorf("invalid sync state %d", handle)
	}

	delete(b.syncStates, handle)

	return nil
}

func (b *Backend) SetSyncReadOnly(
	ctx context.Context,
	handle uint32,
	readOnly bool,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return err
	}

	if state.ReadOnly == readOnly {
		return nil
	}

	if state.ReadOnly && !readOnly {
		peerSupportsReset := state.PeerSupportsReset
		*state = nativeSyncState{
			PeerSupportsReset: peerSupportsReset,
			NeedsReset:        true,
			ModeChanged:       true,
		}
	} else {
		state.ReadOnly = true
		state.InFlight = false
		state.ModeChanged = true
	}

	return nil
}

func (b *Backend) SyncPeerReadOnly(
	ctx context.Context,
	handle uint32,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return false, err
	}

	return state.PeerReadOnly, nil
}

func (b *Backend) GenerateSyncMessage(
	ctx context.Context,
	handle uint32,
) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return nil, false, err
	}

	heads, err := b.Heads(ctx)
	if err != nil {
		return nil, false, err
	}

	// A message is only truly in flight while we have nothing new to say. New
	// local changes (heads advanced past the last sent frontier) must be sent
	// even while a previous message awaits acknowledgement, matching upstream
	// Rust, which never withholds local changes during synchronization. A Need
	// that is unchanged since the last message we sent is not new information:
	// re-requesting the same missing dependencies (for example an orphan change
	// whose base never arrives) would otherwise regenerate an identical message
	// forever and never quiesce.
	if state.InFlight &&
		!state.ModeChanged &&
		!state.NeedsReset &&
		len(state.Requested) == 0 &&
		equalHashes(state.Need, state.LastSentNeed) &&
		equalHashes(heads, state.LastSentHeads) {
		return nil, false, nil
	}

	if state.ModeChanged || state.NeedsReset {
		state.InFlight = false
	}

	// The first message for a sync state is always sent so the peer learns our
	// heads and capabilities, matching upstream Rust's first_response_is_some
	// behavior. Subsequent messages may be suppressed when nothing is pending.
	if state.Sent {
		if state.PeerReadOnly &&
			!state.PeerModeChanged &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			!state.NeedsAck &&
			len(state.Requested) == 0 &&
			equalHashes(state.Need, state.LastSentNeed) &&
			equalHashes(heads, state.LastSentHeads) {
			return nil, false, nil
		}

		if state.ReadOnly &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			!state.NeedsAck &&
			len(state.Requested) == 0 &&
			equalHashes(state.Need, state.LastSentNeed) &&
			equalHashes(heads, state.LastSentHeads) {
			return nil, false, nil
		}

		if !state.NeedsAck &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			len(state.Requested) == 0 &&
			equalHashes(state.Need, state.LastSentNeed) &&
			equalHashes(heads, state.RemoteHeads) {
			return nil, false, nil
		}
	}

	flags := byte(syncFlagSupportsReset)
	if state.ReadOnly {
		flags |= syncFlagReadOnly
	}

	messageHeads := heads

	if state.NeedsReset {
		if state.PeerSupportsReset {
			flags |= syncFlagReset
		} else {
			messageHeads = nil
		}
	}

	message := SyncMessage{
		Version: SyncMessageVersion2,
		Heads:   messageHeads,
		Need:    append([][32]byte(nil), state.Need...),
		Flags:   []byte{2, syncFlagMarker | flags},
	}
	if !state.NeedsAck && !state.PeerReadOnly {
		switch {
		case len(state.Requested) > 0:
			for _, requested := range state.Requested {
				change, ok := b.state.changes[ChangeHash(requested)]
				if !ok || len(change.Raw) == 0 {
					continue
				}

				message.Changes = append(
					message.Changes,
					append([]byte(nil), change.Raw...),
				)
			}

			state.Requested = nil
		case len(state.Need) == 0:
			remoteHeads := make([]ChangeHash, len(state.RemoteHeads))
			for i, head := range state.RemoteHeads {
				remoteHeads[i] = ChangeHash(head)
			}

			changes, incremental := b.state.changesSince(remoteHeads)
			if incremental {
				for _, change := range changes {
					message.Changes = append(
						message.Changes,
						append([]byte(nil), change.Raw...),
					)
				}
			} else {
				document, err := b.Save(ctx)
				if err != nil {
					return nil, false, err
				}

				message.Changes = [][]byte{document}
			}
		}
	}

	if !state.NeedsAck {
		state.InFlight = true
	}

	state.NeedsAck = false
	state.Sent = true
	state.LastSentHeads = append(state.LastSentHeads[:0], heads...)
	state.LastSentNeed = append(state.LastSentNeed[:0], state.Need...)
	state.PeerModeChanged = false
	state.ModeChanged = false
	state.NeedsReset = false

	data, err := message.Encode()
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (b *Backend) ReceiveSyncMessage(
	ctx context.Context,
	handle uint32,
	data []byte,
) error {
	state, err := b.syncState(handle)
	if err != nil {
		return err
	}

	message, err := ParseSyncMessage(data)
	if err != nil {
		return err
	}

	state.InFlight = false

	flags := syncMessageFlagBits(message.Flags)

	peerReadOnly := flags&syncFlagReadOnly != 0
	if peerReadOnly != state.PeerReadOnly {
		state.PeerModeChanged = true
	}

	state.PeerReadOnly = peerReadOnly

	state.PeerSupportsReset = flags&syncFlagSupportsReset != 0
	if flags&syncFlagReset != 0 {
		state.RemoteHeads = nil
		state.Requested = nil
	}

	if !state.ReadOnly {
		for _, change := range message.Changes {
			if _, err := b.Merge(ctx, change); err != nil {
				return fmt.Errorf("cannot merge native sync payload: %w", err)
			}
		}
	}

	state.RemoteHeads = append([][32]byte(nil), message.Heads...)
	if state.PeerReadOnly {
		// A read-only peer cannot receive changes. Retaining a Need it sent
		// before (or together with) the mode transition makes every generation
		// attempt to service an impossible request and prevents quiescence. When
		// the peer becomes writable it will advertise the missing heads again.
		state.Requested = nil
	} else {
		state.Requested = append(state.Requested[:0], message.Need...)
	}

	needed := make(map[[32]byte]struct{})

	if !state.ReadOnly {
		for _, head := range message.Heads {
			if _, ok := b.state.changes[ChangeHash(head)]; !ok {
				needed[head] = struct{}{}
			}
		}

		for _, change := range b.queuedChanges {
			for _, dependency := range change.Dependencies {
				if !b.state.hasChange(dependency) {
					needed[[32]byte(dependency)] = struct{}{}
				}
			}
		}
	}

	state.Need = state.Need[:0]
	for dependency := range needed {
		state.Need = append(state.Need, dependency)
	}

	sort.Slice(state.Need, func(i, j int) bool {
		return bytes.Compare(state.Need[i][:], state.Need[j][:]) < 0
	})

	state.NeedsAck = len(message.Changes) > 0 || state.PeerModeChanged

	return nil
}

func (b *Backend) SaveSyncState(
	ctx context.Context,
	handle uint32,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native sync state: %w", err)
	}

	return data, nil
}

func (b *Backend) LoadSyncState(
	ctx context.Context,
	data []byte,
) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var state nativeSyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, fmt.Errorf("cannot decode native sync state: %w", err)
	}

	// A serialized state cannot retain an in-flight transport message. Allow
	// the restored session to regenerate it instead of waiting forever for an
	// acknowledgement that may have been lost with the previous process.
	state.InFlight = false

	handle := b.nextSyncState
	b.nextSyncState++
	b.syncStates[handle] = &state

	return handle, nil
}
