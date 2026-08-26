// Copyright (c) 2025 Probo Inc.
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
	"context"
	"strings"

	"github.com/rivo/uniseg"
)

// UpdateText replaces the text content of the object with value using a minimal
// grapheme-aware Myers diff, mirroring the Rust AutoCommit::update_text helper.
// It computes the shortest sequence of splice operations transforming the
// current text into value so concurrent edits to untouched regions merge
// cleanly. Splice positions are expressed in UTF-16 code units, matching the
// default text encoding used by the reference backend.
func (b *Engine) UpdateText(ctx context.Context, handle uint32, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := b.textObject(handle); err != nil {
		return err
	}

	current, err := b.Text(ctx, handle)
	if err != nil {
		return err
	}

	if current == value {
		return nil
	}

	oldGraphemes := graphemeClusters(current)
	newGraphemes := graphemeClusters(value)

	hook := &textDiffHook{ctx: ctx, engine: b, handle: handle, old: oldGraphemes, new: newGraphemes}
	myersDiff(hook, oldGraphemes, newGraphemes)

	return hook.err
}

// graphemeClusters splits a string into UAX #29 extended grapheme clusters so
// that diffing operates on user-perceived characters (e.g. emoji ZWJ sequences)
// rather than code points, matching Rust's unicode-segmentation based diff.
func graphemeClusters(text string) []string {
	if text == "" {
		return nil
	}

	clusters := make([]string, 0, len(text))
	state := -1

	for len(text) > 0 {
		var cluster string

		cluster, text, _, state = uniseg.FirstGraphemeClusterInString(text, state)
		clusters = append(clusters, cluster)
	}

	return clusters
}

// utf16Width returns the number of UTF-16 code units required to encode a string.
func utf16Width(text string) int {
	width := 0

	for _, r := range text {
		if r > 0xFFFF {
			width += 2
		} else {
			width++
		}
	}

	return width
}

// textDiffHook applies the edits produced by the Myers diff as splice
// operations on the target text object.
type textDiffHook struct {
	ctx    context.Context
	engine *Engine
	handle uint32
	old    []string
	new    []string
	idx    int
	err    error
}

func (h *textDiffHook) failed() bool {
	return h.err != nil
}

func (h *textDiffHook) equal(oldIndex, _ int, length int) {
	for i := range length {
		h.idx += utf16Width(h.old[oldIndex+i])
	}
}

func (h *textDiffHook) delete(oldIndex, oldLen, _ int) {
	if h.err != nil {
		return
	}

	deleted := 0

	for i := range oldLen {
		deleted += utf16Width(h.old[oldIndex+i])
	}

	if err := h.engine.SpliceText(h.ctx, h.handle, uint32(h.idx), int32(deleted), ""); err != nil {
		h.err = err
	}
}

func (h *textDiffHook) insert(_ int, newIndex, newLen int) {
	if h.err != nil {
		return
	}

	var builder strings.Builder

	for i := range newLen {
		builder.WriteString(h.new[newIndex+i])
	}

	chars := builder.String()
	if err := h.engine.SpliceText(h.ctx, h.handle, uint32(h.idx), 0, chars); err != nil {
		h.err = err
		return
	}

	h.idx += utf16Width(chars)
}

// diffSink receives the edit script produced by the Myers diff. Text and block
// reconciliation implement it over their respective element sequences.
type diffSink interface {
	equal(oldIndex, newIndex, length int)
	delete(oldIndex, oldLen, newIndex int)
	insert(oldIndex, newIndex, newLen int)
	failed() bool
}

// myersDiff computes the difference between old and new using Myers' O((N+M)D)
// algorithm and reports edits to the hook. It is a direct port of the reference
// Rust implementation (copied there from the similar crate) so the emitted
// edit script—and therefore the resulting change—matches byte for byte.
func myersDiff(hook diffSink, before, after []string) {
	maximum := maxD(len(before), len(after))
	vf := newVArray(maximum)
	vb := newVArray(maximum)
	conquer(hook, before, 0, len(before), after, 0, len(after), vf, vb)
}

type vArray struct {
	offset int
	values []int
}

func newVArray(maximum int) *vArray {
	return &vArray{offset: maximum, values: make([]int, 2*maximum)}
}

func (a *vArray) get(k int) int {
	return a.values[k+a.offset]
}

func (a *vArray) set(k, value int) {
	a.values[k+a.offset] = value
}

