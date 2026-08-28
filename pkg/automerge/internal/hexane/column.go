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

package hexane

import (
	"fmt"
	"io"
)

type (
	encodedValue struct {
		data []byte
		err  error
	}

	// Value is either a present column value or a first-class null.
	Value[T any] struct {
		Value   T
		Valid   bool
		encoded *encodedValue
	}

	// ColumnSplice describes one edit in original-column coordinates.
	ColumnSplice[T any] struct {
		Index       int
		DeleteCount int
		Inserted    *Column[T]
	}

	// Column is a mutable chunked sequence of nullable values. Its immutable AVL
	// rope gives Clone O(1) copy-on-write semantics. An edit copies only search
	// paths, boundary chunks, and newly inserted values.
	Column[T any] struct {
		codec     Codec[T]
		root      *ropeNode[Value[T]]
		metrics   ropeMetrics[Value[T]]
		chunkSize int
		encoded   []byte
		tail      encodedColumnTail[T]
	}

	encodedColumnTail[T any] struct {
		valid         bool
		hasPresent    bool
		kind          byte
		count         int
		offset        int
		payloadOffset int
		lastOffset    int
		last          Value[T]
	}
)

const (
	columnChunkSize = 256
	encodedTailNull = iota
	encodedTailRepeat
	encodedTailLiteral
)

// Some constructs a present Value.
func Some[T any](value T) Value[T] {
	return Value[T]{Value: value, Valid: true}
}

// Null constructs a null Value.
func Null[T any]() Value[T] {
	return Value[T]{}
}

// NewColumn constructs an empty column using codec.
func NewColumn[T any](codec Codec[T]) *Column[T] {
	return newColumn(codec, nil, nil)
}

// NewColumnFromValues constructs a column from logical values.
func NewColumnFromValues[T any](codec Codec[T], values ...Value[T]) *Column[T] {
	return newColumn(codec, values, nil)
}

func newColumn[T any](
	codec Codec[T],
	values []Value[T],
	metrics ropeMetrics[Value[T]],
) *Column[T] {
	if codec == nil {
		panic("hexane: nil codec")
	}

	column := &Column[T]{
		codec:     codec,
		metrics:   metrics,
		chunkSize: columnChunkSize,
	}
	column.root = ropeFrom(values, column.chunkSize, column.prepareValue, metrics)
	return column
}

// Len returns the logical number of values.
func (c *Column[T]) Len() int {
	if c.root == nil {
		return 0
	}

	return c.root.len
}

// Get returns the value at index.
func (c *Column[T]) Get(index int) Value[T] {
	c.checkIndex(index)
	return c.cloneValue(ropeGet(c.root, index))
}

// Insert inserts values before index.
func (c *Column[T]) Insert(index int, values ...Value[T]) {
	c.Splice(index, 0, values...)
}

// Set replaces the value at index.
func (c *Column[T]) Set(index int, value Value[T]) {
	c.Splice(index, 1, value)
}

// Delete removes count values at index.
func (c *Column[T]) Delete(index, count int) {
	c.Splice(index, count)
}

// Splice removes deleteCount values and inserts values at index.
func (c *Column[T]) Splice(index, deleteCount int, values ...Value[T]) {
	c.checkRange(index, deleteCount)
	if deleteCount == 0 && len(values) == 0 {
		return
	}

	left, rest := ropeSplit(c.root, index, c.chunkSize, c.metrics)
	_, right := ropeSplit(rest, deleteCount, c.chunkSize, c.metrics)
	inserted := ropeFrom(values, c.chunkSize, c.prepareValue, c.metrics)
	c.root = ropeConcat(
		ropeConcat(left, inserted, c.chunkSize, c.metrics),
		right,
		c.chunkSize,
		c.metrics,
	)
	c.encoded = nil
	c.tail = encodedColumnTail[T]{}
}

