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
	"slices"
)

type (
	ropeMetrics[T any] func(T) (uint64, int64)

	ropeSplice[T any] struct {
		index       int
		deleteCount int
		inserted    *ropeNode[T]
	}

	// ropeNode is immutable after construction. Clone therefore shares the
	// complete tree, while edits copy only search paths and boundary leaves.
	ropeNode[T any] struct {
		left   *ropeNode[T]
		right  *ropeNode[T]
		items  []T
		len    int
		height int
		uint   uint64
		sint   int64
	}
)

func ropeFrom[T any](
	items []T,
	chunkSize int,
	clone func(T) T,
	metrics ropeMetrics[T],
) *ropeNode[T] {
	if len(items) == 0 {
		return nil
	}

	leaves := make([]*ropeNode[T], 0, (len(items)+chunkSize-1)/chunkSize)
	for len(items) > 0 {
		count := min(len(items), chunkSize)

		owned := make([]T, count)
		for i, item := range items[:count] {
			owned[i] = clone(item)
		}

		leaves = append(leaves, ropeLeaf(owned, metrics))
		items = items[count:]
	}

	var build func(start, end int) *ropeNode[T]

	build = func(start, end int) *ropeNode[T] {
		if end-start == 1 {
			return leaves[start]
		}

		middle := start + (end-start)/2

		return ropeBranch(build(start, middle), build(middle, end))
	}

	return build(0, len(leaves))
}

func ropeLeaf[T any](items []T, metrics ropeMetrics[T]) *ropeNode[T] {
	if len(items) == 0 {
		return nil
	}

	node := &ropeNode[T]{
		items:  items,
		len:    len(items),
		height: 1,
	}
	if metrics != nil {
		for _, item := range items {
			uintValue, intValue := metrics(item)
			node.uint += uintValue
			node.sint += intValue
		}
	}

	return node
}

func ropeBranch[T any](left, right *ropeNode[T]) *ropeNode[T] {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	return &ropeNode[T]{
		left:   left,
		right:  right,
		len:    left.len + right.len,
		height: max(left.height, right.height) + 1,
		uint:   left.uint + right.uint,
		sint:   left.sint + right.sint,
	}
}

func ropeConcat[T any](
	left, right *ropeNode[T],
	chunkSize int,
	metrics ropeMetrics[T],
) *ropeNode[T] {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	if left.items != nil && right.items != nil &&
		len(left.items)+len(right.items) <= chunkSize {
		items := make([]T, 0, len(left.items)+len(right.items))
		items = append(items, left.items...)
		items = append(items, right.items...)

		return ropeLeaf(items, metrics)
	}

	if left.height > right.height+1 {
		return ropeBalance(
			left.left,
			ropeConcat(left.right, right, chunkSize, metrics),
		)
	}

	if right.height > left.height+1 {
		return ropeBalance(
			ropeConcat(left, right.left, chunkSize, metrics),
			right.right,
		)
	}

	return ropeBranch(left, right)
}

func ropeBalance[T any](left, right *ropeNode[T]) *ropeNode[T] {
	if left.height > right.height+1 {
		if ropeHeight(left.left) >= ropeHeight(left.right) {
			return ropeBranch(left.left, ropeBranch(left.right, right))
		}

		pivot := left.right

		return ropeBranch(
			ropeBranch(left.left, pivot.left),
			ropeBranch(pivot.right, right),
		)
	}

	if right.height > left.height+1 {
		if ropeHeight(right.right) >= ropeHeight(right.left) {
			return ropeBranch(ropeBranch(left, right.left), right.right)
		}

		pivot := right.left

		return ropeBranch(
			ropeBranch(left, pivot.left),
			ropeBranch(pivot.right, right.right),
		)
	}

	return ropeBranch(left, right)
}

func ropeHeight[T any](root *ropeNode[T]) int {
	if root == nil {
		return 0
	}

	return root.height
}

