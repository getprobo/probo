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

import "go.probo.inc/probo/pkg/automerge/internal/opset"

type (
	ActorID    = opset.ActorID
	ChangeHash = opset.ChangeHash
	OpID       = opset.OpID
	ObjectID   = opset.ObjectID
	Key        = opset.Key
	Action     = opset.Action
	ScalarType = opset.ScalarType
	Scalar     = opset.Scalar
	Operation  = opset.Operation
	Change     = opset.Change
	ChunkType  = opset.ChunkType
	RawColumn  = opset.RawColumn
	Document   = opset.Document
)

const (
	ActionMakeMap   = opset.ActionMakeMap
	ActionSet       = opset.ActionSet
	ActionMakeList  = opset.ActionMakeList
	ActionDelete    = opset.ActionDelete
	ActionMakeText  = opset.ActionMakeText
	ActionIncrement = opset.ActionIncrement
	ActionMakeTable = opset.ActionMakeTable
	ActionMark      = opset.ActionMark

	ScalarNull      = opset.ScalarNull
	ScalarFalse     = opset.ScalarFalse
	ScalarTrue      = opset.ScalarTrue
	ScalarUint      = opset.ScalarUint
	ScalarInt       = opset.ScalarInt
	ScalarFloat64   = opset.ScalarFloat64
	ScalarString    = opset.ScalarString
	ScalarBytes     = opset.ScalarBytes
	ScalarCounter   = opset.ScalarCounter
	ScalarTimestamp = opset.ScalarTimestamp

	ChunkDocument         = opset.ChunkDocument
	ChunkChange           = opset.ChunkChange
	ChunkCompressedChange = opset.ChunkCompressedChange
)

var (
	NewActorID = opset.NewActorID
	RootObject = opset.RootObject
)
