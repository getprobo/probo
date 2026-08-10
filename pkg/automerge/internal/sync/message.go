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

// Package sync implements the Automerge V1/V2 sync message wire format.
package sync

import "fmt"

type (
	MessageVersion byte

	Have struct {
		LastSync [][32]byte
		Bloom    []byte
	}

	Message struct {
		Version MessageVersion
		Heads   [][32]byte
		Need    [][32]byte
		Have    []Have
		Changes [][]byte
		Flags   []byte
	}
)

const (
	MessageVersion1 MessageVersion = 0x42
	MessageVersion2 MessageVersion = 0x43

	maxHaveEntries    = 1024
	maxSyncChanges    = 1024 * 1024
	maxSyncBloomBytes = 1024 * 1024
	maxSyncFlagsBytes = 1024
	maxSyncChunkBytes = 64 * 1024 * 1024
	maxSyncHashes     = 1024 * 1024
)

func ParseMessage(data []byte) (*Message, error) {
	r := &reader{data: data}

	versionByte, err := r.byte()
	if err != nil {
		return nil, fmt.Errorf("cannot read sync message version: %w", err)
	}

	version := MessageVersion(versionByte)
	if version != MessageVersion1 && version != MessageVersion2 {
		return nil, fmt.Errorf("unsupported sync message version 0x%02x", version)
	}

	heads, err := readSyncHashes(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read sync heads: %w", err)
	}

	need, err := readSyncHashes(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read sync needs: %w", err)
	}

	haveCount, err := r.uleb()
	if err != nil {
		return nil, fmt.Errorf("cannot read sync have count: %w", err)
	}

	if haveCount > maxHaveEntries {
		return nil, fmt.Errorf("sync have count %d exceeds limit", haveCount)
	}

	have := make([]Have, int(haveCount))
	for i := range have {
		have[i].LastSync, err = readSyncHashes(r)
		if err != nil {
			return nil, fmt.Errorf("cannot read sync have %d heads: %w", i, err)
		}

		have[i].Bloom, err = readSyncBytes(r, maxSyncBloomBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot read sync have %d bloom: %w", i, err)
		}
	}

	changeCount, err := r.uleb()
	if err != nil {
		return nil, fmt.Errorf("cannot read sync change count: %w", err)
	}

	if changeCount > maxSyncChanges {
		return nil, fmt.Errorf("sync change count %d exceeds limit", changeCount)
	}

	changes := make([][]byte, int(changeCount))
	for i := range changes {
		changes[i], err = readSyncBytes(r, maxSyncChunkBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot read sync change %d: %w", i, err)
		}
	}

	var flags []byte
	if r.remaining() > 0 {
		flags, err = readSyncBytes(r, maxSyncFlagsBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot read sync flags: %w", err)
		}
	}

	if r.remaining() > 0 {
		return nil, fmt.Errorf("sync message contains trailing bytes")
	}

	return &Message{
		Version: version,
		Heads:   heads,
		Need:    need,
		Have:    have,
		Changes: changes,
		Flags:   flags,
	}, nil
}

func (m Message) Encode() ([]byte, error) {
	if m.Version != MessageVersion1 && m.Version != MessageVersion2 {
		return nil, fmt.Errorf("unsupported sync message version 0x%02x", m.Version)
	}

	if len(m.Have) > maxHaveEntries || len(m.Changes) > maxSyncChanges {
		return nil, fmt.Errorf("sync message exceeds collection limits")
	}

	data := []byte{byte(m.Version)}
	data = appendHashes(data, m.Heads)
	data = appendHashes(data, m.Need)

	data = appendULEB(data, uint64(len(m.Have)))
	for _, have := range m.Have {
		data = appendHashes(data, have.LastSync)
		data = appendLengthPrefixedBytes(data, have.Bloom)
	}

	data = appendULEB(data, uint64(len(m.Changes)))
	for _, change := range m.Changes {
		data = appendLengthPrefixedBytes(data, change)
	}

	if m.Flags != nil {
		data = appendLengthPrefixedBytes(data, m.Flags)
	}

	return data, nil
}

func appendHashes(data []byte, hashes [][32]byte) []byte {
	data = appendULEB(data, uint64(len(hashes)))
	for _, hash := range hashes {
		data = append(data, hash[:]...)
	}

	return data
}

func appendLengthPrefixedBytes(data, value []byte) []byte {
	data = appendULEB(data, uint64(len(value)))
	return append(data, value...)
}

func readSyncHashes(r *reader) ([][32]byte, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if count > maxSyncHashes {
		return nil, fmt.Errorf("sync hash count %d exceeds limit", count)
	}

	hashes := make([][32]byte, int(count))
	for i := range hashes {
		value, err := r.bytes(32)
		if err != nil {
			return nil, fmt.Errorf("cannot read sync hash %d: %w", i, err)
		}

		copy(hashes[i][:], value)
	}

	return hashes, nil
}

func readSyncBytes(r *reader, limit uint64) ([]byte, error) {
	length, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if length > limit {
		return nil, fmt.Errorf("sync byte length %d exceeds limit %d", length, limit)
	}

	value, err := r.bytes(length)
	if err != nil {
		return nil, err
	}

	return append([]byte(nil), value...), nil
}

type reader struct {
	data   []byte
	offset int
}

func (r *reader) remaining() int { return len(r.data) - r.offset }

func (r *reader) byte() (byte, error) {
	if r.remaining() < 1 {
		return 0, fmt.Errorf("unexpected end of data")
	}

	b := r.data[r.offset]
	r.offset++

	return b, nil
}

func (r *reader) bytes(length uint64) ([]byte, error) {
	if length > uint64(r.remaining()) {
		return nil, fmt.Errorf("need %d bytes, only %d remain", length, r.remaining())
	}

	start := r.offset
	r.offset += int(length)

	return r.data[start:r.offset], nil
}

func (r *reader) uleb() (uint64, error) {
	var value uint64

	for shift := uint(0); shift < 64; shift += 7 {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}

		if shift == 63 && b > 1 {
			return 0, fmt.Errorf("ULEB128 overflow")
		}

		value |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return value, nil
		}
	}

	return 0, fmt.Errorf("ULEB128 overflow")
}

func appendULEB(data []byte, value uint64) []byte {
	for value >= 0x80 {
		data = append(data, byte(value)|0x80)
		value >>= 7
	}

	return append(data, byte(value))
}
