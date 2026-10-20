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
	// RawSplice describes one edit in original-column coordinates.
	RawSplice struct {
		Index       int
		DeleteCount int
		Inserted    *RawColumn
	}

	// RawColumn is an unencoded byte arena backed by an immutable AVL rope.
	// Clones share all chunks; edits copy only tree paths and boundary chunks.
	RawColumn struct {
		root *ropeNode[byte]
	}
)

const (
	rawChunkSize = 4096
)

// NewRawColumn constructs an empty RawColumn.
func NewRawColumn() *RawColumn {
	return &RawColumn{}
}

// NewRawColumnFromBytes constructs a RawColumn by copying data.
func NewRawColumnFromBytes(data []byte) *RawColumn {
	return &RawColumn{
		root: ropeFrom(
			data,
			rawChunkSize,
			func(value byte) byte { return value },
			nil,
		),
	}
}

// Len returns the number of bytes.
func (c *RawColumn) Len() int {
	if c.root == nil {
		return 0
	}

	return c.root.len
}

// Get returns the byte at index.
func (c *RawColumn) Get(index int) byte {
	c.checkIndex(index)
	return ropeGet(c.root, index)
}

// Slice returns an independent copy of the half-open byte range [start, end).
func (c *RawColumn) Slice(start, end int) []byte {
	c.checkRange(start, end-start)
	data := make([]byte, 0, end-start)
	appendRawRange(c.root, start, end, &data)

	return data
}

// Insert inserts bytes before index.
func (c *RawColumn) Insert(index int, values ...byte) {
	c.Splice(index, 0, values)
}

// Set replaces the byte at index.
func (c *RawColumn) Set(index int, value byte) {
	c.Splice(index, 1, []byte{value})
}

// Delete removes count bytes at index.
func (c *RawColumn) Delete(index, count int) {
	c.Splice(index, count, nil)
}

// Splice removes deleteCount bytes and inserts values at index.
func (c *RawColumn) Splice(index, deleteCount int, values []byte) {
	c.checkRange(index, deleteCount)

	if deleteCount == 0 && len(values) == 0 {
		return
	}

	left, rest := ropeSplit(c.root, index, rawChunkSize, nil)
	_, right := ropeSplit(rest, deleteCount, rawChunkSize, nil)
	inserted := ropeFrom(
		values,
		rawChunkSize,
		func(value byte) byte { return value },
		nil,
	)
	c.root = ropeConcat(
		ropeConcat(left, inserted, rawChunkSize, nil),
		right,
		rawChunkSize,
		nil,
	)
}

// BatchSplice applies non-overlapping edits atomically. Indexes refer to the
// column state before any edit in the batch.
func (c *RawColumn) BatchSplice(splices []RawSplice) error {
	ropeSplices := make([]ropeSplice[byte], len(splices))
	for i, splice := range splices {
		var inserted *ropeNode[byte]
		if splice.Inserted != nil {
			inserted = splice.Inserted.root
		}

		ropeSplices[i] = ropeSplice[byte]{
			index:       splice.Index,
			deleteCount: splice.DeleteCount,
			inserted:    inserted,
		}
	}

	root, err := ropeBatchSplice(c.root, ropeSplices, rawChunkSize, nil)
	if err != nil {
		return fmt.Errorf("cannot splice raw column batch: %w", err)
	}

	c.root = root

	return nil
}

// Clone returns an O(1), copy-on-write clone.
func (c *RawColumn) Clone() *RawColumn {
	return &RawColumn{root: c.root}
}

// Bytes returns an independent contiguous copy of the raw bytes.
func (c *RawColumn) Bytes() []byte {
	data := make([]byte, 0, c.Len())
	ropeEachLeaf(
		c.root,
		func(chunk []byte) bool {
			data = append(data, chunk...)
			return true
		},
	)

	return data
}

// SaveTo writes the raw chunks without framing or flattening.
func (c *RawColumn) SaveTo(w io.Writer) (int64, error) {
	var (
		written int64
		saveErr error
	)

	ropeEachLeaf(
		c.root,
		func(chunk []byte) bool {
			count, err := saveBytes(w, chunk)
			written += count

			if err != nil {
				saveErr = err
				return false
			}

			return true
		},
	)

	return written, saveErr
}

func appendRawRange(
	root *ropeNode[byte],
	start int,
	end int,
	destination *[]byte,
) {
	if root == nil || start == end {
		return
	}

	if root.items != nil {
		*destination = append(*destination, root.items[start:end]...)
		return
	}

	if start < root.left.len {
		appendRawRange(root.left, start, min(end, root.left.len), destination)
	}

	if end > root.left.len {
		appendRawRange(
			root.right,
			max(0, start-root.left.len),
			end-root.left.len,
			destination,
		)
	}
}

func (c *RawColumn) checkIndex(index int) {
	if index < 0 || index >= c.Len() {
		panic("hexane: raw column index out of range")
	}
}

func (c *RawColumn) checkRange(index, count int) {
	length := c.Len()
	if index < 0 || count < 0 || index > length || count > length-index {
		panic("hexane: raw column slice bounds out of range")
	}
}
