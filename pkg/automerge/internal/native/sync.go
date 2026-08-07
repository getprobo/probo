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
	SyncMessageVersion byte

	SyncHave struct {
		LastSync [][32]byte
		Bloom    []byte
	}

	SyncMessage struct {
		Version SyncMessageVersion
		Heads   [][32]byte
		Need    [][32]byte
		Have    []SyncHave
		Changes [][]byte
		Flags   []byte
	}
)

const (
	SyncMessageVersion1 SyncMessageVersion = 0x42
	SyncMessageVersion2 SyncMessageVersion = 0x43

	maxSyncHaveEntries = 1024
	maxSyncChanges     = 1024 * 1024
	maxSyncBloomBytes  = 1024 * 1024
	maxSyncFlagsBytes  = 1024
	maxSyncChunkBytes  = 64 * 1024 * 1024
	maxSyncHashes      = 1024 * 1024
)

func ParseSyncMessage(data []byte) (*SyncMessage, error) {
	r := &reader{data: data}

	versionByte, err := r.byte()
	if err != nil {
		return nil, fmt.Errorf("cannot read sync message version: %w", err)
	}

	version := SyncMessageVersion(versionByte)
	if version != SyncMessageVersion1 && version != SyncMessageVersion2 {
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

	if haveCount > maxSyncHaveEntries {
		return nil, fmt.Errorf("sync have count %d exceeds limit", haveCount)
	}

	have := make([]SyncHave, int(haveCount))
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

	return &SyncMessage{
		Version: version,
		Heads:   heads,
		Need:    need,
		Have:    have,
		Changes: changes,
		Flags:   flags,
	}, nil
}

func (m SyncMessage) Encode() ([]byte, error) {
	if m.Version != SyncMessageVersion1 && m.Version != SyncMessageVersion2 {
		return nil, fmt.Errorf("unsupported sync message version 0x%02x", m.Version)
	}

	if len(m.Have) > maxSyncHaveEntries || len(m.Changes) > maxSyncChanges {
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