// BatchSplice applies non-overlapping edits atomically. Indexes refer to the
// column state before any edit in the batch.
func (c *Column[T]) BatchSplice(splices []ColumnSplice[T]) error {
	appendOnly := len(splices) == 1 &&
		splices[0].Index == c.Len() &&
		splices[0].DeleteCount == 0 &&
		splices[0].Inserted != nil &&
		c.tail.valid
	var appended []Value[T]
	if appendOnly {
		appended = splices[0].Inserted.Values()
	}
	ropeSplices := make([]ropeSplice[Value[T]], len(splices))
	for i, splice := range splices {
		var inserted *ropeNode[Value[T]]
		if splice.Inserted != nil {
			inserted = splice.Inserted.root
		}
		ropeSplices[i] = ropeSplice[Value[T]]{
			index:       splice.Index,
			deleteCount: splice.DeleteCount,
			inserted:    inserted,
		}
	}

	root, err := ropeBatchSplice(
		c.root,
		ropeSplices,
		c.chunkSize,
		c.metrics,
	)
	if err != nil {
		return fmt.Errorf("cannot splice column batch: %w", err)
	}
	c.root = root
	if appendOnly {
		for _, value := range appended {
			if err := c.appendEncodedValue(c.prepareValue(value)); err != nil {
				c.encoded = nil
				c.tail = encodedColumnTail[T]{}
				break
			}
		}
	} else {
		c.encoded = nil
		c.tail = encodedColumnTail[T]{}
	}

	return nil
}

// Clone returns an O(1), copy-on-write clone.
func (c *Column[T]) Clone() *Column[T] {
	return &Column[T]{
		codec:     c.codec,
		root:      c.root,
		metrics:   c.metrics,
		chunkSize: c.chunkSize,
		encoded:   c.encoded,
		tail:      c.tail,
	}
}

// Values returns an independent flat snapshot of the column.
func (c *Column[T]) Values() []Value[T] {
	values := make([]Value[T], 0, c.Len())
	ropeEach(
		c.root,
		func(value Value[T]) bool {
			values = append(values, c.cloneValue(value))
			return true
		},
	)
	return values
}

// Bytes returns the canonical Automerge RLE wire representation.
func (c *Column[T]) Bytes() ([]byte, error) {
	data, err := c.encodedBytes()
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), data...), nil
}

// SaveTo writes the canonical Automerge RLE wire representation.
func (c *Column[T]) SaveTo(w io.Writer) (int64, error) {
	data, err := c.encodedBytes()
	if err != nil {
		return 0, err
	}
	return saveBytes(w, data)
}

func (c *Column[T]) encodedBytes() ([]byte, error) {
	if c.encoded != nil || c.tail.valid {
		return c.encoded, nil
	}
	leaves := make([][]Value[T], 0)
	ropeEachLeaf(
		c.root,
		func(values []Value[T]) bool {
			leaves = append(leaves, values)
			return true
		},
	)
	cursor := columnValueCursor[T]{leaves: leaves}
	hasPresent := false
	for check := cursor; ; {
		value, ok := check.next()
		if !ok {
			break
		}
		if !value.Valid {
			continue
		}
		hasPresent = true
		if value.encoded.err != nil {
			return nil, value.encoded.err
		}
	}
	if !hasPresent {
		c.encoded = []byte{}
		c.tail = encodedColumnTail[T]{
			valid:      true,
			hasPresent: false,
			kind:       encodedTailNull,
			count:      c.Len(),
		}
		return c.encoded, nil
	}

	data := make([]byte, 0)
	for {
		value, ok := cursor.peek()
		if !ok {
			c.encoded = data
			return c.encoded, nil
		}
		if !value.Valid {
			end := cursor
			count := 0
			for {
				value, ok := end.peek()
				if !ok || value.Valid {
					break
				}
				_, _ = end.next()
				count++
			}
			offset := len(data)
			data = appendULEB(appendLEB(data, 0), uint64(count))
			c.tail = encodedColumnTail[T]{
				valid:         true,
				hasPresent:    true,
				kind:          encodedTailNull,
				count:         count,
				offset:        offset,
				payloadOffset: len(data),
			}
			cursor = end
			continue
		}

		next := cursor
		first, _ := next.next()
		second, hasSecond := next.peek()
		if hasSecond && second.Valid && c.codec.Equal(first.Value, second.Value) {
			count := 1
			for {
				value, ok := next.peek()
				if !ok || !value.Valid || !c.codec.Equal(first.Value, value.Value) {
					break
				}
				_, _ = next.next()
				count++
			}
			offset := len(data)
			data = appendLEB(data, int64(count))
			payloadOffset := len(data)
			data = append(data, first.encoded.data...)
			c.tail = encodedColumnTail[T]{
				valid:         true,
				hasPresent:    true,
				kind:          encodedTailRepeat,
				count:         count,
				offset:        offset,
				payloadOffset: payloadOffset,
				lastOffset:    payloadOffset,
				last:          first,
			}
			cursor = next
			continue
		}

		end := cursor
		count := 0
		for {
			value, ok := end.peek()
			if !ok || !value.Valid {
				break
			}
			_, _ = end.next()
			count++
			following, ok := end.peek()
			if !ok || !following.Valid {
				break
			}
			lookahead := end
			current, _ := lookahead.next()
			after, ok := lookahead.peek()
			if ok && after.Valid && c.codec.Equal(current.Value, after.Value) {
				break
			}
		}
		offset := len(data)
		data = appendLEB(data, -int64(count))
		payloadOffset := len(data)
		lastOffset := payloadOffset
		var last Value[T]
		for range count {
			value, _ := cursor.next()
			lastOffset = len(data)
			data = append(data, value.encoded.data...)
			last = value
		}
		c.tail = encodedColumnTail[T]{
			valid:         true,
			hasPresent:    true,
			kind:          encodedTailLiteral,
			count:         count,
			offset:        offset,
			payloadOffset: payloadOffset,
			lastOffset:    lastOffset,
			last:          last,
		}
	}
}

