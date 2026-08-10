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
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

type (
	reader struct {
		data   []byte
		offset int
	}

	columnMeta struct {
		specification uint32
		normalized    uint32
		length        uint64
		compressed    bool
	}

	column struct {
		specification uint32
		data          []byte
	}

	optional[T any] struct {
		value T
		valid bool
	}
)

const (
	maxDecodedItems  = 100_000_000
	maxInflatedBytes = 512 << 20
)

func (r *reader) remaining() int {
	return len(r.data) - r.offset
}

func (r *reader) bytes(length uint64) ([]byte, error) {
	if length > uint64(r.remaining()) {
		return nil, fmt.Errorf(
			"need %d bytes at offset %d, only %d remain",
			length,
			r.offset,
			r.remaining(),
		)
	}

	if length > uint64(math.MaxInt) {
		return nil, fmt.Errorf("byte length %d exceeds platform capacity", length)
	}

	start := r.offset
	r.offset += int(length)

	return r.data[start:r.offset], nil
}

func (r *reader) byte() (byte, error) {
	value, err := r.bytes(1)
	if err != nil {
		return 0, err
	}

	return value[0], nil
}

func (r *reader) uleb() (uint64, error) {
	start := r.offset

	var value uint64

	for i := range 10 {
		current, err := r.byte()
		if err != nil {
			return 0, fmt.Errorf("cannot decode uLEB at offset %d: %w", start, err)
		}

		payload := uint64(current & 0x7f)
		if i == 9 && payload > 1 {
			return 0, fmt.Errorf("uLEB at offset %d overflows uint64", start)
		}

		value |= payload << (7 * i)
		if current&0x80 == 0 {
			if i > 0 && payload == 0 {
				return 0, fmt.Errorf("uLEB at offset %d is not minimally encoded", start)
			}

			return value, nil
		}
	}

	return 0, fmt.Errorf("uLEB at offset %d exceeds 10 bytes", start)
}

func (r *reader) leb() (int64, error) {
	start := r.offset

	var (
		value   uint64
		current byte
		shift   uint
	)

	for i := range 10 {
		var err error

		current, err = r.byte()
		if err != nil {
			return 0, fmt.Errorf("cannot decode LEB at offset %d: %w", start, err)
		}

		payload := uint64(current & 0x7f)
		if i == 9 {
			if payload != 0 && payload != 0x7f {
				return 0, fmt.Errorf("LEB at offset %d overflows int64", start)
			}

			if payload == 0 && current&0x40 != 0 {
				return 0, fmt.Errorf("LEB at offset %d overflows int64", start)
			}

			if payload == 0x7f && current&0x40 == 0 {
				return 0, fmt.Errorf("LEB at offset %d overflows int64", start)
			}
		}

		value |= payload << shift
		shift += 7

		if current&0x80 == 0 {
			if i > 0 {
				previous := r.data[r.offset-2]
				if current == 0 && previous&0x40 == 0 {
					return 0, fmt.Errorf("LEB at offset %d is not minimally encoded", start)
				}

				if current == 0x7f && previous&0x40 != 0 {
					return 0, fmt.Errorf("LEB at offset %d is not minimally encoded", start)
				}
			}

			if shift < 64 && current&0x40 != 0 {
				value |= ^uint64(0) << shift
			}

			return int64(value), nil
		}
	}

	return 0, fmt.Errorf("LEB at offset %d exceeds 10 bytes", start)
}

func decodeRLE[T any](data []byte, decodeValue func(*reader) (T, error)) ([]optional[T], error) {
	r := &reader{data: data}
	values := make([]optional[T], 0)

	for r.remaining() > 0 {
		run, err := r.leb()
		if err != nil {
			return nil, fmt.Errorf("cannot decode run length: %w", err)
		}

		switch {
		case run > 0:
			value, err := decodeValue(r)
			if err != nil {
				return nil, fmt.Errorf("cannot decode repeated value: %w", err)
			}

			if err := appendRepeated(&values, optional[T]{value: value, valid: true}, uint64(run)); err != nil {
				return nil, err
			}
		case run == 0:
			count, err := r.uleb()
			if err != nil {
				return nil, fmt.Errorf("cannot decode null run length: %w", err)
			}

			if count == 0 {
				return nil, fmt.Errorf("null run cannot be empty")
			}

			if err := appendRepeated(&values, optional[T]{}, count); err != nil {
				return nil, err
			}
		default:
			if run == math.MinInt64 {
				return nil, fmt.Errorf("literal run length overflows")
			}

			count := uint64(-run)
			if err := reserveItems(len(values), count); err != nil {
				return nil, err
			}

			for range count {
				value, err := decodeValue(r)
				if err != nil {
					return nil, fmt.Errorf("cannot decode literal value: %w", err)
				}

				values = append(values, optional[T]{value: value, valid: true})
			}
		}
	}

	return values, nil
}

func appendRepeated[T any](values *[]optional[T], value optional[T], count uint64) error {
	if err := reserveItems(len(*values), count); err != nil {
		return err
	}

	for range count {
		*values = append(*values, value)
	}

	return nil
}

