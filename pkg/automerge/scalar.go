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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
)

type (
	// ScalarType identifies one Automerge scalar value type.
	ScalarType string

	// Scalar preserves the exact type of one Automerge scalar.
	Scalar struct {
		Type   ScalarType
		Bool   bool
		Uint   uint64
		Int    int64
		Float  float64
		String string
		Bytes  []byte
	}

	scalarWire struct {
		Type   ScalarType `json:"type"`
		Bool   bool       `json:"bool"`
		Uint   uint64     `json:"uint"`
		Int    int64      `json:"int"`
		Float  uint64     `json:"floatBits"`
		String string     `json:"string"`
		Bytes  string     `json:"bytes"`
	}
)

const (
	ScalarTypeNull      ScalarType = "null"
	ScalarTypeBoolean   ScalarType = "boolean"
	ScalarTypeUint      ScalarType = "uint"
	ScalarTypeInt       ScalarType = "int"
	ScalarTypeFloat64   ScalarType = "float64"
	ScalarTypeString    ScalarType = "string"
	ScalarTypeBytes     ScalarType = "bytes"
	ScalarTypeCounter   ScalarType = "counter"
	ScalarTypeTimestamp ScalarType = "timestamp"
)

// PutScalar assigns a typed scalar at a key in the root map.
func (d *Document) PutScalar(ctx context.Context, key string, value Scalar) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge scalar: %w", err)
	}

	if err := d.backend.PutScalar(ctx, rootObject, key, encoded); err != nil {
		return fmt.Errorf("cannot put Automerge scalar: %w", err)
	}

	return nil
}

// Scalar returns a typed scalar from a key in the root map.
func (d *Document) Scalar(ctx context.Context, key string) (Scalar, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Scalar{}, ErrClosed
	}

	encoded, err := d.backend.GetScalar(ctx, rootObject, key)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot get Automerge scalar: %w", err)
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return Scalar{}, fmt.Errorf("cannot decode Automerge scalar: %w", err)
	}

	return value, nil
}

// Scalars returns every concurrent scalar value at a key in the root map.
func (d *Document) Scalars(ctx context.Context, key string) ([]Scalar, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	encoded, err := d.backend.GetAllScalars(ctx, rootObject, key)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge scalar conflicts: %w", err)
	}

	values, err := decodeScalarWires(encoded)
	if err != nil {
		return nil, fmt.Errorf("cannot decode Automerge scalar conflicts: %w", err)
	}

	return values, nil
}

func encodeScalarWire(value Scalar) ([]byte, error) {
	wire := scalarWire{
		Type:   value.Type,
		Bool:   value.Bool,
		Uint:   value.Uint,
		Int:    value.Int,
		Float:  math.Float64bits(value.Float),
		String: value.String,
		Bytes:  hex.EncodeToString(value.Bytes),
	}
	if !validScalarType(value.Type) {
		return nil, fmt.Errorf("unknown scalar type %q", value.Type)
	}

	return json.Marshal(wire)
}

func decodeScalarWire(data []byte) (Scalar, error) {
	var wire scalarWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return Scalar{}, err
	}

	if !validScalarType(wire.Type) {
		return Scalar{}, fmt.Errorf("unknown scalar type %q", wire.Type)
	}

	var bytes []byte

	if wire.Type == ScalarTypeBytes {
		var err error

		bytes, err = hex.DecodeString(wire.Bytes)
		if err != nil {
			return Scalar{}, fmt.Errorf("cannot decode scalar bytes: %w", err)
		}
	}

	return Scalar{
		Type:   wire.Type,
		Bool:   wire.Bool,
		Uint:   wire.Uint,
		Int:    wire.Int,
		Float:  math.Float64frombits(wire.Float),
		String: wire.String,
		Bytes:  bytes,
	}, nil
}

func decodeScalarWires(data []byte) ([]Scalar, error) {
	var encoded []json.RawMessage
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, err
	}

	values := make([]Scalar, len(encoded))
	for i, value := range encoded {
		decoded, err := decodeScalarWire(value)
		if err != nil {
			return nil, fmt.Errorf("cannot decode scalar %d: %w", i, err)
		}

		values[i] = decoded
	}

	return values, nil
}

func validScalarType(value ScalarType) bool {
	switch value {
	case ScalarTypeNull,
		ScalarTypeBoolean,
		ScalarTypeUint,
		ScalarTypeInt,
		ScalarTypeFloat64,
		ScalarTypeString,
		ScalarTypeBytes,
		ScalarTypeCounter,
		ScalarTypeTimestamp:
		return true
	default:
		return false
	}
}
