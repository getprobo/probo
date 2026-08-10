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

// Package types defines the shared Automerge CRDT and storage model.
// It is dependency-free so storage, sync, and native execution packages can
// exchange changes without importing one another.
package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
)

type (
	// ActorID is an immutable, arbitrary-length Automerge actor identifier.
	ActorID string

	// ChangeHash identifies a change by the SHA-256 digest of its change chunk.
	ChangeHash [32]byte

	// OpID is an Automerge Lamport timestamp.
	OpID struct {
		Actor   ActorID
		Counter uint64
	}

	// ObjectID identifies either the root map or the operation that made an
	// object.
	ObjectID struct {
		OpID   OpID
		IsRoot bool
	}

	// Key identifies either a map property or a sequence element.
	Key struct {
		Property *string
		Element  *OpID
		IsHead   bool
	}

	// Action is the numeric action stored in an operation.
	Action uint64

	// ScalarType is the four-bit scalar tag from a value metadata column.
	ScalarType uint8

	// Scalar preserves the exact Automerge scalar type and its typed value.
	// Raw is populated for unknown, forward-compatible scalar types.
	Scalar struct {
		Type   ScalarType
		Bool   bool
		Uint   uint64
		Int    int64
		Float  float64
		String string
		Bytes  []byte
		Raw    []byte
	}

	// Operation is one decoded Automerge operation.
	Operation struct {
		ID           OpID
		Object       ObjectID
		Key          Key
		Insert       bool
		Action       Action
		Value        *Scalar
		Predecessors []OpID
		Successors   []OpID
		MarkExpand   *bool
		MarkName     *string
	}

	// Change is one decoded change. Hash is nil for non-head changes in a
	// document chunk because snapshots store only head hashes.
	Change struct {
		Hash              *ChangeHash
		Actor             ActorID
		Sequence          uint64
		StartOp           uint64
		MaxOp             uint64
		Time              int64
		Message           string
		Dependencies      []ChangeHash
		DependencyIndexes []uint64
		Operations        []Operation
		Extra             *Scalar
		ExtraBytes        []byte
		Raw               []byte
	}

	// ChunkType identifies an Automerge storage chunk.
	ChunkType uint8

	// RawColumn retains a column not interpreted by this milestone.
	RawColumn struct {
		Specification uint32
		Data          []byte
	}

	// Document is a validated Automerge history.
	Document struct {
		Actors         []ActorID
		Heads          []ChangeHash
		Changes        []Change
		UnknownColumns []RawColumn
		ChunkTypes     []ChunkType
	}
)

const (
	ActionMakeMap   Action = 0
	ActionSet       Action = 1
	ActionMakeList  Action = 2
	ActionDelete    Action = 3
	ActionMakeText  Action = 4
	ActionIncrement Action = 5
	ActionMakeTable Action = 6
	ActionMark      Action = 7

	ScalarNull      ScalarType = 0
	ScalarFalse     ScalarType = 1
	ScalarTrue      ScalarType = 2
	ScalarUint      ScalarType = 3
	ScalarInt       ScalarType = 4
	ScalarFloat64   ScalarType = 5
	ScalarString    ScalarType = 6
	ScalarBytes     ScalarType = 7
	ScalarCounter   ScalarType = 8
	ScalarTimestamp ScalarType = 9

	ChunkDocument         ChunkType = 0
	ChunkChange           ChunkType = 1
	ChunkCompressedChange ChunkType = 2
)

// NewActorID validates and copies an actor ID.
func NewActorID(value []byte) (ActorID, error) {
	if len(value) == 0 {
		return "", fmt.Errorf("actor ID cannot be empty")
	}

	return ActorID(string(value)), nil
}

// Bytes returns a copy of the actor ID bytes.
func (a ActorID) Bytes() []byte {
	return []byte(string(a))
}

// String returns the lowercase hexadecimal actor ID.
func (a ActorID) String() string {
	return hex.EncodeToString(a.Bytes())
}

// Compare returns -1, 0, or 1 using Automerge's bytewise actor ordering.
func (a ActorID) Compare(other ActorID) int {
	return bytes.Compare(a.Bytes(), other.Bytes())
}

// concurrencyMagicBytes prefixes actor IDs derived for isolated writes so they
// cannot collide with real actor IDs. It matches Rust's CONCURRENCY_MAGIC_BYTES.
var concurrencyMagicBytes = [4]byte{0x13, 0xb2, 0x23, 0x09}

// WithConcurrency derives the isolation actor for the given concurrency level,
// mirroring Rust's ActorId::with_concurrency: the magic bytes, a ULEB128 level,
// then the base actor bytes. Level zero returns the base actor unchanged.
func (a ActorID) WithConcurrency(level uint64) ActorID {
	if level == 0 {
		return a
	}

	bytes := make([]byte, 0, 4+16+len(a))
	bytes = append(bytes, concurrencyMagicBytes[:]...)
	bytes = appendULEB(bytes, level)
	bytes = append(bytes, a.Bytes()...)

	return ActorID(string(bytes))
}

// String returns the lowercase hexadecimal change hash.
func (h ChangeHash) String() string {
	return hex.EncodeToString(h[:])
}

// Compare orders operation IDs by their Automerge Lamport timestamp.
func (o OpID) Compare(other OpID) int {
	switch {
	case o.Counter < other.Counter:
		return -1
	case o.Counter > other.Counter:
		return 1
	default:
		return o.Actor.Compare(other.Actor)
	}
}

// RootObject returns the distinguished root map identifier.
func RootObject() ObjectID {
	return ObjectID{IsRoot: true}
}

// IsKnown reports whether the scalar type is defined by Automerge 0.10.
func (s Scalar) IsKnown() bool {
	return s.Type <= ScalarTimestamp
}

// IsFinite reports whether a float scalar is finite. Non-float scalars are
// always finite.
func (s Scalar) IsFinite() bool {
	return s.Type != ScalarFloat64 || (!math.IsInf(s.Float, 0) && !math.IsNaN(s.Float))
}

func appendULEB(data []byte, value uint64) []byte {
	for value >= 0x80 {
		data = append(data, byte(value)|0x80)
		value >>= 7
	}

	return append(data, byte(value))
}
