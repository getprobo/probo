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
	ColumnSpec uint32

	RawColumn struct {
		Spec ColumnSpec
		Data []byte
	}

	columnMetadata struct {
		spec   ColumnSpec
		length int
	}
)

const maxInflatedColumnBytes = 64 * 1024 * 1024

func (s ColumnSpec) Deflated() bool {
	return uint32(s)&0x08 != 0
}

func (s ColumnSpec) normalized() uint32 {
	return uint32(s) &^ 0x08
}

func findColumn(columns []RawColumn, specification ColumnSpec) []byte {
	target := specification.normalized()
	for _, column := range columns {
		if column.Spec.normalized() == target {
			return column.Data
		}
	}
	return nil
}

func readColumnMetadata(r *reader) ([]columnMetadata, error) {
	count, err := r.readULEB128()
	if err != nil {
		return nil, fmt.Errorf("cannot read column count: %w", err)
	}
	if count > 256 {
		return nil, fmt.Errorf("column count %d exceeds limit 256", count)
	}

	metadata := make([]columnMetadata, int(count))
	for i := range metadata {
		spec, err := r.readULEB128()
		if err != nil {
			return nil, fmt.Errorf("cannot read column %d specification: %w", i, err)
		}
		length, err := r.readULEB128()
		if err != nil {
			return nil, fmt.Errorf("cannot read column %d length: %w", i, err)
		}
		if length > maxChunkBytes {
			return nil, fmt.Errorf("column %d length %d exceeds limit", i, length)
		}
		metadata[i] = columnMetadata{
			spec:   ColumnSpec(spec),
			length: int(length),
		}
	}

	return metadata, nil
}

func readColumns(r *reader, metadata []columnMetadata) ([]RawColumn, error) {
	columns := make([]RawColumn, len(metadata))
	for i, description := range metadata {
		data, err := r.read(description.length)
		if err != nil {
			return nil, fmt.Errorf("cannot read column %d data: %w", i, err)
		}
		if description.spec.Deflated() {
			data, err = inflateColumn(data)
			if err != nil {
				return nil, fmt.Errorf("cannot inflate column %d: %w", i, err)
			}
		} else {
			data = append([]byte(nil), data...)
		}
		columns[i] = RawColumn{
			Spec: description.spec,
			Data: data,
		}
	}
	return columns, nil
}

func inflateColumn(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer func() { _ = reader.Close() }()

	value, err := io.ReadAll(io.LimitReader(reader, maxInflatedColumnBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read deflated data: %w", err)
	}
	if len(value) > maxInflatedColumnBytes {
		return nil, fmt.Errorf("inflated column exceeds %d bytes", maxInflatedColumnBytes)
	}
	return value, nil
}
