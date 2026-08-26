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
	"sort"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/encoding"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

type Engine struct {
	state         *State
	actor         opset.ActorID
	nextOp        uint64
	base          []byte
	appended      [][]byte
	saveCursor    int
	pending       []opset.Operation
	objects       map[uint32]opset.ObjectID
	nextHandle    uint32
	syncStates    map[uint32]*syncSessionState
	nextSyncState uint32
	queuedChanges map[opset.ChangeHash]*opset.Change
	queuedBytes   int
	diffCursor    [][32]byte

	// isolation pins reads and writes to a historical frontier. While active,
	// state points at a view built from the isolation heads and fullState keeps
	// the complete history; committed isolated changes are applied to both, while
	// merged changes are applied only to fullState.
	isolationActive bool
	fullState       *State
	baseActor       opset.ActorID

	// isolationDiffTargets records the frontiers isolated to since the diff
	// cursor was last set. When present, an incremental diff replays the
	// transition from the cursor down to each isolation frontier and back up to
	// the current heads, matching the reference's patch-log output across
	// isolate/integrate rather than a direct state comparison.
	isolationDiffTargets [][][32]byte

	// revision increases on every change to the committed history or the
	// retained orphan set. The compacted save is cached against it so repeated
	// saves of an unchanged document skip rebuilding the whole columnar
	// document, which the collaboration snapshot path does on every request.
	revision  uint64
	saveCache saveCacheEntry
}

// saveCacheEntry memoizes one compacted save. It is valid only for the exact
// revision and option combination it was built from.
type saveCacheEntry struct {
	revision      uint64
	retainOrphans bool
	compress      bool
	valid         bool
	data          []byte
}

type syncSessionState struct {
	RemoteHeads       [][32]byte `json:"remoteHeads"`
	LastSentHeads     [][32]byte `json:"lastSentHeads"`
	LastSentNeed      [][32]byte `json:"lastSentNeed"`
	Need              [][32]byte `json:"need"`
	Requested         [][32]byte `json:"requested"`
	NeedsAck          bool       `json:"needsAck"`
	InFlight          bool       `json:"inFlight"`
	Sent              bool       `json:"sent"`
	ReadOnly          bool       `json:"readOnly"`
	PeerReadOnly      bool       `json:"peerReadOnly"`
	PeerModeChanged   bool       `json:"peerModeChanged"`
	PeerSupportsReset bool       `json:"peerSupportsReset"`
	NeedsReset        bool       `json:"needsReset"`
	ModeChanged       bool       `json:"modeChanged"`
}

type scalarWire struct {
	Type   string `json:"type"`
	Bool   bool   `json:"bool"`
	Uint   uint64 `json:"uint"`
	Int    int64  `json:"int"`
	Float  uint64 `json:"floatBits"`
	String string `json:"string"`
	Bytes  string `json:"bytes"`
}

const (
	maxQueuedChangeBytes = 64 * 1024 * 1024
	maxQueuedChanges     = 100_000

	syncFlagReset         = 1 << 0
	syncFlagReadOnly      = 1 << 1
	syncFlagSupportsReset = 1 << 2
	syncFlagMarker        = 0x80
)

func NewEngine() (*Engine, error) {

	actor, err := randomActorID()
	if err != nil {
		return nil, err
	}

	base := encodeEmptyDocument()

	document, err := storage.Decode(base)
	if err != nil {
		return nil, fmt.Errorf("cannot decode native empty document: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native empty state: %w", err)
	}

	return &Engine{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          base,
		objects:       map[uint32]opset.ObjectID{0: opset.RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*syncSessionState),
		nextSyncState: 1,
		queuedChanges: make(map[opset.ChangeHash]*opset.Change),
	}, nil
}

func LoadEngine(data []byte) (*Engine, error) {

	document, err := storage.Decode(data)
	if err != nil {
		// A document may retain orphan changes (changes whose dependencies are
		// not present) that were preserved across a save. Strict decoding
		// rejects them, so fall back to a tolerant load that applies every
		// change whose dependencies are satisfiable and queues the rest. A load
		// that cannot apply a single change (a bare orphan) still fails.
		if engine, ok, tolerantErr := loadEngineRetainingOrphans(data, err); ok {
			return engine, tolerantErr
		}

		return nil, fmt.Errorf("cannot decode native document: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native document state: %w", err)
	}

	actor, err := randomActorID()
	if err != nil {
		return nil, err
	}

	return &Engine{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          append([]byte(nil), data...),
		objects:       map[uint32]opset.ObjectID{0: opset.RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*syncSessionState),
		nextSyncState: 1,
		queuedChanges: make(map[opset.ChangeHash]*opset.Change),
	}, nil
}

