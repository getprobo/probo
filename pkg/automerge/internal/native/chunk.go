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
	"crypto/sha256"
	"fmt"
)

type (
	ChunkType byte

	Chunk struct {
		Type     ChunkType
		Checksum [4]byte
		Hash     [32]byte
		Data     []byte
		Raw      []byte
	}
)

const (
	ChunkTypeDocument   ChunkType = 0
	ChunkTypeChange     ChunkType = 1
	ChunkTypeCompressed ChunkType = 2
	ChunkTypeBundle     ChunkType = 3

	maxChunkBytes = 64 * 1024 * 1024
)

var chunkMagic = [4]byte{0x85, 0x6f, 0x4a, 0x83}

func ParseChunks(data []byte) ([]Chunk, error) {
	chunks := make([]Chunk, 0, 1)

	for offset := 0; offset < len(data); {
		chunk, consumed, err := parseChunk(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Automerge chunk at offset %d: %w", offset, err)
		}
		chunks = append(chunks, chunk)
		offset += consumed
	}

	return chunks, nil
}

func parseChunk(data []byte) (Chunk, int, error) {
	if len(data) < 10 {
		return Chunk{}, 0, fmt.Errorf("truncated chunk header")
	}
	if [4]byte(data[:4]) != chunkMagic {
		return Chunk{}, 0, fmt.Errorf("invalid magic bytes")
	}

	chunkType := ChunkType(data[8])
	if chunkType > ChunkTypeBundle {
		return Chunk{}, 0, fmt.Errorf("unknown chunk type %d", chunkType)
	}

	length, lengthBytes, err := readULEB128(data[9:])
	if err != nil {
		return Chunk{}, 0, fmt.Errorf("cannot decode chunk length: %w", err)
	}
	if length > maxChunkBytes {
		return Chunk{}, 0, fmt.Errorf("chunk length %d exceeds limit %d", length, maxChunkBytes)
	}

	headerLength := 9 + lengthBytes
	if length > uint64(len(data)-headerLength) {
		return Chunk{}, 0, fmt.Errorf(
			"chunk declares %d bytes with only %d available",
			length,
			len(data)-headerLength,
		)
	}
	totalLength := headerLength + int(length)
	payload := data[headerLength:totalLength]
	hash := hashChunk(chunkType, payload)
	checksum := [4]byte(data[4:8])
	if chunkType != ChunkTypeCompressed && checksum != [4]byte(hash[:4]) {
		return Chunk{}, 0, fmt.Errorf("chunk checksum mismatch")
	}

	return Chunk{
		Type:     chunkType,
		Checksum: checksum,
		Hash:     hash,
		Data:     append([]byte(nil), payload...),
		Raw:      append([]byte(nil), data[:totalLength]...),
	}, totalLength, nil
}

func hashChunk(chunkType ChunkType, data []byte) [32]byte {
	input := []byte{byte(chunkType)}
	input = appendULEB128(input, uint64(len(data)))
	input = append(input, data...)
	return sha256.Sum256(input)
}
