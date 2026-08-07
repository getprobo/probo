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

// Package native adapts a pure-Go Automerge engine to Probo's backend contract.
package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
	"unicode/utf16"

	goautomerge "github.com/develerltd/go-automerge/automerge"
)

type (
	Backend struct {
		document  *goautomerge.Doc
		objects   []goautomerge.ObjId
		syncState []*syncState
		pending   bool
		closed    bool
	}

	syncState struct {
		LastSent [][32]byte `json:"lastSent,omitempty"`
	}

	syncMessage struct {
		Document []byte `json:"document"`
	}

	plainSpan struct {
		Type  string         `json:"type"`
		Value string         `json:"value"`
		Marks map[string]any `json:"marks,omitempty"`
	}
)

var (
	ErrClosed              = errors.New("native Automerge backend is closed")
	ErrNoOperations        = errors.New("change contains no operations")
	ErrNegativeDeleteCount = errors.New("negative text delete count is not supported")
)

func New(_ context.Context) (*Backend, error) {
	return &Backend{objects: []goautomerge.ObjId{goautomerge.Root}}, nil
}

func Load(ctx context.Context, data []byte) (*Backend, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("cannot load native Automerge document: %w", err)
	}

	document, err := goautomerge.Load(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode native Automerge document: %w", err)
	}

	return &Backend{
		document: document,
		objects:  []goautomerge.ObjId{goautomerge.Root},
	}, nil
}

func (b *Backend) Close(_ context.Context) error {
	b.closed = true
	b.document = nil
	b.objects = nil
	b.syncState = nil
	return nil
}

func (b *Backend) Save(ctx context.Context) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	data, err := b.document.Save()
	if err != nil {
		return nil, fmt.Errorf("cannot encode native Automerge document: %w", err)
	}

	return data, nil
}

func (b *Backend) SetActor(ctx context.Context, actor []byte) error {
	if err := b.open(ctx); err != nil {
		return err
	}
	if len(actor) == 0 {
		return fmt.Errorf("actor ID cannot be empty")
	}

	actorID := goautomerge.ActorId(bytes.Clone(actor))
	if b.document == nil {
		b.document = goautomerge.NewWithActorId(actorID)
		return nil
	}

	b.document = b.document.ForkWithActorId(actorID)
	return nil
}

func (b *Backend) PutString(ctx context.Context, object uint32, key, value string) error {
	if err := b.ready(ctx); err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return fmt.Errorf("cannot resolve map object: %w", err)
	}
	if err := b.document.Put(objectID, key, goautomerge.NewStr(value)); err != nil {
		return fmt.Errorf("cannot put native map value: %w", err)
	}
	b.pending = true

	return nil
}

func (b *Backend) PutText(ctx context.Context, object uint32, key string) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve text parent: %w", err)
	}
	textID, err := b.document.PutObject(objectID, key, goautomerge.ObjTypeText)
	if err != nil {
		return 0, fmt.Errorf("cannot put native text object: %w", err)
	}
	b.pending = true

	return b.pushObject(textID)
}

func (b *Backend) GetText(ctx context.Context, object uint32, key string) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve text parent: %w", err)
	}
	value, externalID, err := b.document.Get(objectID, goautomerge.MapProp(key))
	if err != nil {
		return 0, fmt.Errorf("cannot get native text object: %w", err)
	}
	if !value.IsObject || value.ObjType != goautomerge.ObjTypeText {
		return 0, fmt.Errorf("value is not text")
	}

	return b.pushObject(b.objectID(externalID))
}

func (b *Backend) SpliceText(
	ctx context.Context,
	object uint32,
	index uint32,
	deleteCount int32,
	value string,
) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if deleteCount < 0 {
		return ErrNegativeDeleteCount
	}

	objectID, err := b.object(object)
	if err != nil {
		return fmt.Errorf("cannot resolve text object: %w", err)
	}
	current, err := b.document.Text(objectID)
	if err != nil {
		return fmt.Errorf("cannot read native text before splice: %w", err)
	}
	position, err := runePosition(current, uint64(index))
	if err != nil {
		return fmt.Errorf("cannot resolve native text splice index: %w", err)
	}
	end, err := runePosition(current, uint64(index)+uint64(deleteCount))
	if err != nil {
		return fmt.Errorf("cannot resolve native text splice end: %w", err)
	}
	if err := b.document.SpliceText(objectID, position, end-position, ""); err != nil {
		return fmt.Errorf("cannot delete native text: %w", err)
	}
	for _, character := range value {
		if err := b.document.SpliceText(objectID, position, 0, string(character)); err != nil {
			return fmt.Errorf("cannot insert native text: %w", err)
		}
		position++
	}
	b.pending = b.pending || deleteCount > 0 || value != ""

	return nil
}

