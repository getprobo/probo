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

package core

import "go.probo.inc/probo/pkg/automerge/internal/opset"

// stateFromSharedColumns builds fork-local query indexes without cloning the
// immutable canonical change graph. attachCanonical then retains only rows not
// represented by those columns.
func stateFromSharedColumns(columns *columnarState) (*State, error) {
	return stateFromCanonicalColumns(columns)
}

func stateChangesFromColumns(columns *columnarState) []opset.Change {
	changes := cloneChanges(columns.changes)
	for i := range changes {
		for j := range changes[i].Operations {
			operation := &changes[i].Operations[j]

			canonical, ok := columns.operation(operation.ID)
			if !ok {
				continue
			}

			predecessors := operation.Predecessors
			*operation = cloneOperation(canonical)
			operation.Predecessors = predecessors
		}
	}

	return changes
}

// newIsolationView materializes only the ancestry of heads. Unlike State.at it
// does not replay every operation through ApplyChange; it filters the immutable
// history once and constructs just the query indexes needed by isolated reads
// and writes.
func newIsolationView(
	full *State,
	columns *columnarState,
	heads []opset.ChangeHash,
) (*State, bool) {
	ordered := make([]opset.Change, 0)
	visited := make(map[opset.ChangeHash]struct{})

	var visit func(opset.ChangeHash) bool

	visit = func(hash opset.ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := full.change(hash)
		if !ok {
			return false
		}

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				return false
			}
		}

		cloned := *change

		cloned.Operations = make([]opset.Operation, len(change.Operations))
		for i, operation := range change.Operations {
			cloned.Operations[i] = cloneOperation(operation)
			// Snapshot changes can retain successors outside the selected
			// frontier. Predecessors are sufficient to rebuild supersession and
			// prevent future operations from leaking into the historical view.
			cloned.Operations[i].Successors = nil
		}

		ordered = append(ordered, cloned)
		visited[hash] = struct{}{}

		return true
	}

	for _, head := range heads {
		if !visit(head) {
			return nil, false
		}
	}

	view, err := NewStateFromDocument(
		&opset.Document{
			Changes: ordered,
			Heads:   append([]opset.ChangeHash(nil), heads...),
		},
	)
	if err != nil {
		return nil, false
	}

	return view, true
}
