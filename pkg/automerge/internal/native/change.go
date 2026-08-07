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
	"compress/flate"
	"fmt"
	"io"
)

type (
	Change struct {
		Hash         [32]byte
		Dependencies [][32]byte
		Actor        []byte
		OtherActors  [][]byte
		Sequence     uint64
		StartOp      uint64
		Timestamp    int64
		Message      string
		Columns      []RawColumn
		ExtraBytes   []byte
		Raw          []byte
	}
)

const (
	maxActorBytes      = 1024
	maxActorsPerChange = 1024
	maxDependencies    = 1024
	maxMessageBytes    = 1024 * 1024
)

func ParseChange(raw []byte) (*Change, error) {
	chunks, err := ParseChunks(raw)
	if err != nil {
		return nil, err
	}
	if len(chunks) != 1 {
		return nil, fmt.Errorf("expected one change chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	switch chunk.Type {
	case ChunkTypeChange:
	case ChunkTypeCompressed:
		data, err := inflateChange(chunk.Data)
		if err != nil {
			return nil, err
		}
		hash := hashChunk(ChunkTypeChange, data)
		if chunk.Checksum != [4]byte(hash[:4]) {
			return nil, fmt.Errorf("compressed change checksum mismatch")
		}
		chunk.Data = data
		chunk.Hash = hash
	default:
		return nil, fmt.Errorf("expected change chunk, got type %d", chunk.Type)
	}

	change, err := parseChangeData(chunk.Data)
	if err != nil {
		return nil, err
	}
	change.Hash = chunk.Hash
	change.Raw = append([]byte(nil), raw...)
	return change, nil
}

func parseChangeData(data []byte) (*Change, error) {
	r := newReader(data)

	dependencies, err := readHashes(r, maxDependencies)
	if err != nil {
		return nil, fmt.Errorf("cannot read change dependencies: %w", err)
	}
	actor, err := r.readLengthPrefixedBytes(maxActorBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot read change actor: %w", err)
	}
	if len(actor) == 0 {
		return nil, fmt.Errorf("change actor cannot be empty")
	}
	sequence, err := r.readULEB128()
	if err != nil {
		return nil, fmt.Errorf("cannot read change sequence: %w", err)
	}
	startOp, err := r.readULEB128()
	if err != nil {
		return nil, fmt.Errorf("cannot read change start operation: %w", err)
	}
	timestamp, err := r.readSLEB128()
	if err != nil {
		return nil, fmt.Errorf("cannot read change timestamp: %w", err)
	}
	messageBytes, err := r.readLengthPrefixedBytes(maxMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot read change message: %w", err)
	}
	otherActors, err := readActors(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read other change actors: %w", err)
	}
	metadata, err := readColumnMetadata(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read change column metadata: %w", err)
	}
	columns, err := readColumns(r, metadata)
	if err != nil {
		return nil, fmt.Errorf("cannot read change columns: %w", err)
	}

	return &Change{
		Dependencies: dependencies,
		Actor:        actor,
		OtherActors:  otherActors,
		Sequence:     sequence,
		StartOp:      startOp,
		Timestamp:    timestamp,
		Message:      string(messageBytes),
		Columns:      columns,
		ExtraBytes:   append([]byte(nil), r.remaining()...),
	}, nil
}

func readHashes(r *reader, maxCount uint64) ([][32]byte, error) {
	count, err := r.readULEB128()
	if err != nil {
		return nil, err
	}
	if count > maxCount {
		return nil, fmt.Errorf("hash count %d exceeds limit %d", count, maxCount)
	}

	hashes := make([][32]byte, int(count))
	for i := range hashes {
		value, err := r.read(32)
		if err != nil {
			return nil, fmt.Errorf("cannot read hash %d: %w", i, err)
		}
		copy(hashes[i][:], value)
	}
	return hashes, nil
}

func readActors(r *reader) ([][]byte, error) {
	count, err := r.readULEB128()
	if err != nil {
		return nil, err
	}
	if count > maxActorsPerChange {
		return nil, fmt.Errorf("actor count %d exceeds limit %d", count, maxActorsPerChange)
	}

	actors := make([][]byte, int(count))
	for i := range actors {
		actors[i], err = r.readLengthPrefixedBytes(maxActorBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot read actor %d: %w", i, err)
		}
	}
	return actors, nil
}

func inflateChange(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer func() { _ = reader.Close() }()

	value, err := io.ReadAll(io.LimitReader(reader, maxChunkBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot inflate change: %w", err)
	}
	if len(value) > maxChunkBytes {
		return nil, fmt.Errorf("inflated change exceeds %d bytes", maxChunkBytes)
	}
	return value, nil
}