// loadEngineRetainingOrphans attempts a tolerant load for documents that carry
// orphan changes. It returns ok=false when the tolerant path does not apply (the
// data is corrupt beyond missing dependencies, or nothing can be applied), so
// the caller reports the original strict error. On success the applied history
// forms the base and the orphan changes are queued for later resolution.
func loadEngineRetainingOrphans(
	data []byte,
	strictErr error,
) (*Engine, bool, error) {
	if !strings.Contains(strictErr.Error(), "missing dependency") {
		return nil, false, nil
	}

	document, err := storage.DecodePartial(data)
	if err != nil {
		return nil, false, nil
	}

	state := NewState()
	queued := make(map[opset.ChangeHash]*opset.Change, len(document.Changes))

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash == nil || len(change.Raw) == 0 {
			return nil, false, nil
		}

		queued[*change.Hash] = change
	}

	applied := make([]*opset.Change, 0, len(document.Changes))

	for {
		progressed := false

		for _, change := range orderedQueuedChanges(queued) {
			if !state.hasDependencies(change) {
				continue
			}

			if err := state.ApplyChange(change); err != nil {
				return nil, false, nil
			}

			applied = append(applied, change)
			delete(queued, *change.Hash)

			progressed = true
		}

		if !progressed {
			break
		}
	}

	if len(applied) == 0 {
		return nil, false, nil
	}

	actor, err := randomActorID()
	if err != nil {
		return nil, true, err
	}

	base := make([]byte, 0, len(data))
	for _, change := range applied {
		base = append(base, change.Raw...)
	}

	queuedClone := make(map[opset.ChangeHash]*opset.Change, len(queued))
	queuedBytes := 0

	for hash, change := range queued {
		clone := *change
		clone.Raw = append([]byte(nil), change.Raw...)
		queuedClone[hash] = &clone
		queuedBytes += len(clone.Raw)
	}

	return &Engine{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          base,
		objects:       map[uint32]opset.ObjectID{0: opset.RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*syncSessionState),
		nextSyncState: 1,
		queuedChanges: queuedClone,
		queuedBytes:   queuedBytes,
	}, true, nil
}

// orderedQueuedChanges returns queued changes in a deterministic order (by hash)
// so tolerant loading applies and re-serializes changes reproducibly.
func orderedQueuedChanges(queued map[opset.ChangeHash]*opset.Change) []*opset.Change {
	changes := make([]*opset.Change, 0, len(queued))
	for _, change := range queued {
		changes = append(changes, change)
	}

	sort.Slice(
		changes,
		func(i, j int) bool {
			return bytes.Compare(changes[i].Hash[:], changes[j].Hash[:]) < 0
		},
	)

	return changes
}

func (b *Engine) Close() error {
	return nil
}

// Save serializes the whole history as one compacted document chunk, the form
// save() produces in the Rust and JavaScript implementations. It replaces the
// change-by-change stream Go used to write, which grew without bound as a
// history accumulated commits. retainOrphans keeps queued changes whose
// dependencies are still missing so they survive a save/load round trip, and
// compress DEFLATEs the document columns and any trailing change chunks.
func (b *Engine) Save(
	retainOrphans bool,
	compress bool,
) ([]byte, error) {
	return b.save(retainOrphans, compress)
}

func (b *Engine) save(
	retainOrphans bool,
	deflate bool,
) ([]byte, error) {
	if len(b.pending) > 0 {
		if _, err := b.Commit("", time.Time{}); err != nil {
			return nil, err
		}
	}

	// Rebuilding the columnar document is by far the costliest part of a save, so
	// an unchanged document returns the bytes built last time. The cache is keyed
	// by the mutation revision and the option combination, and it is invalidated
	// implicitly because every committed change advances the revision.
	if cached, ok := b.cachedSave(retainOrphans, deflate); ok {
		b.saveCursor = len(b.appended)

		return cached, nil
	}

	// A compacted document chunk is the form every other implementation writes
	// and is dramatically smaller than the change stream for a long history. It
	// leaves the incremental cursor at the end, exactly as the stream save did,
	// so a following SaveIncremental still emits only later changes.
	if data, ok, err := b.compact(retainOrphans, deflate); err != nil {
		return nil, err
	} else if ok {
		b.saveCursor = len(b.appended)
		b.storeSave(retainOrphans, deflate, data)

		return data, nil
	}

	data, err := b.streamSave(retainOrphans, deflate)
	if err != nil {
		return nil, err
	}

	b.storeSave(retainOrphans, deflate, data)

	return data, nil
}