func maxD(oldLen, newLen int) int {
	return (oldLen+newLen+1)/2 + 1
}

func commonPrefixLen(before []string, oldStart, oldEnd int, after []string, newStart, newEnd int) int {
	if oldStart >= oldEnd || newStart >= newEnd {
		return 0
	}

	length := 0
	for oldStart+length < oldEnd && newStart+length < newEnd && before[oldStart+length] == after[newStart+length] {
		length++
	}

	return length
}

func commonSuffixLen(before []string, oldStart, oldEnd int, after []string, newStart, newEnd int) int {
	if oldStart >= oldEnd || newStart >= newEnd {
		return 0
	}

	length := 0
	for length < (oldEnd-oldStart) && length < (newEnd-newStart) && before[oldEnd-1-length] == after[newEnd-1-length] {
		length++
	}

	return length
}

func findMiddleSnake(
	before []string,
	oldStart int,
	oldEnd int,
	after []string,
	newStart int,
	newEnd int,
	vf *vArray,
	vb *vArray,
) (int, int, bool) {
	n := oldEnd - oldStart
	m := newEnd - newStart
	delta := n - m
	odd := delta&1 == 1

	vf.set(1, 0)
	vb.set(1, 0)

	dMax := maxD(n, m)

	for d := range dMax {
		for k := d; k >= -d; k -= 2 {
			var x int
			if k == -d || (k != d && vf.get(k-1) < vf.get(k+1)) {
				x = vf.get(k + 1)
			} else {
				x = vf.get(k-1) + 1
			}

			y := x - k

			x0, y0 := x, y
			if x < n && y < m {
				advance := commonPrefixLen(before, oldStart+x, oldEnd, after, newStart+y, newEnd)
				x += advance
			}

			vf.set(k, x)

			if odd && abs(k-delta) <= d-1 {
				if vf.get(k)+vb.get(-(k-delta)) >= n {
					return x0 + oldStart, y0 + newStart, true
				}
			}
		}

		for k := d; k >= -d; k -= 2 {
			var x int
			if k == -d || (k != d && vb.get(k-1) < vb.get(k+1)) {
				x = vb.get(k + 1)
			} else {
				x = vb.get(k-1) + 1
			}

			y := x - k

			if x < n && y < m {
				advance := commonSuffixLen(before, oldStart, oldStart+n-x, after, newStart, newStart+m-y)
				x += advance
				y += advance
			}

			vb.set(k, x)

			if !odd && abs(k-delta) <= d {
				if vb.get(k)+vf.get(-(k-delta)) >= n {
					return n - x + oldStart, m - y + newStart, true
				}
			}
		}
	}

	return 0, 0, false
}

func conquer(
	hook diffSink,
	before []string,
	oldStart int,
	oldEnd int,
	after []string,
	newStart int,
	newEnd int,
	vf *vArray,
	vb *vArray,
) {
	if hook.failed() {
		return
	}

	prefix := commonPrefixLen(before, oldStart, oldEnd, after, newStart, newEnd)
	if prefix > 0 {
		hook.equal(oldStart, newStart, prefix)
	}

	oldStart += prefix
	newStart += prefix

	suffix := commonSuffixLen(before, oldStart, oldEnd, after, newStart, newEnd)
	suffixOld := oldEnd - suffix
	suffixNew := newEnd - suffix
	oldEnd -= suffix
	newEnd -= suffix

	switch {
	case oldStart >= oldEnd && newStart >= newEnd:
		// Nothing to do.
	case newStart >= newEnd:
		hook.delete(oldStart, oldEnd-oldStart, newStart)
	case oldStart >= oldEnd:
		hook.insert(oldStart, newStart, newEnd-newStart)
	default:
		if xStart, yStart, ok := findMiddleSnake(before, oldStart, oldEnd, after, newStart, newEnd, vf, vb); ok {
			conquer(hook, before, oldStart, xStart, after, newStart, yStart, vf, vb)
			conquer(hook, before, xStart, oldEnd, after, yStart, newEnd, vf, vb)
		} else {
			hook.delete(oldStart, oldEnd-oldStart, newStart)
			hook.insert(oldStart, newStart, newEnd-newStart)
		}
	}

	if hook.failed() {
		return
	}

	if suffix > 0 {
		hook.equal(suffixOld, suffixNew, suffix)
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
