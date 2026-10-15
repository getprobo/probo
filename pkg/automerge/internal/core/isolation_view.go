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
// immutable canonical change graph.
func stateFromSharedColumns(columns *columnarState) (*State, error) {
	return stateFromCanonicalColumns(columns)
}

// newIsolationView filters the canonical columns to the ancestry of heads and
// builds query indexes over that pinned columnar view. The full columns remain
// untouched and continue to own changes outside the selected frontier.
func newIsolationView(
	columns *columnarState,
	heads []opset.ChangeHash,
) (*State, *columnarState, bool) {
	if columns == nil || columns.snapshot == nil || columns.columnsDirty {
		return nil, nil, false
	}

	visited := make(map[opset.ChangeHash]struct{})

	var visit func(opset.ChangeHash) bool

	visit = func(hash opset.ChangeHash) bool {
		if _, ok := visited[hash]; ok {
			return true
		}

		change, ok := columns.change(hash)
		if !ok {
			return false
		}

		for _, dependency := range change.Dependencies {
			if !visit(dependency) {
				return false
			}
		}

		visited[hash] = struct{}{}

		return true
	}

	for _, head := range heads {
		if !visit(head) {
			return nil, nil, false
		}
	}

	operationIDs := make(map[opset.OpID]struct{})
	changes := make([]opset.Change, 0, len(visited))
	for i := range columns.changes {
		change := &columns.changes[i]
		if change.Hash == nil {
			return nil, nil, false
		}

		if _, selected := visited[*change.Hash]; !selected {
			continue
		}

		cloned := cloneChange(*change)
		for j := range cloned.Operations {
			operationIDs[cloned.Operations[j].ID] = struct{}{}
		}

		changes = append(changes, cloned)
	}

	for i := range changes {
		for j := range changes[i].Operations {
			operation := &changes[i].Operations[j]
			successors := operation.Successors[:0]
			for _, successor := range operation.Successors {
				if _, selected := operationIDs[successor]; selected {
					successors = append(successors, successor)
				}
			}

			operation.Successors = successors
		}
	}

	order := make([]opset.OpID, 0, len(operationIDs))
	for _, operation := range columns.operations {
		if _, selected := operationIDs[operation.ID]; selected {
			order = append(order, operation.ID)
		}
	}

	pinnedColumns, err := newColumnarState(
		&opset.Document{
			Actors:         append([]opset.ActorID(nil), columns.actors...),
			Changes:        changes,
			Heads:          append([]opset.ChangeHash(nil), heads...),
			OperationOrder: order,
		},
	)
	if err != nil {
		return nil, nil, false
	}

	view, err := stateFromCanonicalColumns(pinnedColumns)
	if err != nil {
		return nil, nil, false
	}

	return view, pinnedColumns, true
}
