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
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestRope_RandomizedSpliceMaintainsBalance(t *testing.T) {
	t.Parallel()

	const chunkSize = 64

	model := make([]int, 10_000)
	for i := range model {
		model[i] = i
	}

	root := ropeFrom(
		model,
		chunkSize,
		func(value int) int { return value },
		func(value int) (uint64, int64) { return uint64(value), int64(value) },
	)
	random := rand.New(rand.NewPCG(110, 120))

	for step := range 2_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if index < len(model) {
			deleteCount = random.IntN(min(100, len(model)-index) + 1)
		}

		inserted := make([]int, random.IntN(100))
		for i := range inserted {
			inserted[i] = random.IntN(10_000)
		}

		left, rest := ropeSplit(root, index, chunkSize, testIntMetrics)
		_, right := ropeSplit(rest, deleteCount, chunkSize, testIntMetrics)
		middle := ropeFrom(
			inserted,
			chunkSize,
			func(value int) int { return value },
			testIntMetrics,
		)
		root = ropeConcat(
			ropeConcat(left, middle, chunkSize, testIntMetrics),
			right,
			chunkSize,
			testIntMetrics,
		)
		model = testSpliceInts(model, index, deleteCount, inserted)

		length, height, uintSum, intSum := assertRopeNode(t, root, chunkSize)
		if length != len(model) {
			t.Fatalf("step %d: root length = %d, want %d", step, length, len(model))
		}

		var values []int

		ropeEach(
			root,
			func(value int) bool {
				values = append(values, value)
				return true
			},
		)

		if !reflect.DeepEqual(values, model) {
			t.Fatalf("step %d: rope values differ", step)
		}

		var expected uint64
		for _, value := range model {
			expected += uint64(value)
		}

		if height != root.height || uintSum != expected || intSum != int64(expected) {
			t.Fatalf("step %d: invalid root aggregates", step)
		}
	}
}

func assertRopeNode[T any](
	t *testing.T,
	root *ropeNode[T],
	chunkSize int,
) (int, int, uint64, int64) {
	t.Helper()

	if root == nil {
		return 0, 0, 0, 0
	}

	if root.items != nil {
		if len(root.items) == 0 || len(root.items) > chunkSize {
			t.Fatalf("leaf size = %d, chunk size = %d", len(root.items), chunkSize)
		}

		if root.left != nil || root.right != nil || root.height != 1 {
			t.Fatal("leaf contains branch metadata")
		}

		return root.len, root.height, root.uint, root.sint
	}

	leftLen, leftHeight, leftUint, leftInt := assertRopeNode(t, root.left, chunkSize)

	rightLen, rightHeight, rightUint, rightInt := assertRopeNode(t, root.right, chunkSize)
	if difference := leftHeight - rightHeight; difference < -1 || difference > 1 {
		t.Fatalf("AVL height difference = %d", difference)
	}

	height := max(leftHeight, rightHeight) + 1
	if root.len != leftLen+rightLen ||
		root.height != height ||
		root.uint != leftUint+rightUint ||
		root.sint != leftInt+rightInt {
		t.Fatal("branch metadata is inconsistent")
	}

	return root.len, height, root.uint, root.sint
}

func testIntMetrics(value int) (uint64, int64) {
	return uint64(value), int64(value)
}

func testSpliceInts(model []int, index, deleteCount int, inserted []int) []int {
	next := make([]int, 0, len(model)-deleteCount+len(inserted))
	next = append(next, model[:index]...)
	next = append(next, inserted...)
	next = append(next, model[index+deleteCount:]...)

	return next
}
