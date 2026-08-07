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

type rleDecoder[T any] struct {
	reader      *reader
	decodeValue func(*reader) (T, error)
	lastValue   *T
	remaining   uint64
	literal     bool
}

const maxRLERunLength = 64 * 1024 * 1024

func newRLEDecoder[T any](
	data []byte,
	decodeValue func(*reader) (T, error),
) *rleDecoder[T] {
	return &rleDecoder[T]{
		reader:      newReader(data),
		decodeValue: decodeValue,
	}
}

func (d *rleDecoder[T]) next() (value T, null bool, err error) {
	for d.remaining == 0 {
		if d.reader.done() {
			return value, true, nil
		}

		count, err := d.reader.readSLEB128()
		if err != nil {
			return value, false, fmt.Errorf("cannot read RLE count: %w", err)
		}
		switch {
		case count > 0:
			if count > maxRLERunLength {
				return value, false, fmt.Errorf("RLE run length %d exceeds limit", count)
			}
			d.remaining = uint64(count)
			decoded, err := d.decodeValue(d.reader)
			if err != nil {
				return value, false, fmt.Errorf("cannot read RLE run value: %w", err)
			}
			d.lastValue = new(decoded)
			d.literal = false
		case count < 0:
			if count == -1<<63 || -count > maxRLERunLength {
				return value, false, fmt.Errorf("RLE literal length %d exceeds limit", count)
			}
			d.remaining = uint64(-count)
			d.lastValue = nil
			d.literal = true
		default:
			nullCount, err := d.reader.readULEB128()
			if err != nil {
				return value, false, fmt.Errorf("cannot read RLE null count: %w", err)
			}
			if nullCount == 0 || nullCount > maxRLERunLength {
				return value, false, fmt.Errorf("invalid RLE null count %d", nullCount)
			}
			d.remaining = nullCount
			d.lastValue = nil
			d.literal = false
		}
	}

	d.remaining--
	if d.literal {
		decoded, err := d.decodeValue(d.reader)
		if err != nil {
			return value, false, fmt.Errorf("cannot read RLE literal value: %w", err)
		}
		return decoded, false, nil
	}
	if d.lastValue == nil {
		return value, true, nil
	}
	return *d.lastValue, false, nil
}

func (d *rleDecoder[T]) done() bool {
	return d.remaining == 0 && d.reader.done()
}

func decodeRLEUint(data []byte) *rleDecoder[uint64] {
	return newRLEDecoder(
		data,
		func(r *reader) (uint64, error) {
			return r.readULEB128()
		},
	)
}

func decodeRLEInt(data []byte) *rleDecoder[int64] {
	return newRLEDecoder(
		data,
		func(r *reader) (int64, error) {
			return r.readSLEB128()
		},
	)
}

func decodeRLEString(data []byte) *rleDecoder[string] {
	return newRLEDecoder(
		data,
		func(r *reader) (string, error) {
			value, err := r.readLengthPrefixedBytes(maxMessageBytes)
			if err != nil {
				return "", err
			}
			return string(value), nil
		},
	)
}
