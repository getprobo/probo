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
	"fmt"
)

type (
	// ObjectType identifies an Automerge composite value.
	ObjectType string

	// Object is a map, list, text, or table inside a document.
	Object struct {
		document *Document
		handle   uint32
		Type     ObjectType
	}
)

const (
	ObjectTypeMap   ObjectType = "map"
	ObjectTypeList  ObjectType = "list"
	ObjectTypeText  ObjectType = "text"
	ObjectTypeTable ObjectType = "table"
)

// Root returns the document's root map.
func (d *Document) Root() *Object {
	return &Object{
		document: d,
		handle:   rootObject,
		Type:     ObjectTypeMap,
	}
}

// CreateObject assigns a new composite value to a map property.
func (o *Object) CreateObject(
	ctx context.Context,
	key string,
	objectType ObjectType,
) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	handle, err := o.document.backend.PutObject(
		ctx,
		o.handle,
		key,
		string(objectType),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge object: %w", err)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}

// Object returns a composite value from a map property.
func (o *Object) Object(ctx context.Context, key string) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	handle, rawType, err := o.document.backend.GetObject(ctx, o.handle, key)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge object: %w", err)
	}

	objectType := ObjectType(rawType)
	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}

// PutScalar assigns a typed scalar to a map property.
func (o *Object) PutScalar(ctx context.Context, key string, value Scalar) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge scalar: %w", err)
	}

	if err := o.document.backend.PutScalar(ctx, o.handle, key, encoded); err != nil {
		return fmt.Errorf("cannot put Automerge scalar: %w", err)
	}

	return nil
}

// Scalar returns a typed scalar from a map property.
func (o *Object) Scalar(ctx context.Context, key string) (Scalar, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return Scalar{}, ErrClosed
	}

	encoded, err := o.document.backend.GetScalar(ctx, o.handle, key)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot get Automerge scalar: %w", err)
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot decode Automerge scalar: %w", err)
	}

	return value, nil
}

// ScalarAtHeads returns a map scalar at a historical causal frontier.
func (o *Object) ScalarAtHeads(
	ctx context.Context,
	key string,
	heads []Hash,
) (Scalar, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return Scalar{}, ErrClosed
	}

	encoded, err := o.document.backend.GetScalarAtHeads(
		ctx,
		o.handle,
		key,
		backendHashes(heads),
	)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot get historical Automerge scalar: %w", err)
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot decode historical Automerge scalar: %w", err)
	}

	return value, nil
}

// Scalars returns every concurrent scalar value at a map property.
func (o *Object) Scalars(ctx context.Context, key string) ([]Scalar, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	encoded, err := o.document.backend.GetAllScalars(
		ctx,
		o.handle,
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge scalar conflicts: %w", err)
	}

	values, err := decodeScalarWires(encoded)
	if err != nil {
		return nil, fmt.Errorf("cannot decode Automerge scalar conflicts: %w", err)
	}

	return values, nil
}

// InsertScalar inserts a typed scalar at a list index.
func (o *Object) InsertScalar(
	ctx context.Context,
	index uint64,
	value Scalar,
) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge scalar: %w", err)
	}

	if err := o.document.backend.InsertScalar(
		ctx,
		o.handle,
		index,
		encoded,
	); err != nil {
		return fmt.Errorf("cannot insert Automerge scalar: %w", err)
	}

	return nil
}

// PutScalarAt replaces a list element with a typed scalar.
func (o *Object) PutScalarAt(
	ctx context.Context,
	index uint64,
	value Scalar,
) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge scalar: %w", err)
	}

	if err := o.document.backend.PutScalarAt(
		ctx,
		o.handle,
		index,
		encoded,
	); err != nil {
		return fmt.Errorf("cannot replace Automerge scalar: %w", err)
	}

	return nil
}