func reserveItems(existing int, additional uint64) error {
	if additional > maxDecodedItems || uint64(existing)+additional > maxDecodedItems {
		return fmt.Errorf("decoded column exceeds %d items", maxDecodedItems)
	}

	return nil
}

func decodeULEBColumn(data []byte) ([]optional[uint64], error) {
	return decodeRLE(data, func(r *reader) (uint64, error) {
		return r.uleb()
	})
}

func decodeDeltaColumn(data []byte) ([]optional[uint64], error) {
	deltas, err := decodeRLE(data, func(r *reader) (int64, error) {
		return r.leb()
	})
	if err != nil {
		return nil, err
	}

	values := make([]optional[uint64], len(deltas))

	var previous uint64

	for i, delta := range deltas {
		if !delta.valid {
			continue
		}

		next, err := addSigned(previous, delta.value)
		if err != nil {
			return nil, fmt.Errorf("delta item %d: %w", i, err)
		}

		values[i] = optional[uint64]{value: next, valid: true}
		previous = next
	}

	return values, nil
}

func decodeSignedDeltaColumn(data []byte) ([]optional[int64], error) {
	deltas, err := decodeRLE(data, func(r *reader) (int64, error) {
		return r.leb()
	})
	if err != nil {
		return nil, err
	}

	values := make([]optional[int64], len(deltas))

	var previous int64

	for i, delta := range deltas {
		if !delta.valid {
			continue
		}

		if delta.value > 0 && previous > math.MaxInt64-delta.value {
			return nil, fmt.Errorf("delta item %d overflows int64", i)
		}

		if delta.value < 0 && previous < math.MinInt64-delta.value {
			return nil, fmt.Errorf("delta item %d underflows int64", i)
		}

		previous += delta.value
		values[i] = optional[int64]{value: previous, valid: true}
	}

	return values, nil
}

func addSigned(value uint64, delta int64) (uint64, error) {
	if delta >= 0 {
		addition := uint64(delta)
		if value > math.MaxUint64-addition {
			return 0, fmt.Errorf("positive delta overflows uint64")
		}

		return value + addition, nil
	}

	if delta == math.MinInt64 {
		subtraction := uint64(math.MaxInt64) + 1
		if subtraction > value {
			return 0, fmt.Errorf("negative delta underflows uint64")
		}

		return value - subtraction, nil
	}

	subtraction := uint64(-delta)
	if subtraction > value {
		return 0, fmt.Errorf("negative delta underflows uint64")
	}

	return value - subtraction, nil
}

func decodeStringColumn(data []byte) ([]optional[string], error) {
	return decodeRLE(data, func(r *reader) (string, error) {
		length, err := r.uleb()
		if err != nil {
			return "", err
		}

		value, err := r.bytes(length)
		if err != nil {
			return "", err
		}

		if !utf8.Valid(value) {
			return "", fmt.Errorf("string is not valid UTF-8")
		}

		return string(value), nil
	})
}

func decodeBooleanColumn(data []byte, expected int) ([]bool, error) {
	r := &reader{data: data}
	values := make([]bool, 0, expected)
	current := false

	for r.remaining() > 0 {
		count, err := r.uleb()
		if err != nil {
			return nil, fmt.Errorf("cannot decode boolean run: %w", err)
		}

		if err := reserveItems(len(values), count); err != nil {
			return nil, err
		}

		for range count {
			values = append(values, current)
		}

		current = !current
	}

	if len(values) != expected {
		return nil, fmt.Errorf("boolean column has %d items, expected %d", len(values), expected)
	}

	return values, nil
}

func parseColumnMetadata(r *reader, allowCompressed bool) ([]columnMeta, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, fmt.Errorf("cannot decode column count: %w", err)
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("column count %d exceeds limit", count)
	}

	metadata := make([]columnMeta, 0, count)

	var previous uint32

	for i := range count {
		rawSpec, err := r.uleb()
		if err != nil {
			return nil, fmt.Errorf("cannot decode column %d specification: %w", i, err)
		}

		if rawSpec > math.MaxUint32 {
			return nil, fmt.Errorf("column %d specification %d exceeds uint32", i, rawSpec)
		}

		specification := uint32(rawSpec)
		compressed := specification&8 != 0
		normalized := specification &^ 8

		if compressed && !allowCompressed {
			return nil, fmt.Errorf("column %d is compressed in a change chunk", i)
		}

		if i > 0 && normalized <= previous {
			return nil, fmt.Errorf("column %d specification %d is not strictly sorted", i, normalized)
		}

		previous = normalized

		length, err := r.uleb()
		if err != nil {
			return nil, fmt.Errorf("cannot decode column %d length: %w", i, err)
		}

		metadata = append(
			metadata,
			columnMeta{
				specification: specification,
				normalized:    normalized,
				length:        length,
				compressed:    compressed,
			},
		)
	}

	return metadata, nil
}

