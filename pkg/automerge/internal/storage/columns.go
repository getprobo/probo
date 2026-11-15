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

package storage

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"slices"
	"sync"
	"unicode/utf8"
	"unsafe"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

var deflateWriterPool = sync.Pool{
	New: func() any {
		writer, err := flate.NewWriter(io.Discard, flate.BestCompression)
		if err != nil {
			panic(err)
		}

		return writer
	},
}

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

	decodeBudget struct {
		used uint64
	}
)

const (
	maxDecodedItems       = 100_000_000
	maxDecodedMemoryBytes = 512 << 20
	maxInflatedBytes      = 512 << 20
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

func decodeRLE[T any](
	data []byte,
	decodeValue func(*reader) (T, error),
	budget *decodeBudget,
) ([]optional[T], error) {
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

			if err := appendRepeated(
				&values,
				optional[T]{value: value, valid: true},
				uint64(run),
				budget,
			); err != nil {
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

			if err := appendRepeated(&values, optional[T]{}, count, budget); err != nil {
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

			values, err = growCharged(values, count, budget)
			if err != nil {
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

	return settleCharge(values, budget), nil
}

// settleCharge leaves a finished RLE column charged for exactly len(values),
// the amount every release site gives back. A large surplus is compacted so the
// memory it held is really freed; a small one (under a sixteenth of the column)
// is dropped from the budget instead of paying for another copy.
func settleCharge[T any](values []optional[T], budget *decodeBudget) []optional[T] {
	surplus := cap(values) - len(values)
	if surplus == 0 {
		return values
	}

	// The copy and the old array coexist, so the copy is charged first. When it
	// does not fit, keep the slice and only drop the surplus from the charge.
	if surplus > len(values)/16 &&
		chargeDecoded[optional[T]](budget, uint64(len(values))) == nil {
		exact := make([]optional[T], len(values))
		copy(exact, values)
		releaseDecoded[optional[T]](budget, uint64(cap(values)))

		return exact
	}

	releaseDecoded[optional[T]](budget, uint64(surplus))

	return values
}

func appendRepeated[T any](
	values *[]optional[T],
	value optional[T],
	count uint64,
	budget *decodeBudget,
) error {
	if err := reserveItems(len(*values), count); err != nil {
		return err
	}

	// One growth for the whole run. Growing element by element would reallocate
	// a long run (every text operation sharing an object and action) as the
	// slice doubled.
	grown, err := growCharged(*values, count, budget)
	if err != nil {
		return err
	}

	for range count {
		grown = append(grown, value)
	}

	*values = grown

	return nil
}

const (
	// Go's small-object cutoff and heap page. Large slice growth rounds the
	// allocation up to the next page; the bound includes one extra page so a
	// malloc header cannot land past it.
	smallAllocMax = 32 << 10
	heapPageSize  = 8 << 10
)

// growCharged reserves the capacity slices.Grow will allocate before growing.
// Repeated short runs double (then grow by about a quarter), so charging only
// the values added by this run under-counts the backing array. Surplus in the
// bound is released once the actual cap is known.
func growCharged[T any](
	values []optional[T],
	count uint64,
	budget *decodeBudget,
) ([]optional[T], error) {
	if count > uint64(math.MaxInt) {
		return nil, fmt.Errorf("decoded column exceeds %d items", maxDecodedItems)
	}

	additional := int(count)
	if additional <= cap(values)-len(values) {
		return values, nil
	}

	elemSize := int(unsafe.Sizeof(optional[T]{}))
	if elemSize == 0 {
		return slices.Grow(values, additional), nil
	}

	bound, err := sliceCapacityBound(cap(values), len(values), additional, elemSize)
	if err != nil {
		return nil, err
	}

	if bound < cap(values) {
		return nil, fmt.Errorf("decoded slice capacity overflows")
	}

	if err := chargeDecoded[optional[T]](budget, uint64(bound-cap(values))); err != nil {
		return nil, err
	}

	grown := slices.Grow(values, additional)
	switch {
	case cap(grown) > bound:
		if err := chargeDecoded[optional[T]](budget, uint64(cap(grown)-bound)); err != nil {
			return nil, err
		}
	case cap(grown) < bound:
		releaseDecoded[optional[T]](budget, uint64(bound-cap(grown)))
	}

	return grown, nil
}

// sliceCapacityBound estimates the cap slices.Grow returns. nextSliceCap
// matches runtime.nextslicecap; large results are then rounded up to a heap
// page, plus one page of slack.
func sliceCapacityBound(oldCap int, length int, additional int, elemSize int) (int, error) {
	spare := oldCap - length
	if additional <= spare {
		return oldCap, nil
	}

	extra := additional - spare
	if oldCap > math.MaxInt-extra {
		return 0, fmt.Errorf("decoded slice capacity overflows")
	}

	newLen := oldCap + extra
	newCap := nextSliceCap(newLen, oldCap)

	if elemSize <= 0 || newCap > math.MaxInt/elemSize {
		return 0, fmt.Errorf("decoded slice capacity overflows")
	}

	rounded, err := roundedAllocBytes(newCap * elemSize)
	if err != nil {
		return 0, err
	}

	elems := rounded / elemSize
	if elems < newCap {
		return newCap, nil
	}

	return elems, nil
}

// sizeClassSizes are the classes Go uses for allocations up to 32KiB.
var sizeClassSizes = [...]int{
	8, 16, 24, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192,
	208, 224, 240, 256, 288, 320, 352, 384, 416, 448, 480, 512,
	576, 640, 704, 768, 896, 1024, 1152, 1280, 1408, 1536, 1792,
	2048, 2304, 2688, 3072, 3200, 3456, 4096, 4864, 5376, 6144,
	6528, 6784, 6912, 8192, 9472, 9728, 10240, 10880, 12288,
	13568, 14336, 16384, 18432, 19072, 20480, 21760, 24576, 27264,
	28672, 32768,
}

// roundedAllocBytes bounds the block size for a payload of the given size.
// Pointer values add a malloc header once they pass the header threshold, which
// can move the block into the next size class. The header is assumed for every
// payload so a noscan value only overestimates, and large blocks include one
// extra page.
func roundedAllocBytes(payload int) (int, error) {
	const mallocHeader = 8

	if payload <= 0 {
		return 0, nil
	}

	if payload > smallAllocMax-mallocHeader {
		if payload > math.MaxInt-(heapPageSize-1)-heapPageSize {
			return 0, fmt.Errorf("decoded slice capacity overflows")
		}

		aligned := (payload + heapPageSize - 1) &^ (heapPageSize - 1)

		return aligned + heapPageSize, nil
	}

	// PtrSize * PtrBits, the size at which Go starts adding a malloc header.
	minHeader := int(unsafe.Sizeof(uintptr(0))) * int(unsafe.Sizeof(uintptr(0))) * 8
	req := payload

	header := 0
	if payload > minHeader {
		header = mallocHeader
		req += header
	}

	class := roundSizeClass(req)
	if class < header {
		return payload, nil
	}

	return class - header, nil
}

func roundSizeClass(size int) int {
	index, _ := slices.BinarySearch(sizeClassSizes[:], size)
	if index >= len(sizeClassSizes) {
		return sizeClassSizes[len(sizeClassSizes)-1]
	}

	return sizeClassSizes[index]
}

// nextSliceCap matches runtime.nextslicecap: a request longer than twice the
// current cap is taken as-is, smaller slices double until 256 elements, and
// larger slices grow by about a quarter.
func nextSliceCap(newLen int, oldCap int) int {
	if uint(newLen) > uint(oldCap)+uint(oldCap) {
		return newLen
	}

	const threshold = 256
	if oldCap < threshold {
		return oldCap + oldCap
	}

	newCap := oldCap
	for {
		step := (newCap + 3*threshold) >> 2
		if step <= 0 || newCap > math.MaxInt-step {
			return newLen
		}

		newCap += step
		if newCap >= newLen {
			return newCap
		}
	}
}

func chargeDecoded[T any](budget *decodeBudget, count uint64) error {
	if budget == nil || count == 0 {
		return nil
	}

	size := uint64(unsafe.Sizeof(*new(T)))
	if size != 0 && count > math.MaxUint64/size {
		return fmt.Errorf("decoded memory size overflows uint64")
	}

	bytes := size * count
	if budget.used > maxDecodedMemoryBytes ||
		bytes > maxDecodedMemoryBytes-budget.used {
		return fmt.Errorf("decoded columns exceed %d bytes", maxDecodedMemoryBytes)
	}

	budget.used += bytes

	return nil
}

func chargeDecodedSliceGrowth[T any](budget *decodeBudget, count uint64) error {
	if count > math.MaxUint64/2 {
		return fmt.Errorf("decoded slice growth overflows uint64")
	}

	// append and slices.Grow may reserve more than the requested logical length.
	// Charging twice the added elements bounds Go's geometric slice growth.
	return chargeDecoded[T](budget, count*2)
}

func chargeDecodedBytes(budget *decodeBudget, bytes uint64) error {
	if budget == nil || bytes == 0 {
		return nil
	}

	if budget.used > maxDecodedMemoryBytes ||
		bytes > maxDecodedMemoryBytes-budget.used {
		return fmt.Errorf("decoded columns exceed %d bytes", maxDecodedMemoryBytes)
	}

	budget.used += bytes

	return nil
}

func releaseDecodedBytes(budget *decodeBudget, bytes uint64) {
	if budget == nil {
		return
	}

	if bytes > budget.used {
		panic("decoded memory budget underflow")
	}

	budget.used -= bytes
}

func budgetMark(budget *decodeBudget) uint64 {
	if budget == nil {
		return 0
	}

	return budget.used
}

func releaseToMark(budget *decodeBudget, mark uint64) {
	if budget == nil {
		return
	}

	if budget.used < mark {
		panic("decoded memory budget underflow")
	}

	budget.used = mark
}

func releaseDecoded[T any](budget *decodeBudget, count uint64) {
	if budget == nil || count == 0 {
		return
	}

	size := uint64(unsafe.Sizeof(*new(T)))
	if size != 0 && count > math.MaxUint64/size {
		panic("decoded memory budget underflow")
	}

	releaseDecodedBytes(budget, size*count)
}

func releaseDecodedSliceGrowth[T any](budget *decodeBudget, count uint64) {
	if count > math.MaxUint64/2 {
		panic("decoded memory budget underflow")
	}

	releaseDecoded[T](budget, count*2)
}

func chargeDecodedMap[K comparable, V any](budget *decodeBudget, count uint64) error {
	type entry struct {
		key   K
		value V
	}

	size := uint64(unsafe.Sizeof(entry{}))
	if size != 0 && count > math.MaxUint64/size/2 {
		return fmt.Errorf("decoded map size overflows uint64")
	}

	// Go maps allocate buckets, overflow storage, and bookkeeping in addition to
	// their logical entries. Charging twice the entry size is a conservative
	// bound for the maps used by the decoder.
	return chargeDecodedBytes(budget, size*count*2)
}

func reserveItems(existing int, additional uint64) error {
	if additional > maxDecodedItems || uint64(existing)+additional > maxDecodedItems {
		return fmt.Errorf("decoded column exceeds %d items", maxDecodedItems)
	}

	return nil
}

func decodeULEBColumn(data []byte) ([]optional[uint64], error) {
	return decodeULEBColumnWithBudget(data, nil)
}

func decodeULEBColumnWithBudget(
	data []byte,
	budget *decodeBudget,
) ([]optional[uint64], error) {
	return decodeRLE(
		data,
		func(r *reader) (uint64, error) {
			return r.uleb()
		},
		budget,
	)
}

func decodeDeltaColumn(data []byte) ([]optional[uint64], error) {
	return decodeDeltaColumnWithBudget(data, nil)
}

func decodeDeltaColumnWithBudget(
	data []byte,
	budget *decodeBudget,
) ([]optional[uint64], error) {
	deltas, err := decodeRLE(
		data,
		func(r *reader) (int64, error) {
			return r.leb()
		},
		budget,
	)
	if err != nil {
		return nil, err
	}

	if err := chargeDecoded[optional[uint64]](budget, uint64(len(deltas))); err != nil {
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
	return decodeSignedDeltaColumnWithBudget(data, nil)
}

func decodeSignedDeltaColumnWithBudget(
	data []byte,
	budget *decodeBudget,
) ([]optional[int64], error) {
	deltas, err := decodeRLE(
		data,
		func(r *reader) (int64, error) {
			return r.leb()
		},
		budget,
	)
	if err != nil {
		return nil, err
	}

	if err := chargeDecoded[optional[int64]](budget, uint64(len(deltas))); err != nil {
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
	return decodeStringColumnWithBudget(data, nil)
}

func decodeStringColumnWithBudget(
	data []byte,
	budget *decodeBudget,
) ([]optional[string], error) {
	return decodeRLE(
		data,
		func(r *reader) (string, error) {
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

			if err := chargeDecodedBytes(budget, uint64(len(value))); err != nil {
				return "", err
			}

			return string(value), nil
		},
		budget,
	)
}

func decodeBooleanColumn(data []byte, expected int) ([]bool, error) {
	return decodeBooleanColumnWithBudget(data, expected, nil)
}

func decodeBooleanColumnWithBudget(
	data []byte,
	expected int,
	budget *decodeBudget,
) ([]bool, error) {
	r := &reader{data: data}

	if err := chargeDecoded[bool](budget, uint64(expected)); err != nil {
		return nil, err
	}

	values := make([]bool, 0, expected)
	current := false

	for r.remaining() > 0 {
		count, err := r.uleb()
		if err != nil {
			return nil, fmt.Errorf("cannot decode boolean run: %w", err)
		}

		if count > uint64(expected-len(values)) {
			return nil, fmt.Errorf("boolean column exceeds expected %d items", expected)
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

func parseColumnMetadata(
	r *reader,
	allowCompressed bool,
	budget *decodeBudget,
) ([]columnMeta, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, fmt.Errorf("cannot decode column count: %w", err)
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("column count %d exceeds limit", count)
	}

	if err := chargeDecoded[columnMeta](budget, count); err != nil {
		return nil, err
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

func readColumns(
	r *reader,
	metadata []columnMeta,
	budget *decodeBudget,
) (map[uint32]column, error) {
	if err := chargeDecodedMap[uint32, column](budget, uint64(len(metadata))); err != nil {
		return nil, err
	}

	columns := make(map[uint32]column, len(metadata))
	for i, meta := range metadata {
		data, err := r.bytes(meta.length)
		if err != nil {
			return nil, fmt.Errorf("cannot read column %d: %w", i, err)
		}

		if meta.compressed {
			data, err = inflate(data, budget)
			if err != nil {
				return nil, fmt.Errorf("cannot inflate column %d: %w", i, err)
			}
		}

		columns[meta.normalized] = column{
			specification: meta.specification,
			data:          data,
		}
	}

	return columns, nil
}

// deflate compresses data with raw DEFLATE at best compression, matching the
// stream inflate reads. It is used to produce compressed change chunks.
func deflate(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	writer := deflateWriterPool.Get().(*flate.Writer)

	writer.Reset(&buffer)
	defer func() {
		writer.Reset(io.Discard)
		deflateWriterPool.Put(writer)
	}()

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("cannot write DEFLATE stream: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("cannot finish DEFLATE stream: %w", err)
	}

	return buffer.Bytes(), nil
}

func inflate(data []byte, budget *decodeBudget) ([]byte, error) {
	compressed := flate.NewReader(bytes.NewReader(data))
	defer func() { _ = compressed.Close() }()

	const readSize = 32 << 10
	if err := chargeDecodedBytes(budget, readSize); err != nil {
		return nil, err
	}

	buffer := make([]byte, readSize)
	defer releaseDecodedBytes(budget, readSize)

	var output []byte

	for {
		count, err := compressed.Read(buffer)
		if count > 0 {
			if len(output) > maxInflatedBytes-count {
				return nil, fmt.Errorf("inflated data exceeds %d bytes", maxInflatedBytes)
			}

			required := len(output) + count
			if required > cap(output) {
				capacity := max(required, max(readSize, cap(output)*2))

				capacity = min(capacity, maxInflatedBytes)
				if err := chargeDecodedBytes(budget, uint64(capacity)); err != nil {
					return nil, err
				}

				grown := make([]byte, len(output), capacity)
				copy(grown, output)
				releaseDecodedBytes(budget, uint64(cap(output)))
				output = grown
			}

			output = append(output, buffer[:count]...)
		}

		if err == io.EOF {
			return output, nil
		}

		if err != nil {
			return nil, fmt.Errorf("cannot read DEFLATE stream: %w", err)
		}
	}
}

func decodeScalars(metaData, rawData []byte, expected int) ([]optional[opset.Scalar], error) {
	values, _, err := decodeScalarsInternal(metaData, rawData, expected, nil, false)

	return values, err
}

func decodeScalarsWithRaw(
	metaData []byte,
	rawData []byte,
	expected int,
	budget *decodeBudget,
) ([]optional[opset.Scalar], [][]byte, error) {
	return decodeScalarsInternal(metaData, rawData, expected, budget, true)
}

func decodeScalarsInternal(
	metaData []byte,
	rawData []byte,
	expected int,
	budget *decodeBudget,
	retainRaw bool,
) ([]optional[opset.Scalar], [][]byte, error) {
	metadata, err := decodeULEBColumnWithBudget(metaData, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode value metadata: %w", err)
	}

	if len(metadata) != expected {
		return nil, nil, fmt.Errorf(
			"value metadata has %d items, expected %d",
			len(metadata),
			expected,
		)
	}

	raw := &reader{data: rawData}

	if err := chargeDecoded[optional[opset.Scalar]](budget, uint64(expected)); err != nil {
		return nil, nil, err
	}

	if retainRaw {
		if err := chargeDecoded[[]byte](budget, uint64(expected)); err != nil {
			return nil, nil, err
		}
	}

	values := make([]optional[opset.Scalar], expected)

	var rawValues [][]byte
	if retainRaw {
		rawValues = make([][]byte, expected)
	}

	for i, item := range metadata {
		if !item.valid {
			continue
		}

		scalarType := opset.ScalarType(item.value & 0x0f)
		length := item.value >> 4

		valueBytes, err := raw.bytes(length)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot read scalar %d: %w", i, err)
		}

		scalar, err := decodeScalarWithBudget(scalarType, valueBytes, budget)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode scalar %d: %w", i, err)
		}

		values[i] = optional[opset.Scalar]{value: scalar, valid: true}
		if retainRaw {
			rawValues[i] = valueBytes
		}
	}

	if raw.remaining() != 0 {
		return nil, nil, fmt.Errorf("value column has %d trailing bytes", raw.remaining())
	}

	return values, rawValues, nil
}

func decodeScalar(scalarType opset.ScalarType, data []byte) (opset.Scalar, error) {
	return decodeScalarWithBudget(scalarType, data, nil)
}

func decodeScalarWithBudget(
	scalarType opset.ScalarType,
	data []byte,
	budget *decodeBudget,
) (opset.Scalar, error) {
	scalar := opset.Scalar{Type: scalarType}
	switch scalarType {
	case opset.ScalarNull:
		if len(data) != 0 {
			return opset.Scalar{}, fmt.Errorf("null has length %d, expected 0", len(data))
		}
	case opset.ScalarFalse, opset.ScalarTrue:
		if len(data) != 0 {
			return opset.Scalar{}, fmt.Errorf("boolean has length %d, expected 0", len(data))
		}

		scalar.Bool = scalarType == opset.ScalarTrue
	case opset.ScalarUint:
		r := &reader{data: data}

		value, err := r.uleb()
		if err != nil || r.remaining() != 0 {
			return opset.Scalar{}, fmt.Errorf("invalid unsigned integer scalar")
		}

		scalar.Uint = value
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		r := &reader{data: data}

		value, err := r.leb()
		if err != nil || r.remaining() != 0 {
			return opset.Scalar{}, fmt.Errorf("invalid signed integer scalar")
		}

		scalar.Int = value
	case opset.ScalarFloat64:
		if len(data) != 8 {
			return opset.Scalar{}, fmt.Errorf("float has length %d, expected 8", len(data))
		}

		scalar.Float = math.Float64frombits(binary.LittleEndian.Uint64(data))
	case opset.ScalarString:
		if !utf8.Valid(data) {
			return opset.Scalar{}, fmt.Errorf("string scalar is not valid UTF-8")
		}

		if err := chargeDecodedBytes(budget, uint64(len(data))); err != nil {
			return opset.Scalar{}, err
		}

		scalar.String = string(data)
	case opset.ScalarBytes:
		scalar.Bytes = data
	default:
		scalar.Raw = data
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

func copyHash(data []byte) opset.ChangeHash {
	var hash opset.ChangeHash
	copy(hash[:], data)

	return hash
}

// Deflate compresses an Automerge change body using the wire-format codec.
func Deflate(data []byte) ([]byte, error) { return deflate(data) }
