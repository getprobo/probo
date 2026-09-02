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
	"slices"
)

type (
	boolCodec struct{}

	// DeltaSplice describes one absolute-value edit in original coordinates.
	DeltaSplice struct {
		Index       int
		DeleteCount int
		Inserted    *DeltaColumn
	}

	// PrefixSplice describes one prefix-column edit in original coordinates.
	PrefixSplice struct {
		Index       int
		DeleteCount int
		Inserted    *PrefixColumn
	}

	// BooleanSplice describes one boolean-column edit in original coordinates.
	BooleanSplice struct {
		Index       int
		DeleteCount int
		Inserted    *BooleanColumn
	}

	// DeltaColumn maintains nullable signed deltas in a copy-on-write rope.
	// Editing absolute values rewrites inserted deltas and at most the next
	// present boundary delta.
	DeltaColumn struct {
		deltas *Column[int64]
	}

	// PrefixColumn stores unsigned values with subtree sums, making Prefix
	// logarithmic in the number of chunks.
	PrefixColumn struct {
		values *Column[uint64]
	}

	// BooleanColumn stores booleans in copy-on-write logical byte chunks.
	// Serialization always starts each wire stream at a false-run boundary, so
	// edits cannot leave a partial-bit or parity-dependent chunk boundary.
	BooleanColumn struct {
		values     *Column[bool]
		encoded    []byte
		tailOffset int
		tailValue  bool
		tailCount  uint64
		encodedOK  bool
	}
)

// NewDeltaColumn constructs an empty DeltaColumn.
func NewDeltaColumn() *DeltaColumn {
	return NewDeltaColumnFromValues()
}

// NewDeltaColumnFromValues constructs a DeltaColumn from absolute values.
func NewDeltaColumnFromValues(values ...Value[int64]) *DeltaColumn {
	deltas := absoluteToDeltas(values)

	return &DeltaColumn{
		deltas: newColumn(
			Int64Codec(),
			deltas,
			func(value Value[int64]) (uint64, int64) {
				if !value.Valid {
					return 0, 0
				}

				return 0, value.Value
			},
		),
	}
}

// Len returns the number of values.
func (c *DeltaColumn) Len() int { return c.deltas.Len() }

// Get returns the absolute value at index.
func (c *DeltaColumn) Get(index int) Value[int64] {
	delta := c.deltas.Get(index)
	if !delta.Valid {
		return Null[int64]()
	}

	_, absolute := ropePrefix(c.deltas.root, index+1, c.deltas.metrics)

	return Some(absolute)
}

// Insert inserts absolute values before index.
func (c *DeltaColumn) Insert(index int, values ...Value[int64]) {
	c.Splice(index, 0, values...)
}

// Set replaces the absolute value at index.
func (c *DeltaColumn) Set(index int, value Value[int64]) {
	c.Splice(index, 1, value)
}

// Delete removes count values at index.
func (c *DeltaColumn) Delete(index, count int) {
	c.Splice(index, count)
}

// Splice removes deleteCount absolute values and inserts values at index.
func (c *DeltaColumn) Splice(
	index int,
	deleteCount int,
	values ...Value[int64],
) {
	c.deltas.checkRange(index, deleteCount)

	if deleteCount == 0 && len(values) == 0 {
		return
	}

	_, previous := ropePrefix(c.deltas.root, index, c.deltas.metrics)
	following, hasFollowing := c.nextPresent(index + deleteCount)

	inserted := make([]Value[int64], len(values))
	for i, value := range values {
		if !value.Valid {
			continue
		}

		inserted[i] = Some(value.Value - previous)
		previous = value.Value
	}

	c.deltas.Splice(index, deleteCount, inserted...)

	if !hasFollowing {
		return
	}

	for boundary := index + len(inserted); boundary < c.Len(); boundary++ {
		if c.deltas.Get(boundary).Valid {
			c.deltas.Set(boundary, Some(following-previous))
			return
		}
	}

	panic("hexane: missing delta boundary")
}