func ropeSplit[T any](
	root *ropeNode[T],
	index int,
	chunkSize int,
	metrics ropeMetrics[T],
) (*ropeNode[T], *ropeNode[T]) {
	if root == nil {
		return nil, nil
	}

	if index == 0 {
		return nil, root
	}

	if index == root.len {
		return root, nil
	}

	if root.items != nil {
		left := append([]T(nil), root.items[:index]...)
		right := append([]T(nil), root.items[index:]...)

		return ropeLeaf(left, metrics), ropeLeaf(right, metrics)
	}

	if index < root.left.len {
		left, middle := ropeSplit(root.left, index, chunkSize, metrics)
		return left, ropeConcat(middle, root.right, chunkSize, metrics)
	}

	if index == root.left.len {
		return root.left, root.right
	}

	middle, right := ropeSplit(
		root.right,
		index-root.left.len,
		chunkSize,
		metrics,
	)

	return ropeConcat(root.left, middle, chunkSize, metrics), right
}

func ropeBatchSplice[T any](
	root *ropeNode[T],
	splices []ropeSplice[T],
	chunkSize int,
	metrics ropeMetrics[T],
) (*ropeNode[T], error) {
	if len(splices) == 0 {
		return root, nil
	}

	length := 0
	if root != nil {
		length = root.len
	}

	sorted := append([]ropeSplice[T](nil), splices...)
	slices.SortStableFunc(
		sorted,
		func(left, right ropeSplice[T]) int {
			return left.index - right.index
		},
	)

	previousEnd := 0

	for i, splice := range sorted {
		if splice.index < 0 ||
			splice.deleteCount < 0 ||
			splice.index > length ||
			splice.deleteCount > length-splice.index {
			return nil, fmt.Errorf("hexane: batch splice %d is out of bounds", i)
		}

		if i > 0 && splice.index < previousEnd {
			return nil, fmt.Errorf("hexane: batch splice %d overlaps its predecessor", i)
		}

		previousEnd = splice.index + splice.deleteCount
	}

	var result *ropeNode[T]

	remaining := root
	cursor := 0

	for _, splice := range sorted {
		var retained *ropeNode[T]

		retained, remaining = ropeSplit(
			remaining,
			splice.index-cursor,
			chunkSize,
			metrics,
		)
		_, remaining = ropeSplit(
			remaining,
			splice.deleteCount,
			chunkSize,
			metrics,
		)
		result = ropeConcat(result, retained, chunkSize, metrics)
		result = ropeConcat(result, splice.inserted, chunkSize, metrics)
		cursor = splice.index + splice.deleteCount
	}

	return ropeConcat(result, remaining, chunkSize, metrics), nil
}

func ropeGet[T any](root *ropeNode[T], index int) T {
	for root.items == nil {
		if index < root.left.len {
			root = root.left
			continue
		}

		index -= root.left.len
		root = root.right
	}

	return root.items[index]
}

func ropeEach[T any](root *ropeNode[T], yield func(T) bool) bool {
	if root == nil {
		return true
	}

	if root.items != nil {
		for _, item := range root.items {
			if !yield(item) {
				return false
			}
		}

		return true
	}

	return ropeEach(root.left, yield) && ropeEach(root.right, yield)
}

func ropeEachLeaf[T any](root *ropeNode[T], yield func([]T) bool) bool {
	if root == nil {
		return true
	}

	if root.items != nil {
		return yield(root.items)
	}

	return ropeEachLeaf(root.left, yield) && ropeEachLeaf(root.right, yield)
}

func ropePrefix[T any](
	root *ropeNode[T],
	index int,
	metrics ropeMetrics[T],
) (uint64, int64) {
	if root == nil || index == 0 {
		return 0, 0
	}

	if index == root.len {
		return root.uint, root.sint
	}

	if root.items != nil {
		var (
			uintValue uint64
			intValue  int64
		)

		for _, item := range root.items[:index] {
			itemUint, itemInt := metrics(item)
			uintValue += itemUint
			intValue += itemInt
		}

		return uintValue, intValue
	}

	if index <= root.left.len {
		return ropePrefix(root.left, index, metrics)
	}

	rightUint, rightInt := ropePrefix(
		root.right,
		index-root.left.len,
		metrics,
	)

	return root.left.uint + rightUint, root.left.sint + rightInt
}
