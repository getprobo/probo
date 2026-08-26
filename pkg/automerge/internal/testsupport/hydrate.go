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

package testsupport

import (
	"fmt"
	"slices"
	"time"
)

type (
	// ValueType identifies one hydrated Automerge value.
	ValueType string

	// Value is a recursively hydrated Automerge value.
	Value struct {
		Type   ValueType
		Scalar Scalar
		Map    map[string]Value
		List   []Value
		Text   string
	}
)

const (
	ValueTypeScalar ValueType = "scalar"
	ValueTypeMap    ValueType = "map"
	ValueTypeList   ValueType = "list"
	ValueTypeText   ValueType = "text"
)

// NewFrom creates and commits a document from a hydrated root map.
func NewFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
) (*Document, error) {
	return newFrom(actorID, value, message, timestamp, New)
}

// NewReferenceFrom creates a hydrated document using the Rust/WASM oracle.
func NewReferenceFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
) (*Document, error) {
	return newFrom(
		actorID,
		value,
		message,
		timestamp,
		NewReference,
	)
}

func newFrom(
	actorID ActorID,
	value map[string]Value,
	message string,
	timestamp time.Time,
	factory func(ActorID) (*Document, error),
) (*Document, error) {
	document, err := factory(actorID)
	if err != nil {
		return nil, err
	}

	if err := document.Root().PutMap(value); err != nil {
		_ = document.Close()
		return nil, err
	}

	if _, err := document.Commit(message, timestamp); err != nil {
		_ = document.Close()
		return nil, err
	}

	return document, nil
}

// PutMap assigns a batch of recursively hydrated map properties.
func (o *Object) PutMap(values map[string]Value) error {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		if err := o.PutValue(key, values[key]); err != nil {
			return fmt.Errorf("cannot put hydrated property %q: %w", key, err)
		}
	}

	return nil
}

// PutValue assigns one recursively hydrated value to a map property.
func (o *Object) PutValue(key string, value Value) error {
	switch value.Type {
	case ValueTypeScalar:
		return o.PutScalar(key, value.Scalar)
	case ValueTypeMap:
		child, err := o.CreateObject(key, ObjectTypeMap)
		if err != nil {
			return err
		}

		return child.PutMap(value.Map)
	case ValueTypeList:
		child, err := o.CreateObject(key, ObjectTypeList)
		if err != nil {
			return err
		}

		return child.InsertValues(0, value.List)
	case ValueTypeText:
		child, err := o.CreateObject(key, ObjectTypeText)
		if err != nil {
			return err
		}

		text := &Text{document: child.document, handle: child.handle}

		return text.Splice(0, 0, value.Text)
	default:
		return fmt.Errorf("unknown hydrated value type %q", value.Type)
	}
}

// InsertValues inserts recursively hydrated values into a list.
func (o *Object) InsertValues(
	index uint64,
	values []Value,
) error {
	for offset, value := range values {
		if err := o.InsertValue(index+uint64(offset), value); err != nil {
			return fmt.Errorf("cannot insert hydrated value %d: %w", offset, err)
		}
	}

	return nil
}

// InsertValue inserts one recursively hydrated value into a list.
func (o *Object) InsertValue(
	index uint64,
	value Value,
) error {
	switch value.Type {
	case ValueTypeScalar:
		return o.InsertScalar(index, value.Scalar)
	case ValueTypeMap:
		child, err := o.InsertObject(index, ObjectTypeMap)
		if err != nil {
			return err
		}

		return child.PutMap(value.Map)
	case ValueTypeList:
		child, err := o.InsertObject(index, ObjectTypeList)
		if err != nil {
			return err
		}

		return child.InsertValues(0, value.List)
	case ValueTypeText:
		child, err := o.InsertObject(index, ObjectTypeText)
		if err != nil {
			return err
		}

		text := &Text{document: child.document, handle: child.handle}

		return text.Splice(0, 0, value.Text)
	default:
		return fmt.Errorf("unknown hydrated value type %q", value.Type)
	}
}

// PutValueAt replaces a list element with one recursively hydrated value.
func (o *Object) PutValueAt(
	index uint64,
	value Value,
) error {
	switch value.Type {
	case ValueTypeScalar:
		return o.PutScalarAt(index, value.Scalar)
	case ValueTypeMap:
		child, err := o.putObjectAt(index, ObjectTypeMap)
		if err != nil {
			return err
		}

		return child.PutMap(value.Map)
	case ValueTypeList:
		child, err := o.putObjectAt(index, ObjectTypeList)
		if err != nil {
			return err
		}

		return child.InsertValues(0, value.List)
	case ValueTypeText:
		child, err := o.putObjectAt(index, ObjectTypeText)
		if err != nil {
			return err
		}

		text := &Text{document: child.document, handle: child.handle}

		return text.Splice(0, 0, value.Text)
	default:
		return fmt.Errorf("unknown hydrated value type %q", value.Type)
	}
}

// SpliceValues deletes and inserts recursively hydrated list values.
func (o *Object) SpliceValues(
	index uint64,
	deleteCount uint64,
	values []Value,
) error {
	for range deleteCount {
		if err := o.DeleteIndex(index); err != nil {
			return err
		}
	}

	return o.InsertValues(index, values)
}

func (o *Object) putObjectAt(
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

	handle, err := o.document.engine.PutObjectAt(
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