// BatchSplice applies non-overlapping absolute-value edits atomically.
func (c *DeltaColumn) BatchSplice(splices []DeltaSplice) error {
	ranges := make([]batchRange, len(splices))
	for i, splice := range splices {
		ranges[i] = batchRange{index: splice.Index, deleteCount: splice.DeleteCount}
	}

	if err := validateBatchRanges(c.Len(), ranges); err != nil {
		return fmt.Errorf("cannot splice delta column batch: %w", err)
	}

	if len(splices) == 1 &&
		splices[0].Index == c.Len() &&
		splices[0].DeleteCount == 0 &&
		splices[0].Inserted != nil {
		values := splices[0].Inserted.Values()
		_, previous := ropePrefix(c.deltas.root, c.Len(), c.deltas.metrics)

		deltas := make([]Value[int64], len(values))
		for i, value := range values {
			if !value.Valid {
				continue
			}

			deltas[i] = Some(value.Value - previous)
			previous = value.Value
		}

		inserted := newColumn(
			Int64Codec(),
			deltas,
			c.deltas.metrics,
		)

		next := c.deltas.Clone()
		if err := next.BatchSplice([]ColumnSplice[int64]{{
			Index:    c.Len(),
			Inserted: inserted,
		}}); err != nil {
			return fmt.Errorf("cannot append delta column batch: %w", err)
		}

		c.deltas = next

		return nil
	}

	sorted := append([]DeltaSplice(nil), splices...)
	slices.SortStableFunc(
		sorted,
		func(left, right DeltaSplice) int {
			return left.Index - right.Index
		},
	)
	slices.Reverse(sorted)

	next := c.Clone()

	for _, splice := range sorted {
		var values []Value[int64]
		if splice.Inserted != nil {
			values = splice.Inserted.Values()
		}

		next.Splice(splice.Index, splice.DeleteCount, values...)
	}

	c.deltas = next.deltas

	return nil
}

// Clone returns an O(1), copy-on-write clone.
func (c *DeltaColumn) Clone() *DeltaColumn {
	return &DeltaColumn{deltas: c.deltas.Clone()}
}

// Values returns an independent flat snapshot of absolute values.
func (c *DeltaColumn) Values() []Value[int64] {
	values := make([]Value[int64], 0, c.Len())

	var absolute int64

	ropeEach(
		c.deltas.root,
		func(delta Value[int64]) bool {
			if delta.Valid {
				absolute += delta.Value
				values = append(values, Some(absolute))
			} else {
				values = append(values, Null[int64]())
			}

			return true
		},
	)

	return values
}

// Bytes returns the maintained delta column's canonical RLE representation.
func (c *DeltaColumn) Bytes() ([]byte, error) {
	return c.deltas.Bytes()
}

// SaveTo writes the maintained delta column's canonical RLE representation.
func (c *DeltaColumn) SaveTo(w io.Writer) (int64, error) {
	return c.deltas.SaveTo(w)
}

func (c *DeltaColumn) nextPresent(index int) (int64, bool) {
	for ; index < c.Len(); index++ {
		value := c.Get(index)
		if value.Valid {
			return value.Value, true
		}
	}

	return 0, false
}

func absoluteToDeltas(values []Value[int64]) []Value[int64] {
	deltas := make([]Value[int64], len(values))

	var previous int64

	for i, value := range values {
		if !value.Valid {
			continue
		}

		deltas[i] = Some(value.Value - previous)
		previous = value.Value
	}

	return deltas
}

// NewPrefixColumn constructs an empty PrefixColumn.
func NewPrefixColumn() *PrefixColumn {
	return NewPrefixColumnFromValues()
}

// NewPrefixColumnFromValues constructs a PrefixColumn from values.
func NewPrefixColumnFromValues(values ...uint64) *PrefixColumn {
	present := presentValues(values)

	return &PrefixColumn{
		values: newColumn(
			Uint64Codec(),
			present,
			func(value Value[uint64]) (uint64, int64) {
				return value.Value, 0
			},
		),
	}
}

// Len returns the number of values.
func (c *PrefixColumn) Len() int { return c.values.Len() }

// Get returns the value at index.
func (c *PrefixColumn) Get(index int) uint64 {
	return c.values.Get(index).Value
}

