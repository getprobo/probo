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
	"fmt"
	"math"
	"unicode/utf8"
)

// rleRun is one run of a run-length encoded column: count copies of value, or
// count nulls when valid is false. A literal item is a run of one.
type rleRun[T any] struct {
	value T
	count uint64
	valid bool
}

// decodeRLERuns reads a run-length encoded column without expanding it. It
// walks the column exactly as decodeRLE does and reports the same errors, but
// keeps one entry per run rather than one per item, so a column that repeats a
// value across every row costs a single entry. It returns the item total.
func decodeRLERuns[T any](
	data []byte,
	decodeValue func(*reader) (T, error),
	budget *decodeBudget,
) ([]rleRun[T], uint64, error) {
	r := &reader{data: data}

	var (
		runs  []rleRun[T]
		total uint64
	)

	reserve := func(count uint64) error {
		if total > maxDecodedItems || count > maxDecodedItems-total {
			return fmt.Errorf("decoded column exceeds %d items", maxDecodedItems)
		}

		return nil
	}

	for r.remaining() > 0 {
		run, err := r.leb()
		if err != nil {
			return nil, 0, fmt.Errorf("cannot decode run length: %w", err)
		}

		switch {
		case run > 0:
			value, err := decodeValue(r)
			if err != nil {
				return nil, 0, fmt.Errorf("cannot decode repeated value: %w", err)
			}

			if err := reserve(uint64(run)); err != nil {
				return nil, 0, err
			}

			if err := chargeDecodedSliceGrowth[rleRun[T]](budget, 1); err != nil {
				return nil, 0, err
			}

			runs = append(runs, rleRun[T]{value: value, count: uint64(run), valid: true})
			total += uint64(run)
		case run == 0:
			count, err := r.uleb()
			if err != nil {
				return nil, 0, fmt.Errorf("cannot decode null run length: %w", err)
			}

			if count == 0 {
				return nil, 0, fmt.Errorf("null run cannot be empty")
			}

			if err := reserve(count); err != nil {
				return nil, 0, err
			}

			if err := chargeDecodedSliceGrowth[rleRun[T]](budget, 1); err != nil {
				return nil, 0, err
			}

			runs = append(runs, rleRun[T]{count: count})
			total += count
		default:
			if run == math.MinInt64 {
				return nil, 0, fmt.Errorf("literal run length overflows")
			}

			count := uint64(-run)
			if err := reserve(count); err != nil {
				return nil, 0, err
			}

			// A literal item takes at least one byte, so the bytes left bound
			// how many entries a hostile length can make this reserve.
			if err := chargeDecodedSliceGrowth[rleRun[T]](
				budget,
				min(count, uint64(r.remaining())),
			); err != nil {
				return nil, 0, err
			}

			for range count {
				value, err := decodeValue(r)
				if err != nil {
					return nil, 0, fmt.Errorf("cannot decode literal value: %w", err)
				}

				runs = append(runs, rleRun[T]{value: value, count: 1, valid: true})
			}

			total += count
		}
	}

	return runs, total, nil
}

func decodeULEBRuns(data []byte, budget *decodeBudget) ([]rleRun[uint64], uint64, error) {
	return decodeRLERuns(
		data,
		func(r *reader) (uint64, error) {
			return r.uleb()
		},
		budget,
	)
}

// decodeDeltaRuns reads a delta column as runs of deltas and checks, item by
// item, that the running total stays within uint64, as decoding the column in
// full would.
func decodeDeltaRuns(data []byte, budget *decodeBudget) ([]rleRun[int64], uint64, error) {
	runs, total, err := decodeRLERuns(
		data,
		func(r *reader) (int64, error) {
			return r.leb()
		},
		budget,
	)
	if err != nil {
		return nil, 0, err
	}

	var (
		previous uint64
		item     uint64
	)

	for _, run := range runs {
		if !run.valid || run.value == 0 {
			item += run.count

			continue
		}

		for range run.count {
			next, err := addSigned(previous, run.value)
			if err != nil {
				return nil, 0, fmt.Errorf("delta item %d: %w", item, err)
			}

			previous = next
			item++
		}
	}

	return runs, total, nil
}

func decodeStringRuns(data []byte, budget *decodeBudget) ([]rleRun[string], uint64, error) {
	return decodeRLERuns(
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

// requireRunItems is requireItems for a column held as runs.
func requireRunItems[T any](
	name string,
	runs []rleRun[T],
	total uint64,
	expected int,
	nullable bool,
) error {
	if total != uint64(expected) {
		return fmt.Errorf("%s column has %d items, expected %d", name, total, expected)
	}

	if nullable {
		return nil
	}

	var item uint64

	for _, run := range runs {
		if !run.valid {
			return fmt.Errorf("%s column item %d is null", name, item)
		}

		item += run.count
	}

	return nil
}

// runCursor yields a run-held column one item at a time. Once the runs are
// spent, or when the column is absent, every further item is null.
type runCursor[T any] struct {
	runs []rleRun[T]
	used uint64
}

func (c *runCursor[T]) next() optional[T] {
	if len(c.runs) == 0 {
		return optional[T]{}
	}

	run := &c.runs[0]
	item := optional[T]{value: run.value, valid: run.valid}

	c.used++
	if c.used == run.count {
		c.runs = c.runs[1:]
		c.used = 0
	}

	return item
}

// deltaCursor yields the running totals of a delta column whose deltas
// decodeDeltaRuns has already checked.
type deltaCursor struct {
	deltas   runCursor[int64]
	previous uint64
}

func (c *deltaCursor) next() optional[uint64] {
	delta := c.deltas.next()
	if !delta.valid {
		return optional[uint64]{}
	}

	c.previous += uint64(delta.value)

	return optional[uint64]{value: c.previous, valid: true}
}

// optionalULEBRuns reads a nullable column that may be absent, in which case
// every item is null.
func optionalULEBRuns(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
	budget *decodeBudget,
) (runCursor[uint64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return runCursor[uint64]{}, nil
	}

	runs, total, err := decodeULEBRuns(data, budget)
	if err != nil {
		return runCursor[uint64]{}, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireRunItems(name, runs, total, expected, true); err != nil {
		return runCursor[uint64]{}, err
	}

	return runCursor[uint64]{runs: runs}, nil
}

func optionalDeltaRuns(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
	budget *decodeBudget,
) (*deltaCursor, error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return &deltaCursor{}, nil
	}

	runs, total, err := decodeDeltaRuns(data, budget)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireRunItems(name, runs, total, expected, true); err != nil {
		return nil, err
	}

	return &deltaCursor{deltas: runCursor[int64]{runs: runs}}, nil
}

func optionalStringRuns(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
	budget *decodeBudget,
) (runCursor[string], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return runCursor[string]{}, nil
	}

	runs, total, err := decodeStringRuns(data, budget)
	if err != nil {
		return runCursor[string]{}, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireRunItems(name, runs, total, expected, true); err != nil {
		return runCursor[string]{}, err
	}

	return runCursor[string]{runs: runs}, nil
}
