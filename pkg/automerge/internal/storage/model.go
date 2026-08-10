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

// Package storage implements Automerge chunk and column encoding, decoding, and
// graph validation independently from the native execution engine.
package storage

import "go.probo.inc/probo/pkg/automerge/internal/types"

type (
	ActorID    = types.ActorID
	ChangeHash = types.ChangeHash
	OpID       = types.OpID
	ObjectID   = types.ObjectID
	Key        = types.Key
	Action     = types.Action
	ScalarType = types.ScalarType
	Scalar     = types.Scalar
	Operation  = types.Operation
	Change     = types.Change
	ChunkType  = types.ChunkType
	RawColumn  = types.RawColumn
	Document   = types.Document
)

const (
	ActionMakeMap   = types.ActionMakeMap
	ActionSet       = types.ActionSet
	ActionMakeList  = types.ActionMakeList
	ActionDelete    = types.ActionDelete
	ActionMakeText  = types.ActionMakeText
	ActionIncrement = types.ActionIncrement
	ActionMakeTable = types.ActionMakeTable
	ActionMark      = types.ActionMark

	ScalarNull      = types.ScalarNull
	ScalarFalse     = types.ScalarFalse
	ScalarTrue      = types.ScalarTrue
	ScalarUint      = types.ScalarUint
	ScalarInt       = types.ScalarInt
	ScalarFloat64   = types.ScalarFloat64
	ScalarString    = types.ScalarString
	ScalarBytes     = types.ScalarBytes
	ScalarCounter   = types.ScalarCounter
	ScalarTimestamp = types.ScalarTimestamp

	ChunkDocument         = types.ChunkDocument
	ChunkChange           = types.ChunkChange
	ChunkCompressedChange = types.ChunkCompressedChange
)

var (
	NewActorID = types.NewActorID
	RootObject = types.RootObject
)
