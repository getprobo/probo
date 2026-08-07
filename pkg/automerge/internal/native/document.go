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

import "fmt"

type (
	Document struct {
		Hash          [32]byte
		Actors        [][]byte
		Heads         [][32]byte
		ChangeColumns []RawColumn
		OpColumns     []RawColumn
		HeadIndices   []uint64
		Raw           []byte
	}
)

const maxDocumentActors = 1024 * 1024

func ParseDocument(raw []byte) (*Document, error) {
	chunks, err := ParseChunks(raw)
	if err != nil {
		return nil, err
	}
	if len(chunks) != 1 {
		return nil, fmt.Errorf("expected one document chunk, got %d", len(chunks))
	}
	chunk := chunks[0]
	if chunk.Type != ChunkTypeDocument {
		return nil, fmt.Errorf("expected document chunk, got type %d", chunk.Type)
	}

	r := newReader(chunk.Data)
	actors, err := readActorArray(r, maxDocumentActors)
	if err != nil {
		return nil, fmt.Errorf("cannot read document actors: %w", err)
	}
	heads, err := readHashes(r, maxDependencies)
	if err != nil {
		return nil, fmt.Errorf("cannot read document heads: %w", err)
	}
	changeMetadata, err := readColumnMetadata(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read document change metadata: %w", err)
	}
	opMetadata, err := readColumnMetadata(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read document operation metadata: %w", err)
	}
	changeColumns, err := readColumns(r, changeMetadata)
	if err != nil {
		return nil, fmt.Errorf("cannot read document change columns: %w", err)
	}
	opColumns, err := readColumns(r, opMetadata)
	if err != nil {
		return nil, fmt.Errorf("cannot read document operation columns: %w", err)
	}

	headIndices := make([]uint64, 0, len(heads))
	for !r.done() {
		if len(headIndices) == len(heads) {
			return nil, fmt.Errorf("document has more head indices than heads")
		}
		index, err := r.readULEB128()
		if err != nil {
			return nil, fmt.Errorf("cannot read document head index %d: %w", len(headIndices), err)
		}
		headIndices = append(headIndices, index)
	}

	return &Document{
		Hash:          chunk.Hash,
		Actors:        actors,
		Heads:         heads,
		ChangeColumns: changeColumns,
		OpColumns:     opColumns,
		HeadIndices:   headIndices,
		Raw:           append([]byte(nil), raw...),
	}, nil
}

func readActorArray(r *reader, maxCount uint64) ([][]byte, error) {
	count, err := r.readULEB128()
	if err != nil {
		return nil, err
	}
	if count > maxCount {
		return nil, fmt.Errorf("actor count %d exceeds limit %d", count, maxCount)
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