// InsertObject inserts a new composite value at a list index.
func (o *Object) InsertObject(
	ctx context.Context,
	index uint64,
	objectType ObjectType,
) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	handle, err := o.document.backend.InsertObject(
		ctx,
		o.handle,
		index,
		string(objectType),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot insert Automerge object: %w", err)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}

// PutObjectAt replaces a list element with a new composite value.
func (o *Object) PutObjectAt(
	ctx context.Context,
	index uint64,
	objectType ObjectType,
) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	handle, err := o.document.backend.PutObjectAt(
		ctx,
		o.handle,
		index,
		string(objectType),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot replace Automerge object: %w", err)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}

// ScalarAt returns a typed scalar from a list index.
func (o *Object) ScalarAt(ctx context.Context, index uint64) (Scalar, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return Scalar{}, ErrClosed
	}

	encoded, err := o.document.backend.GetScalarAt(ctx, o.handle, index)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot get Automerge scalar: %w", err)
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot decode Automerge scalar: %w", err)
	}

	return value, nil
}

// ObjectAt returns a composite value from a list index.
func (o *Object) ObjectAt(ctx context.Context, index uint64) (*Object, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	handle, rawType, err := o.document.backend.GetObjectAt(
		ctx,
		o.handle,
		index,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge object: %w", err)
	}

	objectType := ObjectType(rawType)
	if !validObjectType(objectType) {
		return nil, fmt.Errorf("unknown Automerge object type %q", objectType)
	}

	return &Object{
		document: o.document,
		handle:   handle,
		Type:     objectType,
	}, nil
}

// DeleteKey deletes a map property.
func (o *Object) DeleteKey(ctx context.Context, key string) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	if err := o.document.backend.DeleteMap(ctx, o.handle, key); err != nil {
		return fmt.Errorf("cannot delete Automerge map property: %w", err)
	}

	return nil
}

// DeleteIndex deletes a list element.
func (o *Object) DeleteIndex(ctx context.Context, index uint64) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	if err := o.document.backend.DeleteSequence(ctx, o.handle, index); err != nil {
		return fmt.Errorf("cannot delete Automerge sequence element: %w", err)
	}

	return nil
}

// Increment adds delta to a counter stored at a map property.
func (o *Object) Increment(ctx context.Context, key string, delta int64) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	if err := o.document.backend.Increment(
		ctx,
		o.handle,
		key,
		delta,
	); err != nil {
		return fmt.Errorf("cannot increment Automerge counter: %w", err)
	}

	return nil
}

// IncrementAt adds delta to a counter stored at a list index.
func (o *Object) IncrementAt(
	ctx context.Context,
	index uint64,
	delta int64,
) error {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return ErrClosed
	}

	if err := o.document.backend.IncrementAt(
		ctx,
		o.handle,
		index,
		delta,
	); err != nil {
		return fmt.Errorf("cannot increment Automerge sequence counter: %w", err)
	}

	return nil
}

// Len returns the visible length of a list or text object.
func (o *Object) Len(ctx context.Context) (uint64, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return 0, ErrClosed
	}

	length, err := o.document.backend.Length(ctx, o.handle)
	if err != nil {
		return 0, fmt.Errorf("cannot get Automerge object length: %w", err)
	}

	return length, nil
}

// Keys returns visible map property names in lexical order.
func (o *Object) Keys(ctx context.Context) ([]string, error) {
	o.document.mu.Lock()
	defer o.document.mu.Unlock()

	if o.document.closed {
		return nil, ErrClosed
	}

	keys, err := o.document.backend.Keys(ctx, o.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge map keys: %w", err)
	}

	return keys, nil
}

// Text returns a collaborative text wrapper for a text object.
func (o *Object) Text() (*Text, error) {
	if o.Type != ObjectTypeText {
		return nil, fmt.Errorf("automerge object is %q, not text", o.Type)
	}

	return &Text{document: o.document, handle: o.handle}, nil
}

func validObjectType(value ObjectType) bool {
	switch value {
	case ObjectTypeMap, ObjectTypeList, ObjectTypeText, ObjectTypeTable:
		return true
	default:
		return false
	}
}

func backendHashes(heads []Hash) [][32]byte {
	result := make([][32]byte, len(heads))
	for i, head := range heads {
		result[i] = [32]byte(head)
	}

	return result
}