func (c *Column[T]) appendEncodedValue(value Value[T]) error {
	tail := c.tail
	if !tail.valid {
		return fmt.Errorf("hexane: encoded tail is unavailable")
	}
	if value.Valid && value.encoded.err != nil {
		return value.encoded.err
	}
	if !tail.hasPresent {
		if !value.Valid {
			tail.count++
			c.tail = tail
			return nil
		}
		data := make([]byte, 0, len(value.encoded.data)+8)
		if tail.count > 0 {
			data = appendULEB(appendLEB(data, 0), uint64(tail.count))
		}
		offset := len(data)
		data = appendLEB(data, -1)
		payloadOffset := len(data)
		data = append(data, value.encoded.data...)
		c.encoded = data
		c.tail = encodedColumnTail[T]{
			valid:         true,
			hasPresent:    true,
			kind:          encodedTailLiteral,
			count:         1,
			offset:        offset,
			payloadOffset: payloadOffset,
			lastOffset:    payloadOffset,
			last:          value,
		}
		return nil
	}

	switch tail.kind {
	case encodedTailNull:
		if !value.Valid {
			data := append([]byte(nil), c.encoded[:tail.offset]...)
			data = appendULEB(appendLEB(data, 0), uint64(tail.count+1))
			c.encoded = data
			tail.count++
			c.tail = tail
			return nil
		}
		data := append([]byte(nil), c.encoded...)
		offset := len(data)
		data = appendLEB(data, -1)
		payloadOffset := len(data)
		data = append(data, value.encoded.data...)
		c.encoded = data
		c.tail = encodedColumnTail[T]{
			valid:         true,
			hasPresent:    true,
			kind:          encodedTailLiteral,
			count:         1,
			offset:        offset,
			payloadOffset: payloadOffset,
			lastOffset:    payloadOffset,
			last:          value,
		}
	case encodedTailRepeat:
		if value.Valid && c.codec.Equal(tail.last.Value, value.Value) {
			data := append([]byte(nil), c.encoded[:tail.offset]...)
			data = appendLEB(data, int64(tail.count+1))
			payloadOffset := len(data)
			data = append(data, tail.last.encoded.data...)
			c.encoded = data
			tail.count++
			tail.payloadOffset = payloadOffset
			tail.lastOffset = payloadOffset
			c.tail = tail
			return nil
		}
		data := append([]byte(nil), c.encoded...)
		if !value.Valid {
			offset := len(data)
			data = appendULEB(appendLEB(data, 0), 1)
			c.encoded = data
			c.tail = encodedColumnTail[T]{
				valid:         true,
				hasPresent:    true,
				kind:          encodedTailNull,
				count:         1,
				offset:        offset,
				payloadOffset: len(data),
			}
			return nil
		}
		offset := len(data)
		data = appendLEB(data, -1)
		payloadOffset := len(data)
		data = append(data, value.encoded.data...)
		c.encoded = data
		c.tail = encodedColumnTail[T]{
			valid:         true,
			hasPresent:    true,
			kind:          encodedTailLiteral,
			count:         1,
			offset:        offset,
			payloadOffset: payloadOffset,
			lastOffset:    payloadOffset,
			last:          value,
		}
	case encodedTailLiteral:
		if !value.Valid {
			data := append([]byte(nil), c.encoded...)
			offset := len(data)
			data = appendULEB(appendLEB(data, 0), 1)
			c.encoded = data
			c.tail = encodedColumnTail[T]{
				valid:         true,
				hasPresent:    true,
				kind:          encodedTailNull,
				count:         1,
				offset:        offset,
				payloadOffset: len(data),
			}
			return nil
		}
		if c.codec.Equal(tail.last.Value, value.Value) {
			data := append([]byte(nil), c.encoded[:tail.offset]...)
			if tail.count > 1 {
				data = appendLEB(data, -int64(tail.count-1))
				data = append(data, c.encoded[tail.payloadOffset:tail.lastOffset]...)
			}
			offset := len(data)
			data = appendLEB(data, 2)
			payloadOffset := len(data)
			data = append(data, tail.last.encoded.data...)
			c.encoded = data
			c.tail = encodedColumnTail[T]{
				valid:         true,
				hasPresent:    true,
				kind:          encodedTailRepeat,
				count:         2,
				offset:        offset,
				payloadOffset: payloadOffset,
				lastOffset:    payloadOffset,
				last:          value,
			}
			return nil
		}
		data := append([]byte(nil), c.encoded[:tail.offset]...)
		data = appendLEB(data, -int64(tail.count+1))
		payloadOffset := len(data)
		data = append(data, c.encoded[tail.payloadOffset:]...)
		lastOffset := len(data)
		data = append(data, value.encoded.data...)
		c.encoded = data
		tail.count++
		tail.payloadOffset = payloadOffset
		tail.lastOffset = lastOffset
		tail.last = value
		c.tail = tail
	default:
		return fmt.Errorf("hexane: invalid encoded tail")
	}

	return nil
}