func (b *Backend) Text(ctx context.Context, object uint32) (string, error) {
	if err := b.ready(ctx); err != nil {
		return "", err
	}

	objectID, err := b.object(object)
	if err != nil {
		return "", fmt.Errorf("cannot resolve text object: %w", err)
	}
	text, err := b.document.Text(objectID)
	if err != nil {
		return "", fmt.Errorf("cannot read native text: %w", err)
	}

	return text, nil
}

func (b *Backend) TextSpans(ctx context.Context, object uint32) ([]byte, error) {
	text, err := b.Text(ctx, object)
	if err != nil {
		return nil, fmt.Errorf("cannot read native text spans: %w", err)
	}

	data, err := json.Marshal([]plainSpan{{Type: "text", Value: text}})
	if err != nil {
		return nil, fmt.Errorf("cannot encode native text spans: %w", err)
	}

	return data, nil
}

func (b *Backend) TextCursor(ctx context.Context, object, index uint32) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve cursor text object: %w", err)
	}
	current, err := b.document.Text(objectID)
	if err != nil {
		return nil, fmt.Errorf("cannot read native cursor text: %w", err)
	}
	position, err := runePosition(current, uint64(index))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve native cursor index: %w", err)
	}

	var cursor goautomerge.Cursor
	if position == uint64(len([]rune(current))) {
		cursor = goautomerge.EndCursor()
	} else {
		cursor, err = b.document.GetCursor(objectID, position, goautomerge.MoveAfter)
		if err != nil {
			return nil, fmt.Errorf("cannot create native cursor: %w", err)
		}
	}

	return cursor.ToBytes(), nil
}

func (b *Backend) TextCursorPosition(
	ctx context.Context,
	object uint32,
	data []byte,
) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve cursor text object: %w", err)
	}
	cursor, err := goautomerge.CursorFromBytes(data)
	if err != nil {
		return 0, fmt.Errorf("cannot decode native cursor: %w", err)
	}
	position, err := b.document.GetCursorPosition(objectID, cursor)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve native cursor: %w", err)
	}
	current, err := b.document.Text(objectID)
	if err != nil {
		return 0, fmt.Errorf("cannot read native cursor text: %w", err)
	}
	index, err := utf16Position(current, position.Index)
	if err != nil {
		return 0, fmt.Errorf("cannot convert native cursor position: %w", err)
	}
	if index > math.MaxUint32 {
		return 0, fmt.Errorf("native cursor position %d exceeds uint32", index)
	}

	return uint32(index), nil
}

func (b *Backend) Commit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return [32]byte{}, err
	}
	if !b.pending {
		return [32]byte{}, ErrNoOperations
	}

	b.document.Commit(message, timestamp.Unix())
	b.pending = false
	heads := b.document.Heads()
	if len(heads) != 1 {
		return [32]byte{}, fmt.Errorf("native commit produced %d heads", len(heads))
	}

	return [32]byte(heads[0]), nil
}

func (b *Backend) Heads(ctx context.Context) ([][32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	return copyHeads(b.document.Heads()), nil
}

func (b *Backend) Merge(ctx context.Context, other []byte) ([][32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	otherDocument, err := goautomerge.Load(other)
	if err != nil {
		return nil, fmt.Errorf("cannot load native merge source: %w", err)
	}
	if err := b.document.Merge(otherDocument); err != nil {
		return nil, fmt.Errorf("cannot merge native document: %w", err)
	}
	b.pending = false

	return copyHeads(b.document.Heads()), nil
}

func (b *Backend) NewSyncState(ctx context.Context) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}
	if len(b.syncState) == math.MaxUint32 {
		return 0, fmt.Errorf("too many native sync states")
	}

	handle := uint32(len(b.syncState))
	b.syncState = append(b.syncState, &syncState{})
	return handle, nil
}

func (b *Backend) CloseSyncState(ctx context.Context, handle uint32) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if _, err := b.state(handle); err != nil {
		return err
	}

	b.syncState[handle] = nil
	return nil
}

func (b *Backend) GenerateSyncMessage(
	ctx context.Context,
	handle uint32,
) ([]byte, bool, error) {
	if err := b.ready(ctx); err != nil {
		return nil, false, err
	}

	state, err := b.state(handle)
	if err != nil {
		return nil, false, err
	}
	heads := copyHeads(b.document.Heads())
	if equalHeads(heads, state.LastSent) {
		return nil, false, nil
	}
	document, err := b.document.Save()
	if err != nil {
		return nil, false, fmt.Errorf("cannot save native sync document: %w", err)
	}
	message, err := json.Marshal(syncMessage{Document: document})
	if err != nil {
		return nil, false, fmt.Errorf("cannot encode native sync message: %w", err)
	}
	state.LastSent = heads

	return message, true, nil
}