// Prefix returns the sum of values before index in O(log chunks) time.
func (c *PrefixColumn) Prefix(index int) uint64 {
	c.values.checkRange(index, 0)
	total, _ := ropePrefix(c.values.root, index, c.values.metrics)

	return total
}

// Insert inserts values before index.
func (c *PrefixColumn) Insert(index int, values ...uint64) {
	c.values.Insert(index, presentValues(values)...)
}

// Set replaces the value at index.
func (c *PrefixColumn) Set(index int, value uint64) {
	c.values.Set(index, Some(value))
}

// Delete removes count values at index.
func (c *PrefixColumn) Delete(index, count int) {
	c.values.Delete(index, count)
}

// Splice removes deleteCount values and inserts values at index.
func (c *PrefixColumn) Splice(index, deleteCount int, values ...uint64) {
	c.values.Splice(index, deleteCount, presentValues(values)...)
}

// BatchSplice applies non-overlapping edits atomically.
func (c *PrefixColumn) BatchSplice(splices []PrefixSplice) error {
	columnSplices := make([]ColumnSplice[uint64], len(splices))
	for i, splice := range splices {
		var inserted *Column[uint64]
		if splice.Inserted != nil {
			inserted = splice.Inserted.values
		}

		columnSplices[i] = ColumnSplice[uint64]{
			Index:       splice.Index,
			DeleteCount: splice.DeleteCount,
			Inserted:    inserted,
		}
	}

	return c.values.BatchSplice(columnSplices)
}

// Clone returns an O(1), copy-on-write clone.
func (c *PrefixColumn) Clone() *PrefixColumn {
	return &PrefixColumn{values: c.values.Clone()}
}

// Values returns an independent flat snapshot.
func (c *PrefixColumn) Values() []uint64 {
	items := c.values.Values()

	values := make([]uint64, len(items))
	for i, item := range items {
		values[i] = item.Value
	}

	return values
}

// Bytes returns the canonical unsigned RLE wire representation.
func (c *PrefixColumn) Bytes() ([]byte, error) { return c.values.Bytes() }

// SaveTo writes the canonical unsigned RLE wire representation.
func (c *PrefixColumn) SaveTo(w io.Writer) (int64, error) {
	return c.values.SaveTo(w)
}

// NewBooleanColumn constructs an empty BooleanColumn.
func NewBooleanColumn() *BooleanColumn {
	return NewBooleanColumnFromValues()
}

// NewBooleanColumnFromValues constructs a BooleanColumn from values.
func NewBooleanColumnFromValues(values ...bool) *BooleanColumn {
	return &BooleanColumn{
		values: newColumn[bool](boolCodec{}, presentValues(values), nil),
	}
}

// Len returns the number of values.
func (c *BooleanColumn) Len() int { return c.values.Len() }

// Get returns the value at index.
func (c *BooleanColumn) Get(index int) bool {
	return c.values.Get(index).Value
}

// Insert inserts values before index.
func (c *BooleanColumn) Insert(index int, values ...bool) {
	c.Splice(index, 0, values...)
}

// Set replaces the value at index.
func (c *BooleanColumn) Set(index int, value bool) {
	c.Splice(index, 1, value)
}

// Delete removes count values at index.
func (c *BooleanColumn) Delete(index, count int) {
	c.Splice(index, count)
}

// Splice removes deleteCount values and inserts values at index.
func (c *BooleanColumn) Splice(index, deleteCount int, values ...bool) {
	c.values.Splice(index, deleteCount, presentValues(values)...)
	c.encoded = nil
	c.encodedOK = false
}