func (c *Column[T]) cloneValue(value Value[T]) Value[T] {
	if value.Valid {
		value.Value = c.codec.Clone(value.Value)
	}
	value.encoded = nil

	return value
}

func (c *Column[T]) prepareValue(value Value[T]) Value[T] {
	value = c.cloneValue(value)
	if !value.Valid {
		return value
	}

	data, err := c.codec.Append(nil, value.Value)
	value.encoded = &encodedValue{data: data, err: err}
	return value
}

func (c *Column[T]) checkIndex(index int) {
	if index < 0 || index >= c.Len() {
		panic("hexane: column index out of range")
	}
}

func (c *Column[T]) checkRange(index, count int) {
	length := c.Len()
	if index < 0 || count < 0 || index > length || count > length-index {
		panic("hexane: column slice bounds out of range")
	}
}

type columnValueCursor[T any] struct {
	leaves [][]Value[T]
	leaf   int
	index  int
}

func (c *columnValueCursor[T]) peek() (Value[T], bool) {
	for c.leaf < len(c.leaves) && c.index == len(c.leaves[c.leaf]) {
		c.leaf++
		c.index = 0
	}
	if c.leaf == len(c.leaves) {
		return Value[T]{}, false
	}

	return c.leaves[c.leaf][c.index], true
}

func (c *columnValueCursor[T]) next() (Value[T], bool) {
	value, ok := c.peek()
	if ok {
		c.index++
	}

	return value, ok
}
