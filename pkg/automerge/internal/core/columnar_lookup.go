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

func (i *operationRowIndex) lookup(identifier opset.OpID) (int, bool) {
	if i == nil {
		return 0, false
	}
	if row, ok := i.rows[identifier]; ok {
		return row, true
	}
	row, ok := i.parent.lookup(identifier)
	if !ok || !i.hasSplice {
		return row, ok
	}
	if row < i.spliceStart {
		return row, true
	}
	if row < i.spliceEnd {
		return 0, false
	}
	return row + i.spliceDelta, true
}

func (c *columnarState) operation(
	identifier opset.OpID,
) (opset.Operation, bool) {
	if c == nil || c.operationRows == nil {
		return opset.Operation{}, false
	}
	row, ok := c.operationRows.lookup(identifier)
	if !ok || row < 0 || row >= len(c.operations) {
		for _, operation := range c.operations {
			if operation.ID == identifier {
				return operation, true
			}
		}
		return opset.Operation{}, false
	}
	operation := c.operations[row]
	if operation.ID != identifier {
		for _, candidate := range c.operations {
			if candidate.ID == identifier {
				return candidate, true
			}
		}
		return opset.Operation{}, false
	}
	return operation, true
}

func (c *columnarState) change(
	hash opset.ChangeHash,
) (*opset.Change, bool) {
	if c == nil {
		return nil, false
	}
	row, ok := c.changeRows[hash]
	if !ok || row < 0 || row >= len(c.changes) {
		return nil, false
	}
	return &c.changes[row], true
}

func (c *columnarState) currentHeads() []opset.ChangeHash {
	if c == nil {
		return nil
	}
	return append([]opset.ChangeHash(nil), c.heads...)
}

// attachCanonical makes columns the committed authority and reduces the
// operation/change maps to rows not represented by the canonical snapshot
// (principally pending operations, deletes, and retained orphan data).
func (s *State) attachCanonical(columns *columnarState) {
	if columns == nil ||
		columns.snapshot == nil ||
		columns.columnsDirty ||
		s.columns == columns {
		return
	}

	retainedOperationIDs := make(map[opset.OpID]struct{})
	for hash, change := range s.changes {
		if _, canonical := columns.change(hash); canonical {
			continue
		}
		for _, operation := range change.Operations {
			retainedOperationIDs[operation.ID] = struct{}{}
		}
	}
	operations := make(map[opset.OpID]opset.Operation, len(retainedOperationIDs))
	for identifier := range retainedOperationIDs {
		if operation, ok := s.operations[identifier]; ok {
			operations[identifier] = operation
		}
	}
	changes := make(map[opset.ChangeHash]*opset.Change)
	for hash, change := range s.changes {
		if _, canonical := columns.change(hash); canonical {
			metadata := *change
			metadata.Operations = nil
			metadata.Raw = nil
			changes[hash] = &metadata
		} else {
			changes[hash] = change
		}
	}

	s.columns = columns
	s.operations = operations
	s.changes = changes
	clear(s.heads)
	clear(s.removedHeads)
}

func (s *State) operation(identifier opset.OpID) (opset.Operation, bool) {
	if operation, ok := s.columns.operation(identifier); ok {
		return operation, true
	}
	if operation, ok := s.operations[identifier]; ok {
		return operation, true
	}
	return opset.Operation{}, false
}

func (s *State) hasOperationID(identifier opset.OpID) bool {
	_, ok := s.operationIDs[identifier]
	return ok
}

func (s *State) eachOperation(yield func(opset.Operation) bool) {
	if s.columns != nil {
		for _, operation := range s.columns.operations {
			if !yield(operation) {
				return
			}
		}
	}
	for _, operation := range s.operations {
		if _, canonical := s.columns.operation(operation.ID); canonical {
			continue
		}
		if !yield(operation) {
			return
		}
	}
}

func (s *State) change(hash opset.ChangeHash) (*opset.Change, bool) {
	if change, ok := s.columns.change(hash); ok {
		if _, retained := s.changes[hash]; !retained {
			return nil, false
		}
		return change, true
	}
	if change, ok := s.changes[hash]; ok {
		return change, true
	}
	return nil, false
}

func (s *State) eachChange(yield func(opset.ChangeHash, *opset.Change) bool) {
	if s.columns != nil {
		for i := range s.columns.changes {
			change := &s.columns.changes[i]
			if change.Hash == nil {
				continue
			}
			if _, retained := s.changes[*change.Hash]; !retained {
				continue
			}
			if !yield(*change.Hash, change) {
				return
			}
		}
	}
	for hash, change := range s.changes {
		if _, canonical := s.columns.change(hash); canonical {
			continue
		}
		if !yield(hash, change) {
			return
		}
	}
}

func (s *State) operationCount() int {
	return len(s.operationIDs)
}

func (s *State) changeCount() int {
	count := 0
	if s.columns != nil {
		for i := range s.columns.changes {
			hash := s.columns.changes[i].Hash
			if hash != nil {
				if _, retained := s.changes[*hash]; retained {
					count++
				}
			}
		}
	}
	for hash := range s.changes {
		if _, canonical := s.columns.change(hash); !canonical {
			count++
		}
	}
	return count
}

func (b *Engine) bindColumnarState() {
	if b == nil ||
		b.isolationActive ||
		b.columns == nil ||
		b.columns.snapshot == nil ||
		b.columns.columnsDirty {
		return
	}
	b.state.attachCanonical(b.columns)
}

// lookupOperation overlays uncommitted rows on the canonical columnar OpSet.
// Historical isolation retains the row-state lookup until frontier-filtered
// columnar views replace fullState.
func (b *Engine) lookupOperation(
	identifier opset.OpID,
) (opset.Operation, bool) {
	b.bindColumnarState()
	if operation, ok := b.state.operations[identifier]; ok {
		return operation, true
	}
	if b.isolationActive {
		operation, ok := b.state.operation(identifier)
		return operation, ok
	}
	if operation, ok := b.columns.operation(identifier); ok {
		return operation, true
	}
	// Partial/orphan streams can retain semantic rows that have no canonical
	// operation row. Keep them readable until the columnar load view represents
	// retained orphan operations directly.
	operation, ok := b.state.operation(identifier)
	return operation, ok
}

func (b *Engine) lookupChange(
	hash opset.ChangeHash,
) (*opset.Change, bool) {
	b.bindColumnarState()
	if b.isolationActive {
		change, ok := b.state.change(hash)
		return change, ok
	}
	if change, ok := b.columns.change(hash); ok {
		return change, true
	}
	change, ok := b.state.change(hash)
	return change, ok
}

func (b *Engine) currentHeads() []opset.ChangeHash {
	b.bindColumnarState()
	if b.isolationActive {
		return b.state.Heads()
	}
	return b.columns.currentHeads()
}
