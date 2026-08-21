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

// Package encoding contains the bounded binary primitives shared by Automerge
// storage, cursor, and protocol codecs.
package encoding

import "fmt"

type Reader struct {
	data   []byte
	offset int
}

func NewReader(data []byte) *Reader               { return &Reader{data: data} }
func NewReaderAt(data []byte, offset int) *Reader { return &Reader{data: data, offset: offset} }
func (r *Reader) Offset() int                     { return r.offset }
func (r *Reader) Remaining() int                  { return len(r.data) - r.offset }
func (r *Reader) Byte() (byte, error) {
	if r.Remaining() < 1 {
		return 0, fmt.Errorf("unexpected end of data")
	}

	value := r.data[r.offset]
	r.offset++

	return value, nil
}
func (r *Reader) Bytes(length uint64) ([]byte, error) {
	if length > uint64(r.Remaining()) {
		return nil, fmt.Errorf("need %d bytes, only %d remain", length, r.Remaining())
	}

	start := r.offset
	r.offset += int(length)

	return r.data[start:r.offset], nil
}
func (r *Reader) ULEB() (uint64, error) {
	var value uint64

	for shift := uint(0); shift < 64; shift += 7 {
		b, err := r.Byte()
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
func AppendULEB(data []byte, value uint64) []byte {
	for value >= 0x80 {
		data = append(data, byte(value)|0x80)
		value >>= 7
	}

	return append(data, byte(value))
}
func AppendLengthPrefixed(data, value []byte) []byte {
	data = AppendULEB(data, uint64(len(value)))
	return append(data, value...)
}
func DecodeLengthPrefixed(r *Reader) ([]byte, error) {
	length, err := r.ULEB()
	if err != nil {
		return nil, err
	}

	return r.Bytes(length)
}