// BatchSplice applies non-overlapping edits atomically.
func (c *BooleanColumn) BatchSplice(splices []BooleanSplice) error {
	appendOnly := len(splices) == 1 &&
		splices[0].Index == c.Len() &&
		splices[0].DeleteCount == 0 &&
		splices[0].Inserted != nil &&
		c.encodedOK

	var appended []bool
	if appendOnly {
		appended = splices[0].Inserted.Values()
	}

	columnSplices := make([]ColumnSplice[bool], len(splices))
	for i, splice := range splices {
		var inserted *Column[bool]
		if splice.Inserted != nil {
			inserted = splice.Inserted.values
		}

		columnSplices[i] = ColumnSplice[bool]{
			Index:       splice.Index,
			DeleteCount: splice.DeleteCount,
			Inserted:    inserted,
		}
	}

	if err := c.values.BatchSplice(columnSplices); err != nil {
		return err
	}

	if appendOnly {
		for _, value := range appended {
			if value == c.tailValue {
				data := append([]byte(nil), c.encoded[:c.tailOffset]...)
				data = appendULEB(data, c.tailCount+1)
				c.encoded = data
				c.tailCount++

				continue
			}

			c.tailOffset = len(c.encoded)
			c.encoded = appendULEB(append([]byte(nil), c.encoded...), 1)
			c.tailValue = value
			c.tailCount = 1
		}
	} else {
		c.encoded = nil
		c.encodedOK = false
	}

	return nil
}

// Clone returns an O(1), copy-on-write clone.
func (c *BooleanColumn) Clone() *BooleanColumn {
	return &BooleanColumn{
		values:     c.values.Clone(),
		encoded:    c.encoded,
		tailOffset: c.tailOffset,
		tailValue:  c.tailValue,
		tailCount:  c.tailCount,
		encodedOK:  c.encodedOK,
	}
}

// Values returns an independent flat snapshot.
func (c *BooleanColumn) Values() []bool {
	items := c.values.Values()

	values := make([]bool, len(items))
	for i, item := range items {
		values[i] = item.Value
	}

	return values
}

// Bytes returns the canonical byte-aligned alternating-run representation.
func (c *BooleanColumn) Bytes() []byte {
	data, _ := c.encodedBytes()
	return append([]byte(nil), data...)
}

// SaveTo writes the canonical byte-aligned alternating-run representation.
func (c *BooleanColumn) SaveTo(w io.Writer) (int64, error) {
	data, err := c.encodedBytes()
	if err != nil {
		return 0, err
	}

	return saveBytes(w, data)
}

func (c *BooleanColumn) encodedBytes() ([]byte, error) {
	if c.encodedOK {
		return c.encoded, nil
	}

	if c.Len() == 0 {
		c.encoded = []byte{}
		c.encodedOK = true

		return c.encoded, nil
	}

	var (
		data    []byte
		current bool
		count   uint64
	)

	writeRun := func() {
		c.tailOffset = len(data)
		data = appendULEB(data, count)
		c.tailValue = current
		c.tailCount = count
	}

	ropeEach(
		c.values.root,
		func(value Value[bool]) bool {
			if value.Value == current {
				count++
				return true
			}

			writeRun()

			current = value.Value
			count = 1

			return true
		},
	)
	writeRun()

	c.encoded = data
	c.encodedOK = true

	return c.encoded, nil
}

func (boolCodec) Equal(left, right bool) bool { return left == right }
func (boolCodec) Clone(value bool) bool       { return value }
func (boolCodec) Append(dst []byte, value bool) ([]byte, error) {
	if value {
		return append(dst, 1), nil
	}

	return append(dst, 0), nil
}

func presentValues[T any](values []T) []Value[T] {
	present := make([]Value[T], len(values))
	for i, value := range values {
		present[i] = Some(value)
	}

	return present
}

type batchRange struct {
	index       int
	deleteCount int
}

func validateBatchRanges(length int, ranges []batchRange) error {
	sorted := append([]batchRange(nil), ranges...)
	slices.SortStableFunc(
		sorted,
		func(left, right batchRange) int {
			return left.index - right.index
		},
	)

	previousEnd := 0

	for i, item := range sorted {
		if item.index < 0 ||
			item.deleteCount < 0 ||
			item.index > length ||
			item.deleteCount > length-item.index {
			return fmt.Errorf("splice %d is out of bounds", i)
		}

		if i > 0 && item.index < previousEnd {
			return fmt.Errorf("splice %d overlaps its predecessor", i)
		}

		previousEnd = item.index + item.deleteCount
	}

	return nil
}
