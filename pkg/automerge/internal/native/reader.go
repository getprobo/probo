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

type reader struct {
	data   []byte
	offset int
}

func newReader(data []byte) *reader {
	return &reader{data: data}
}

func (r *reader) done() bool {
	return r.offset == len(r.data)
}

func (r *reader) remaining() []byte {
	return r.data[r.offset:]
}

func (r *reader) read(length int) ([]byte, error) {
	if length < 0 || length > len(r.data)-r.offset {
		return nil, fmt.Errorf(
			"read of %d bytes exceeds %d remaining bytes",
			length,
			len(r.data)-r.offset,
		)
	}

	value := r.data[r.offset : r.offset+length]
	r.offset += length
	return value, nil
}

func (r *reader) readULEB128() (uint64, error) {
	value, consumed, err := readULEB128(r.remaining())
	if err != nil {
		return 0, err
	}
	r.offset += consumed
	return value, nil
}

func (r *reader) readSLEB128() (int64, error) {
	var (
		value int64
		shift uint
	)

	for i, current := range r.remaining() {
		if i == 10 {
			return 0, fmt.Errorf("signed LEB128 overflow")
		}

		value |= int64(current&0x7f) << shift
		shift += 7
		r.offset++
		if current&0x80 == 0 {
			if shift < 64 && current&0x40 != 0 {
				value |= ^int64(0) << shift
			}
			return value, nil
		}
	}

	return 0, fmt.Errorf("truncated signed LEB128")
}

func (r *reader) readLengthPrefixedBytes(maxLength uint64) ([]byte, error) {
	length, err := r.readULEB128()
	if err != nil {
		return nil, fmt.Errorf("cannot read byte length: %w", err)
	}
	if length > maxLength {
		return nil, fmt.Errorf("byte length %d exceeds limit %d", length, maxLength)
	}
	value, err := r.read(int(length))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), value...), nil
}
