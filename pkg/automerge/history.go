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

package automerge

import (
	"fmt"
	"time"
)

// ChangeInfo is one Automerge change with authorship metadata, in causal order
// from ChangeHistory.
type ChangeInfo struct {
	Hash  Hash
	Actor []byte
	Time  time.Time
}

// ChangeHistory returns every change reachable from the document frontier in
// causal order. Time is the Unix timestamp stored on the change; a zero value
// means the change was committed without a timestamp.
func (d *Document) ChangeHistory() ([]ChangeInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	entries, err := d.engine.ChangeHistory()
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge change history: %w", err)
	}

	result := make([]ChangeInfo, len(entries))
	for i, entry := range entries {
		info := ChangeInfo{
			Hash:  Hash(entry.Hash),
			Actor: append([]byte(nil), entry.Actor...),
		}
		if entry.Time != 0 {
			info.Time = time.Unix(entry.Time, 0).UTC()
		}

		result[i] = info
	}

	return result, nil
}
