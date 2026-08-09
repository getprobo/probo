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

package automerge

import (
	"context"
	"encoding/json"
	"fmt"
)

type (
	// PatchActionType identifies the kind of change a patch represents.
	PatchActionType string

	// PatchValue is the value carried by a put or insert patch. Object holds a
	// composite object type when the value is an object; otherwise Scalar holds
	// the scalar value.
	PatchValue struct {
		Scalar   *Scalar
		Object   ObjectType
		ObjectID string
	}

	// InsertedValue is one value produced by an insert patch.
	InsertedValue struct {
		Value    PatchValue
		Conflict bool
	}

	// Patch describes a single change to a document's materialized value.
	Patch struct {
		Object   string
		Action   PatchActionType
		Key      string
		Index    uint64
		Length   uint64
		Value    PatchValue
		Values   []InsertedValue
		Text     string
		Delta    int64
		Conflict bool
		Marks    []Mark
	}
)

const (
	PatchPutMap     PatchActionType = "put_map"
	PatchPutSeq     PatchActionType = "put_seq"
	PatchInsert     PatchActionType = "insert"
	PatchSpliceText PatchActionType = "splice_text"
	PatchIncrement  PatchActionType = "increment"
	PatchConflict   PatchActionType = "conflict"
	PatchDeleteMap  PatchActionType = "delete_map"
	PatchDeleteSeq  PatchActionType = "delete_seq"
	PatchMark       PatchActionType = "mark"
)

type (
	encodedPatch struct {
		Obj    string             `json:"obj"`
		Action encodedPatchAction `json:"action"`
	}

	encodedPatchAction struct {
		Type     string                `json:"type"`
		Key      string                `json:"key"`
		Index    uint64                `json:"index"`
		Length   uint64                `json:"length"`
		Value    *encodedPatchValue    `json:"value"`
		Values   []encodedInsertsValue `json:"values"`
		Text     string                `json:"text,omitempty"`
		Delta    int64                 `json:"delta"`
		Conflict bool                  `json:"conflict"`
		Prop     *encodedPatchProp     `json:"prop"`
		Marks    []encodedMark         `json:"marks"`
	}

	encodedInsertsValue struct {
		Value    encodedPatchValue `json:"value"`
		Conflict bool              `json:"conflict"`
	}

	encodedPatchValue struct {
		Scalar json.RawMessage `json:"scalar"`
		Object string          `json:"object"`
		ID     string          `json:"id"`
	}

	encodedPatchProp struct {
		Key   *string `json:"key"`
		Index *uint64 `json:"index"`
	}
)

// CurrentState returns the patches that materialize the document's current
// value from an empty document.
func (d *Document) CurrentState(ctx context.Context) ([]Patch, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.backend.CurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge current state: %w", err)
	}

	return decodePatches(data)
}

// UpdateDiffCursor records the current heads as the incremental diff cursor so a
// following DiffIncremental reports only the changes committed since this call.
func (d *Document) UpdateDiffCursor(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.backend.UpdateDiffCursor(ctx); err != nil {
		return fmt.Errorf("cannot update Automerge diff cursor: %w", err)
	}

	return nil
}

// DiffIncremental returns the patches for the changes committed since the diff
// cursor and advances the cursor to the current heads. It mirrors the Rust
// AutoCommit::diff_incremental helper, reporting operations from the recorded
// patch log so an in-place text replacement is a put rather than a splice.
func (d *Document) DiffIncremental(ctx context.Context) ([]Patch, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.backend.DiffIncremental(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot compute Automerge incremental diff: %w", err)
	}

	return decodePatches(data)
}

// Diff returns the patches that transform the document state at the before
// heads into the state at the after heads.
func (d *Document) Diff(
	ctx context.Context,
	before []Hash,
	after []Hash,
) ([]Patch, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.backend.Diff(ctx, backendHashes(before), backendHashes(after))
	if err != nil {
		return nil, fmt.Errorf("cannot diff Automerge document: %w", err)
	}

	return decodePatches(data)
}

func decodePatches(data []byte) ([]Patch, error) {
	var encoded []encodedPatch
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, fmt.Errorf("cannot decode Automerge patches: %w", err)
	}

	patches := make([]Patch, len(encoded))
	for i, source := range encoded {
		patch := Patch{
			Object:   source.Obj,
			Action:   PatchActionType(source.Action.Type),
			Key:      source.Action.Key,
			Index:    source.Action.Index,
			Length:   source.Action.Length,
			Text:     source.Action.Text,
			Delta:    source.Action.Delta,
			Conflict: source.Action.Conflict,
		}

		if source.Action.Value != nil {
			value, err := decodePatchValue(*source.Action.Value)
			if err != nil {
				return nil, err
			}

			patch.Value = value
		}

		for _, inserted := range source.Action.Values {
			value, err := decodePatchValue(inserted.Value)
			if err != nil {
				return nil, err
			}

			patch.Values = append(patch.Values, InsertedValue{
				Value:    value,
				Conflict: inserted.Conflict,
			})
		}

		if source.Action.Prop != nil {
			if source.Action.Prop.Key != nil {
				patch.Key = *source.Action.Prop.Key
			}

			if source.Action.Prop.Index != nil {
				patch.Index = *source.Action.Prop.Index
			}
		}

		for _, mark := range source.Action.Marks {
			value, err := decodeScalarWire(mark.Value)
			if err != nil {
				return nil, err
			}

			patch.Marks = append(patch.Marks, Mark{
				Start: mark.Start,
				End:   mark.End,
				Name:  mark.Name,
				Value: value,
			})
		}

		patches[i] = patch
	}

	return patches, nil
}

func decodePatchValue(source encodedPatchValue) (PatchValue, error) {
	if source.Object != "" {
		return PatchValue{
			Object:   ObjectType(source.Object),
			ObjectID: source.ID,
		}, nil
	}

	scalar, err := decodeScalarWire(source.Scalar)
	if err != nil {
		return PatchValue{}, fmt.Errorf("cannot decode patch scalar: %w", err)
	}

	return PatchValue{Scalar: &scalar}, nil
}
