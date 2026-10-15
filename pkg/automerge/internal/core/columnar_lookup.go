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
	if ok && row >= 0 && row < len(c.operations) {
		operation := c.operations[row]
		if operation.ID == identifier {
			return operation, true
		}
	}

	if operation, ok := c.nonCanonicalOperations[identifier]; ok {
		return operation, true
	}

	return opset.Operation{}, false
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

// attachCanonical publishes columns as the complete committed authority. Any
// staged change and operation rows have just been incorporated, so only a later
// open transaction may repopulate the overlays.
func (s *State) attachCanonical(columns *columnarState) {
	if columns == nil || columns.snapshot == nil || columns.columnsDirty {
		panic("cannot attach incomplete canonical columns")
	}

	s.columns = columns
	clear(s.operations)
	clear(s.changes)
	clear(s.heads)
	clear(s.removedHeads)
}

func (s *State) operation(identifier opset.OpID) (opset.Operation, bool) {
	if operation, ok := s.operations[identifier]; ok {
		return operation, true
	}

	if operation, ok := s.columns.operation(identifier); ok {
		return operation, true
	}

	return opset.Operation{}, false
}

func (s *State) hasOperationID(identifier opset.OpID) bool {
	_, ok := s.operationIDs[identifier]
	return ok
}

func (s *State) eachOperation(yield func(opset.Operation) bool) {
	for _, operation := range s.operations {
		if !yield(operation) {
			return
		}
	}

	if s.columns != nil {
		for _, operation := range s.columns.operations {
			if _, overlaid := s.operations[operation.ID]; overlaid {
				continue
			}

			if !yield(operation) {
				return
			}
		}

		for identifier, operation := range s.columns.nonCanonicalOperations {
			if _, overlaid := s.operations[identifier]; overlaid {
				continue
			}

			if !yield(operation) {
				return
			}
		}
	}
}

func (s *State) change(hash opset.ChangeHash) (*opset.Change, bool) {
	if change, ok := s.changes[hash]; ok {
		return change, true
	}

	if change, ok := s.columns.change(hash); ok {
		return change, true
	}

	return nil, false
}

func (s *State) eachChange(yield func(opset.ChangeHash, *opset.Change) bool) {
	for hash, change := range s.changes {
		if !yield(hash, change) {
			return
		}
	}

	if s.columns != nil {
		for i := range s.columns.changes {
			change := &s.columns.changes[i]
			if change.Hash == nil {
				continue
			}

			if _, overlaid := s.changes[*change.Hash]; overlaid {
				continue
			}

			if !yield(*change.Hash, change) {
				return
			}
		}
	}
}

func (s *State) operationCount() int {
	return len(s.operationIDs)
}

func (s *State) changeCount() int {
	count := len(s.changes)
	for hash := range s.changes {
		if _, canonical := s.columns.change(hash); canonical {
			count--
		}
	}

	if s.columns != nil {
		count += len(s.columns.changes)
	}

	return count
}

func (b *Engine) lookupOperation(
	identifier opset.OpID,
) (opset.Operation, bool) {
	return b.state.operation(identifier)
}

func (b *Engine) lookupChange(
	hash opset.ChangeHash,
) (*opset.Change, bool) {
	return b.state.change(hash)
}

func (b *Engine) currentHeads() []opset.ChangeHash {
	return b.state.Heads()
}
