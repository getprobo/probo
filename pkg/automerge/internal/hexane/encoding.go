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

// Package hexane contains dependency-free, Go-native mutable column
// primitives. Its behavior and wire encodings are derived from Automerge
// Hexane 1.0.0-alpha.5, copyright its contributors and licensed under MIT.
package hexane

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

type (
	// Codec defines value equality, ownership, and Automerge wire encoding for
	// a Column. Clone must return a value that callers may mutate independently.
	Codec[T any] interface {
		Equal(left, right T) bool
		Clone(value T) T
		Append(dst []byte, value T) ([]byte, error)
	}

	// CodecFuncs adapts functions into a Codec.
	CodecFuncs[T any] struct {
		EqualFunc  func(left, right T) bool
		CloneFunc  func(value T) T
		AppendFunc func(dst []byte, value T) ([]byte, error)
	}

	uint64Codec struct{}
	int64Codec  struct{}
	stringCodec struct{}
	bytesCodec  struct{}
)

// Equal implements Codec.
func (c CodecFuncs[T]) Equal(left, right T) bool {
	return c.EqualFunc(left, right)
}

// Clone implements Codec.
func (c CodecFuncs[T]) Clone(value T) T {
	return c.CloneFunc(value)
}

// Append implements Codec.
func (c CodecFuncs[T]) Append(dst []byte, value T) ([]byte, error) {
	return c.AppendFunc(dst, value)
}

// Uint64Codec returns the standard unsigned LEB128 codec.
func Uint64Codec() Codec[uint64] {
	return uint64Codec{}
}

// Int64Codec returns the standard signed LEB128 codec.
func Int64Codec() Codec[int64] {
	return int64Codec{}
}

// StringCodec returns the Automerge length-prefixed UTF-8 string codec.
func StringCodec() Codec[string] {
	return stringCodec{}
}

// BytesCodec returns the Automerge length-prefixed byte-string codec.
func BytesCodec() Codec[[]byte] {
	return bytesCodec{}
}

func (uint64Codec) Equal(left, right uint64) bool { return left == right }
func (uint64Codec) Clone(value uint64) uint64     { return value }
func (uint64Codec) Append(dst []byte, value uint64) ([]byte, error) {
	return appendULEB(dst, value), nil
}

func (int64Codec) Equal(left, right int64) bool { return left == right }
func (int64Codec) Clone(value int64) int64      { return value }
func (int64Codec) Append(dst []byte, value int64) ([]byte, error) {
	return appendLEB(dst, value), nil
}

func (stringCodec) Equal(left, right string) bool { return left == right }
func (stringCodec) Clone(value string) string     { return value }
func (stringCodec) Append(dst []byte, value string) ([]byte, error) {
	if !utf8.ValidString(value) {
		return nil, fmt.Errorf("string is not valid UTF-8")
	}

	dst = appendULEB(dst, uint64(len(value)))

	return append(dst, value...), nil
}

func (bytesCodec) Equal(left, right []byte) bool { return bytes.Equal(left, right) }
func (bytesCodec) Clone(value []byte) []byte     { return bytes.Clone(value) }
func (bytesCodec) Append(dst []byte, value []byte) ([]byte, error) {
	dst = appendULEB(dst, uint64(len(value)))
	return append(dst, value...), nil
}

func appendULEB(dst []byte, value uint64) []byte {
	for {
		current := byte(value & 0x7f)

		value >>= 7
		if value != 0 {
			current |= 0x80
		}

		dst = append(dst, current)
		if value == 0 {
			return dst
		}
	}
}

func appendLEB(dst []byte, value int64) []byte {
	for {
		current := byte(value & 0x7f)
		value >>= 7

		done := (value == 0 && current&0x40 == 0) ||
			(value == -1 && current&0x40 != 0)
		if !done {
			current |= 0x80
		}

		dst = append(dst, current)
		if done {
			return dst
		}
	}
}

func saveBytes(w io.Writer, data []byte) (int64, error) {
	written := 0
	for written < len(data) {
		count, err := w.Write(data[written:])

		written += count
		if err != nil {
			return int64(written), fmt.Errorf("cannot write column: %w", err)
		}

		if count == 0 {
			return int64(written), fmt.Errorf("cannot write column: %w", io.ErrShortWrite)
		}
	}

	return int64(written), nil
}