func (b *Backend) ReceiveSyncMessage(ctx context.Context, handle uint32, data []byte) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if _, err := b.state(handle); err != nil {
		return err
	}

	var message syncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("cannot decode native sync message: %w", err)
	}
	other, err := goautomerge.Load(message.Document)
	if err != nil {
		return fmt.Errorf("cannot load native sync document: %w", err)
	}
	if err := b.document.Merge(other); err != nil {
		return fmt.Errorf("cannot merge native sync document: %w", err)
	}

	return nil
}

func (b *Backend) SaveSyncState(ctx context.Context, handle uint32) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	state, err := b.state(handle)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native sync state: %w", err)
	}

	return data, nil
}

func (b *Backend) LoadSyncState(ctx context.Context, data []byte) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	var state syncState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, fmt.Errorf("cannot decode native sync state: %w", err)
	}
	if len(b.syncState) == math.MaxUint32 {
		return 0, fmt.Errorf("too many native sync states")
	}
	handle := uint32(len(b.syncState))
	b.syncState = append(b.syncState, &state)

	return handle, nil
}

func (b *Backend) ready(ctx context.Context) error {
	if err := b.open(ctx); err != nil {
		return err
	}
	if b.document == nil {
		return fmt.Errorf("native Automerge actor is not initialized")
	}
	return nil
}

func (b *Backend) open(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("native Automerge operation canceled: %w", err)
	}
	if b.closed {
		return ErrClosed
	}
	return nil
}

func (b *Backend) object(handle uint32) (goautomerge.ObjId, error) {
	if int(handle) >= len(b.objects) {
		return goautomerge.ObjId{}, fmt.Errorf("invalid object handle %d", handle)
	}
	return b.objects[handle], nil
}

func (b *Backend) pushObject(object goautomerge.ObjId) (uint32, error) {
	if len(b.objects) == math.MaxUint32 {
		return 0, fmt.Errorf("too many object handles")
	}
	handle := uint32(len(b.objects))
	b.objects = append(b.objects, object)
	return handle, nil
}

func (b *Backend) objectID(externalID goautomerge.ExId) goautomerge.ObjId {
	if externalID.IsRoot {
		return goautomerge.Root
	}
	actors := b.document.Actors()
	for index, actor := range actors {
		if actor.Compare(externalID.Actor) == 0 {
			return goautomerge.ObjId{
				OpId: goautomerge.OpId{
					Counter:  externalID.Counter,
					ActorIdx: uint32(index),
				},
			}
		}
	}
	return goautomerge.ObjId{
		OpId: goautomerge.OpId{
			Counter:  externalID.Counter,
			ActorIdx: externalID.ActorIdx,
		},
	}
}

func (b *Backend) state(handle uint32) (*syncState, error) {
	if int(handle) >= len(b.syncState) || b.syncState[handle] == nil {
		return nil, fmt.Errorf("invalid sync state handle %d", handle)
	}
	return b.syncState[handle], nil
}

func runePosition(value string, utf16Index uint64) (uint64, error) {
	var units uint64
	var runes uint64
	for _, character := range value {
		if units == utf16Index {
			return runes, nil
		}
		width := uint64(1)
		if utf16.RuneLen(character) == 2 {
			width = 2
		}
		if utf16Index < units+width {
			return 0, fmt.Errorf("UTF-16 index %d splits a surrogate pair", utf16Index)
		}
		units += width
		runes++
	}
	if units == utf16Index {
		return runes, nil
	}
	return 0, fmt.Errorf("UTF-16 index %d exceeds text length %d", utf16Index, units)
}

func utf16Position(value string, runeIndex uint64) (uint64, error) {
	var units uint64
	var runes uint64
	for _, character := range value {
		if runes == runeIndex {
			return units, nil
		}
		if utf16.RuneLen(character) == 2 {
			units += 2
		} else {
			units++
		}
		runes++
	}
	if runes == runeIndex {
		return units, nil
	}
	return 0, fmt.Errorf("rune index %d exceeds text length %d", runeIndex, runes)
}

func copyHeads(source []goautomerge.ChangeHash) [][32]byte {
	heads := make([][32]byte, len(source))
	for index := range source {
		heads[index] = [32]byte(source[index])
	}
	sort.Slice(
		heads,
		func(left, right int) bool {
			return bytes.Compare(heads[left][:], heads[right][:]) < 0
		},
	)
	return heads
}

func equalHeads(left, right [][32]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