func readColumns(r *reader, metadata []columnMeta) (map[uint32]column, error) {
	columns := make(map[uint32]column, len(metadata))
	for i, meta := range metadata {
		data, err := r.bytes(meta.length)
		if err != nil {
			return nil, fmt.Errorf("cannot read column %d: %w", i, err)
		}

		if meta.compressed {
			data, err = inflate(data)
			if err != nil {
				return nil, fmt.Errorf("cannot inflate column %d: %w", i, err)
			}
		}

		columns[meta.normalized] = column{
			specification: meta.specification,
			data:          append([]byte(nil), data...),
		}
	}

	return columns, nil
}

// deflate compresses data with raw DEFLATE at best compression, matching the
// stream inflate reads. It is used to produce compressed change chunks.
func deflate(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	writer, err := flate.NewWriter(&buffer, flate.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("cannot create DEFLATE writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("cannot write DEFLATE stream: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("cannot finish DEFLATE stream: %w", err)
	}

	return buffer.Bytes(), nil
}

func inflate(data []byte) ([]byte, error) {
	compressed := flate.NewReader(bytes.NewReader(data))
	defer func() { _ = compressed.Close() }()

	output, err := io.ReadAll(io.LimitReader(compressed, maxInflatedBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read DEFLATE stream: %w", err)
	}

	if len(output) > maxInflatedBytes {
		return nil, fmt.Errorf("inflated data exceeds %d bytes", maxInflatedBytes)
	}

	return output, nil
}

func decodeScalars(metaData, rawData []byte, expected int) ([]optional[Scalar], error) {
	metadata, err := decodeULEBColumn(metaData)
	if err != nil {
		return nil, fmt.Errorf("cannot decode value metadata: %w", err)
	}

	if len(metadata) != expected {
		return nil, fmt.Errorf("value metadata has %d items, expected %d", len(metadata), expected)
	}

	raw := &reader{data: rawData}
	values := make([]optional[Scalar], expected)

	for i, item := range metadata {
		if !item.valid {
			continue
		}

		scalarType := ScalarType(item.value & 0x0f)
		length := item.value >> 4

		valueBytes, err := raw.bytes(length)
		if err != nil {
			return nil, fmt.Errorf("cannot read scalar %d: %w", i, err)
		}

		scalar, err := decodeScalar(scalarType, valueBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot decode scalar %d: %w", i, err)
		}

		values[i] = optional[Scalar]{value: scalar, valid: true}
	}

	if raw.remaining() != 0 {
		return nil, fmt.Errorf("value column has %d trailing bytes", raw.remaining())
	}

	return values, nil
}

func decodeScalar(scalarType ScalarType, data []byte) (Scalar, error) {
	scalar := Scalar{Type: scalarType}
	switch scalarType {
	case ScalarNull:
		if len(data) != 0 {
			return Scalar{}, fmt.Errorf("null has length %d, expected 0", len(data))
		}
	case ScalarFalse, ScalarTrue:
		if len(data) != 0 {
			return Scalar{}, fmt.Errorf("boolean has length %d, expected 0", len(data))
		}

		scalar.Bool = scalarType == ScalarTrue
	case ScalarUint:
		r := &reader{data: data}

		value, err := r.uleb()
		if err != nil || r.remaining() != 0 {
			return Scalar{}, fmt.Errorf("invalid unsigned integer scalar")
		}

		scalar.Uint = value
	case ScalarInt, ScalarCounter, ScalarTimestamp:
		r := &reader{data: data}

		value, err := r.leb()
		if err != nil || r.remaining() != 0 {
			return Scalar{}, fmt.Errorf("invalid signed integer scalar")
		}

		scalar.Int = value
	case ScalarFloat64:
		if len(data) != 8 {
			return Scalar{}, fmt.Errorf("float has length %d, expected 8", len(data))
		}

		scalar.Float = math.Float64frombits(binary.LittleEndian.Uint64(data))
	case ScalarString:
		if !utf8.Valid(data) {
			return Scalar{}, fmt.Errorf("string scalar is not valid UTF-8")
		}

		scalar.String = string(data)
	case ScalarBytes:
		scalar.Bytes = append([]byte(nil), data...)
	default:
		scalar.Raw = append([]byte(nil), data...)
	}

	return scalar, nil
}

func requireColumn(columns map[uint32]column, specification uint32) ([]byte, error) {
	value, ok := columns[specification]
	if !ok {
		return nil, fmt.Errorf("required column %d is missing", specification)
	}

	delete(columns, specification)

	return value.data, nil
}

func optionalColumn(columns map[uint32]column, specification uint32) []byte {
	value, ok := columns[specification]
	if !ok {
		return nil
	}

	delete(columns, specification)

	return value.data
}

func requireItems[T any](name string, values []optional[T], expected int, nullable bool) error {
	if len(values) != expected {
		return fmt.Errorf("%s column has %d items, expected %d", name, len(values), expected)
	}

	if nullable {
		return nil
	}

	for i, value := range values {
		if !value.valid {
			return fmt.Errorf("%s column item %d is null", name, i)
		}
	}

	return nil
}

func copyHash(data []byte) ChangeHash {
	var hash ChangeHash
	copy(hash[:], data)

	return hash
}
