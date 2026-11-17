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

// The writers below encode a column one item at a time. They produce exactly
// the bytes of the slice encoders (encodeRLE, encodeDelta, encodeStrings and
// encodeBooleans) without first holding the column as one entry per item,
// which is what encoding a change otherwise costs for every one of its
// columns.

type (
	rleWriterState uint8

	// rleWriter encodes a run-length column. It keeps the reference encoder's
	// canonical grouping, because change hashes cover the encoded bytes: a
	// maximal run of nulls, a maximal run of two or more equal values, or a
	// literal run that stops before a value that repeats.
	rleWriter[T comparable] struct {
		appendValue func([]byte, T) []byte
		data        []byte
		// An open literal run is written in place behind a one-byte slot for
		// its length, which is only known when the run ends.
		literalStart int
		literalCount int64
		// last is the lone or run value, or a literal run's newest value. It
		// stays unwritten until the next item shows it does not repeat.
		last  T
		count int64
		state rleWriterState
		valid bool
	}

	// deltaWriter encodes a column as the differences between its successive
	// non-null values.
	deltaWriter struct {
		runs     rleWriter[int64]
		previous int64
	}

	// booleanWriter encodes a column as alternating run lengths, starting
	// with a run of false.
	booleanWriter struct {
		data    []byte
		count   uint64
		current bool
		any     bool
	}
)

const (
	rleWriterEmpty rleWriterState = iota
	rleWriterNulls
	rleWriterLone
	rleWriterRun
	rleWriterLiteral
)

func newRLEWriter[T comparable](appendValue func([]byte, T) []byte) rleWriter[T] {
	return rleWriter[T]{appendValue: appendValue}
}

func newStringWriter() rleWriter[string] {
	return newRLEWriter(
		func(data []byte, value string) []byte {
			data = appendULEB(data, uint64(len(value)))

			return append(data, value...)
		},
	)
}

func newDeltaWriter() deltaWriter {
	return deltaWriter{runs: newRLEWriter(appendLEB)}
}

func (w *rleWriter[T]) appendNull() {
	switch w.state {
	case rleWriterNulls:
		w.count++

		return
	case rleWriterLone, rleWriterLiteral:
		w.appendLiteral(w.last)
		w.flushLiteral()
	case rleWriterRun:
		w.flushRun()
	}

	w.state = rleWriterNulls
	w.count = 1
}

func (w *rleWriter[T]) append(value T) {
	w.valid = true

	switch w.state {
	case rleWriterEmpty:
		w.state = rleWriterLone
		w.last = value
	case rleWriterNulls:
		w.flushNulls()
		w.state = rleWriterLone
		w.last = value
	case rleWriterLone:
		if value == w.last {
			w.state = rleWriterRun
			w.count = 2

			return
		}

		w.appendLiteral(w.last)
		w.state = rleWriterLiteral
		w.last = value
	case rleWriterRun:
		if value == w.last {
			w.count++

			return
		}

		w.flushRun()
		w.state = rleWriterLone
		w.last = value
	case rleWriterLiteral:
		if value == w.last {
			// The newest literal value repeats, so it opens a run instead
			// and the literal run ends before it.
			w.flushLiteral()
			w.state = rleWriterRun
			w.count = 2

			return
		}

		w.appendLiteral(w.last)
		w.last = value
	}
}

func (w *rleWriter[T]) appendOptional(value optional[T]) {
	if value.valid {
		w.append(value.value)

		return
	}

	w.appendNull()
}

func (w *rleWriter[T]) flushNulls() {
	w.data = appendLEB(w.data, 0)
	w.data = appendULEB(w.data, uint64(w.count))
}

func (w *rleWriter[T]) flushRun() {
	w.data = appendLEB(w.data, w.count)
	w.data = w.appendValue(w.data, w.last)
}

func (w *rleWriter[T]) appendLiteral(value T) {
	if w.literalCount == 0 {
		w.literalStart = len(w.data)
		w.data = append(w.data, 0)
	}

	w.data = w.appendValue(w.data, value)
	w.literalCount++
}

// flushLiteral closes the open literal run by writing its length into the
// slot reserved for it. A length that needs more than the one reserved byte
// moves the run's values up to make room.
func (w *rleWriter[T]) flushLiteral() {
	if w.literalCount == 0 {
		return
	}

	var scratch [10]byte

	length := appendLEB(scratch[:0], -w.literalCount)
	if extra := len(length) - 1; extra > 0 {
		end := len(w.data)
		w.data = append(w.data, scratch[:extra]...)
		copy(w.data[w.literalStart+len(length):], w.data[w.literalStart+1:end])
	}

	copy(w.data[w.literalStart:], length)
	w.literalCount = 0
}

// finish returns the encoded column, which is empty when every item was null.
// The writer must not be used afterwards.
func (w *rleWriter[T]) finish() []byte {
	if !w.valid {
		return nil
	}

	switch w.state {
	case rleWriterNulls:
		w.flushNulls()
	case rleWriterLone, rleWriterLiteral:
		w.appendLiteral(w.last)
		w.flushLiteral()
	case rleWriterRun:
		w.flushRun()
	}

	w.state = rleWriterEmpty

	return w.data
}

func (w *deltaWriter) append(value int64) {
	w.runs.append(value - w.previous)
	w.previous = value
}

func (w *deltaWriter) appendNull() {
	w.runs.appendNull()
}

func (w *deltaWriter) finish() []byte {
	return w.runs.finish()
}

func (w *booleanWriter) append(value bool) {
	w.any = true

	if value == w.current {
		w.count++

		return
	}

	w.data = appendULEB(w.data, w.count)
	w.current = value
	w.count = 1
}

// finish returns the encoded column, which is empty when it has no items.
func (w *booleanWriter) finish() []byte {
	if !w.any {
		return nil
	}

	return appendULEB(w.data, w.count)
}