// cachedSave returns a copy of the previously built save when it is still valid
// for this revision and option combination. A copy is returned because callers
// own the bytes and may retain or mutate them.
func (b *Engine) cachedSave(retainOrphans, compress bool) ([]byte, bool) {
	if !b.saveCache.valid ||
		b.saveCache.revision != b.revision ||
		b.saveCache.retainOrphans != retainOrphans ||
		b.saveCache.compress != compress {
		return nil, false
	}

	return append([]byte(nil), b.saveCache.data...), true
}

// storeSave records a freshly built save so an unchanged document can return it
// without rebuilding. The stored bytes are copied so a caller mutating the
// returned slice cannot corrupt the cache.
func (b *Engine) storeSave(retainOrphans, compress bool, data []byte) {
	b.saveCache = saveCacheEntry{
		revision:      b.revision,
		retainOrphans: retainOrphans,
		compress:      compress,
		valid:         true,
		data:          append([]byte(nil), data...),
	}
}

// streamSave serializes the history as the loaded base followed by each change
// chunk since. It preserves the loaded bytes verbatim, including columns this
// version does not understand, and is the fallback when a history cannot be
// compacted (while isolated, or when the change graph is inconsistent).
func (b *Engine) streamSave(retainOrphans, deflate bool) ([]byte, error) {
	total := len(b.base)
	for _, change := range b.appended {
		total += len(change)
	}

	data := make([]byte, 0, total)
	data = append(data, b.base...)

	for _, change := range b.appended {
		data = append(data, maybeCompressChangeChunk(change, deflate)...)
	}

	b.saveCursor = len(b.appended)

	if retainOrphans {
		for _, change := range orderedQueuedChanges(b.queuedChanges) {
			data = append(data, maybeCompressChangeChunk(change.Raw, deflate)...)
		}
	}

	return data, nil
}

// deflateMinSize matches Rust's change::DEFLATE_MIN_SIZE: change chunks whose
// body is at least this many bytes are worth compressing.
const deflateMinSize = 250

// maybeCompressChangeChunk reframes an uncompressed change chunk as a compressed
// change chunk when compression is requested and the body is large enough to
// benefit. The 4-byte checksum is preserved because the reference (and native
// decoder) recompute the hash from the inflated body. Any other chunk kind, a
// small body, or a non-shrinking result is returned unchanged.
func maybeCompressChangeChunk(raw []byte, deflateEnabled bool) []byte {
	const headerSize = 9 // 4 magic + 4 checksum + 1 type

	if !deflateEnabled || len(raw) <= headerSize || opset.ChunkType(raw[8]) != opset.ChunkChange {
		return raw
	}

	reader := encoding.NewReaderAt(raw, headerSize)

	bodyLength, err := reader.ULEB()
	if err != nil || reader.Offset()+int(bodyLength) > len(raw) {
		return raw
	}

	body := raw[reader.Offset() : reader.Offset()+int(bodyLength)]
	if len(body) < deflateMinSize {
		return raw
	}

	compressed, err := storage.Deflate(body)
	if err != nil || len(compressed) >= len(body) {
		return raw
	}

	out := make([]byte, 0, headerSize+len(compressed)+8)
	out = append(out, raw[:8]...)
	out = append(out, byte(opset.ChunkCompressedChange))
	out = encoding.AppendULEB(out, uint64(len(compressed)))
	out = append(out, compressed...)

	return out
}

func (b *Engine) SaveIncremental() ([]byte, error) {
	if len(b.pending) > 0 {
		if _, err := b.Commit("", time.Time{}); err != nil {
			return nil, err
		}
	}

	if b.saveCursor > len(b.appended) {
		b.saveCursor = len(b.appended)
	}

	total := 0
	for _, change := range b.appended[b.saveCursor:] {
		total += len(change)
	}

	data := make([]byte, 0, total)
	for _, change := range b.appended[b.saveCursor:] {
		data = append(data, change...)
	}

	b.saveCursor = len(b.appended)

	return data, nil
}

func (b *Engine) LoadIncremental(
	data []byte,
) (uint64, error) {
	_, consumed, err := storage.DecodeIncremental(data)
	if err != nil {
		return 0, err
	}

	before := len(b.state.changes)
	if _, err := b.Merge(data[:consumed]); err != nil {
		return 0, err
	}

	after := len(b.state.changes)
	if after < before {
		return 0, fmt.Errorf("incremental load reduced the change count")
	}

	return uint64(after - before), nil
}

func (b *Engine) SetActor(value []byte) error {

	actor, err := opset.NewActorID(value)
	if err != nil {
		return err
	}

	if len(b.pending) > 0 {
		return fmt.Errorf("cannot change actor with pending operations")
	}

	b.actor = actor

	return nil
}
