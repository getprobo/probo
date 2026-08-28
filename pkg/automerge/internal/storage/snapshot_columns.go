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

import (
	"fmt"
	"slices"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

type (
	// DecodedDocument carries the canonical snapshot rows and Hexane roots
	// produced by Decode. The semantic Document fields remain available for
	// compatibility paths, but the direct runtime adopts this value without
	// rebuilding or cloning the snapshot.
	DecodedDocument struct {
		Snapshot   *SnapshotColumns
		Operations []opset.Operation
	}

	// SnapshotSplice describes one atomic update to both snapshot tables.
	// Actors, Heads, and HeadIndexes are the complete post-edit metadata.
	// DependencyIndexes resolves dependencies of inserted change rows.
	SnapshotSplice struct {
		Actors               []opset.ActorID
		Heads                []opset.ChangeHash
		HeadIndexes          []uint64
		DependencyIndexes    map[opset.ChangeHash]uint64
		ChangeIndex          int
		ChangeDeleteCount    int
		Changes              []*opset.Change
		OperationIndex       int
		OperationDeleteCount int
		Operations           []opset.Operation
		OperationSplices     []SnapshotOperationSplice
	}

	SnapshotOperationSplice struct {
		Index       int
		DeleteCount int
		Operations  []opset.Operation
	}

	// SnapshotColumns is a mutable, export-ready columnar document. Clone shares
	// immutable Hexane rope roots, and subsequent splices copy only edited paths.
	SnapshotColumns struct {
		actors            []opset.ActorID
		heads             []opset.ChangeHash
		headIndexes       []uint64
		changeColumns     *hexaneChangeColumns
		operationColumns  *hexaneOperationColumns
		encodedChanges    []encodedColumn
		encodedOperations []encodedColumn
		changeCount       int
		operationCount    int
	}
)

func newDecodedSnapshotColumns(
	actors []opset.ActorID,
	heads []opset.ChangeHash,
	headIndexes []uint64,
	changeColumns map[uint32]column,
	operationColumns map[uint32]column,
	changeCount int,
	operationCount int,
) (*SnapshotColumns, error) {
	encodedChanges := encodedColumnsFromDecoded(changeColumns, changeColumnSpecifications)
	encodedOperations := encodedColumnsFromDecoded(
		operationColumns,
		operationColumnSpecifications,
	)
	return &SnapshotColumns{
		actors:            actors,
		heads:             heads,
		headIndexes:       headIndexes,
		encodedChanges:    encodedChanges,
		encodedOperations: encodedOperations,
		changeCount:       changeCount,
		operationCount:    operationCount,
	}, nil
}

func encodedColumnsFromDecoded(
	columns map[uint32]column,
	specifications []uint32,
) []encodedColumn {
	encoded := make([]encodedColumn, 0, len(specifications))
	for _, specification := range specifications {
		decoded, ok := columns[specification]
		if !ok {
			continue
		}
		encoded = append(encoded, encodedColumn{
			specification: specification,
			data:          decoded.data,
		})
	}
	return encoded
}

// NewSnapshotColumns builds mutable export-ready columns from semantic rows.
func NewSnapshotColumns(
	document *opset.Document,
	operations []opset.Operation,
) (*SnapshotColumns, error) {
	if decoded, ok := document.Canonical.(*DecodedDocument); ok &&
		decoded.Snapshot != nil {
		return decoded.Snapshot.Clone(), nil
	}

	changes := make([]*opset.Change, len(document.Changes))
	for i := range document.Changes {
		changes[i] = &document.Changes[i]
	}

	actors := documentActorTable(changes, operations)
	actorIndexes := make(map[opset.ActorID]uint64, len(actors))
	for i, actor := range actors {
		actorIndexes[actor] = uint64(i)
	}

	heads, headIndexes, err := documentHeads(changes)
	if err != nil {
		return nil, err
	}

	changeColumns, err := newHexaneChangeColumns(changes, actorIndexes)
	if err != nil {
		return nil, err
	}
	operationColumns, err := newHexaneOperationColumns(operations, actorIndexes)
	if err != nil {
		return nil, err
	}
	encodedChanges := rawEncodedColumns(document.ChangeColumns)
	if len(encodedChanges) == 0 {
		recordFullColumnEncoding()
		encodedChanges, err = changeColumns.Encoded()
		if err != nil {
			return nil, err
		}
	}
	encodedOperations := rawEncodedColumns(document.OperationColumns)
	if len(encodedOperations) == 0 {
		recordFullColumnEncoding()
		encodedOperations, err = operationColumns.Encoded()
		if err != nil {
			return nil, err
		}
	}

	return &SnapshotColumns{
		actors:            append([]opset.ActorID(nil), actors...),
		heads:             append([]opset.ChangeHash(nil), heads...),
		headIndexes:       append([]uint64(nil), headIndexes...),
		changeColumns:     changeColumns,
		operationColumns:  operationColumns,
		encodedChanges:    cloneEncodedColumns(encodedChanges),
		encodedOperations: cloneEncodedColumns(encodedOperations),
		changeCount:       len(document.Changes),
		operationCount:    len(operations),
	}, nil
}

// Clone creates a constant-size copy whose Hexane column roots are shared until
// either copy is edited.
func (c *SnapshotColumns) Clone() *SnapshotColumns {
	if c == nil {
		return nil
	}

	cloned := &SnapshotColumns{
		actors:            c.actors,
		heads:             c.heads,
		headIndexes:       c.headIndexes,
		encodedChanges:    c.encodedChanges,
		encodedOperations: c.encodedOperations,
		changeCount:       c.changeCount,
		operationCount:    c.operationCount,
	}
	if c.changeColumns != nil {
		cloned.changeColumns = c.changeColumns.Clone()
	}
	if c.operationColumns != nil {
		cloned.operationColumns = c.operationColumns.Clone()
	}
	return cloned
}

// PrepareMutation materializes the persistent Hexane roots retained by a
// decoded snapshot. Loaded documents can defer this work until an object is
// opened for editing, while subsequent mutations only copy touched paths.
func (c *SnapshotColumns) PrepareMutation() error {
	if c == nil {
		return fmt.Errorf("snapshot columns are nil")
	}
	deferred := c.changeColumns == nil || c.operationColumns == nil
	if err := c.ensureRoots(); err != nil {
		return err
	}
	if !deferred {
		return nil
	}
	if _, err := c.changeColumns.Encoded(); err != nil {
		return fmt.Errorf("cannot prepare change encodings: %w", err)
	}
	if _, err := c.operationColumns.Encoded(); err != nil {
		return fmt.Errorf("cannot prepare operation encodings: %w", err)
	}
	return nil
}

// Replace transactionally rebuilds both tables from complete semantic rows.
func (c *SnapshotColumns) Replace(
	document *opset.Document,
	operations []opset.Operation,
) error {
	replacement, err := NewSnapshotColumns(document, operations)
	if err != nil {
		return fmt.Errorf("cannot build replacement snapshot columns: %w", err)
	}

	*c = *replacement

	return nil
}

// Splice transactionally edits row ranges and their grouped and raw subranges.
// Change-table edits that shift existing indexes must use Replace; appends and
// same-width replacements remain localized.
func (c *SnapshotColumns) Splice(edit SnapshotSplice) error {
	if c == nil {
		return fmt.Errorf("snapshot columns are nil")
	}
	if err := c.ensureRoots(); err != nil {
		return err
	}
	if edit.ChangeDeleteCount != len(edit.Changes) &&
		edit.ChangeIndex+edit.ChangeDeleteCount != c.changeColumns.actors.Len() {
		return fmt.Errorf("change splice shifts existing indexes; use Replace")
	}

	actorIndexes := make(map[opset.ActorID]uint64, len(edit.Actors))
	for i, actor := range edit.Actors {
		actorIndexes[actor] = uint64(i)
	}

	changes := c.changeColumns.Clone()
	operations := c.operationColumns.Clone()
	actorsAppended := len(edit.Actors) >= len(c.actors) &&
		slices.Equal(c.actors, edit.Actors[:len(c.actors)])
	actorsRemapped := !slices.Equal(c.actors, edit.Actors) && !actorsAppended
	if actorsRemapped {
		if err := changes.RemapActors(c.actors, actorIndexes); err != nil {
			return fmt.Errorf("cannot remap change columns: %w", err)
		}
		if err := operations.RemapActors(c.actors, actorIndexes); err != nil {
			return fmt.Errorf("cannot remap operation columns: %w", err)
		}
	}

	insertedChanges, err := newHexaneChangeColumnsForSplice(
		edit.Changes,
		actorIndexes,
		edit.DependencyIndexes,
	)
	if err != nil {
		return fmt.Errorf("cannot build inserted change columns: %w", err)
	}
	if _, err := insertedChanges.Encoded(); err != nil {
		return fmt.Errorf("cannot validate inserted change columns: %w", err)
	}
	changesEdited := edit.ChangeDeleteCount > 0 || len(edit.Changes) > 0
	if changesEdited {
		if err := changes.Splice(
			edit.ChangeIndex,
			edit.ChangeDeleteCount,
			insertedChanges,
		); err != nil {
			return err
		}
	}
	operationSplices := edit.OperationSplices
	if len(operationSplices) == 0 {
		operationSplices = []SnapshotOperationSplice{{
			Index:       edit.OperationIndex,
			DeleteCount: edit.OperationDeleteCount,
			Operations:  edit.Operations,
		}}
	}
	preparedOperationSplices := make(
		[]hexaneOperationSplice,
		len(operationSplices),
	)
	for i, operationSplice := range operationSplices {
		insertedOperations, err := newHexaneOperationColumns(
			operationSplice.Operations,
			actorIndexes,
		)
		if err != nil {
			return fmt.Errorf("cannot build inserted operation columns: %w", err)
		}
		preparedOperationSplices[i] = hexaneOperationSplice{
			index:       operationSplice.Index,
			deleteCount: operationSplice.DeleteCount,
			inserted:    insertedOperations,
		}
	}
	if err := operations.BatchSplice(preparedOperationSplices); err != nil {
		return fmt.Errorf("cannot splice operation column batch: %w", err)
	}
	encodedChanges := c.encodedChanges
	if actorsRemapped || changesEdited {
		encodedChanges = nil
	}
	operationsEdited := actorsRemapped
	for _, splice := range operationSplices {
		operationsEdited = operationsEdited ||
			splice.DeleteCount > 0 ||
			len(splice.Operations) > 0
	}
	encodedOperations := c.encodedOperations
	if operationsEdited {
		encodedOperations = nil
	}

	c.actors = append([]opset.ActorID(nil), edit.Actors...)
	c.heads = append([]opset.ChangeHash(nil), edit.Heads...)
	c.headIndexes = append([]uint64(nil), edit.HeadIndexes...)
	c.changeColumns = changes
	c.operationColumns = operations
	c.encodedChanges = encodedChanges
	c.encodedOperations = encodedOperations
	c.changeCount = c.changeColumns.actors.Len()
	c.operationCount = c.operationColumns.idActors.Len()

	return nil
}

func (c *SnapshotColumns) ensureRoots() error {
	if c.changeColumns == nil {
		columns, err := decodeHexaneChangeColumns(c.encodedChanges, c.changeCount)
		if err != nil {
			return fmt.Errorf("cannot decode retained change columns: %w", err)
		}
		c.changeColumns = columns
	}
	if c.operationColumns == nil {
		columns, err := decodeHexaneOperationColumns(
			c.encodedOperations,
			c.operationCount,
		)
		if err != nil {
			return fmt.Errorf("cannot decode retained operation columns: %w", err)
		}
		c.operationColumns = columns
	}
	return nil
}

// Encode assembles a compact document from maintained columns.
func (c *SnapshotColumns) Encode(
	unknown []opset.RawColumn,
	compress bool,
) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("snapshot columns are nil")
	}

	if c.encodedChanges == nil {
		encoded, err := c.changeColumns.Encoded()
		if err != nil {
			return nil, fmt.Errorf("cannot encode changed snapshot changes: %w", err)
		}
		c.encodedChanges = encoded
	}
	if c.encodedOperations == nil {
		encoded, err := c.operationColumns.Encoded()
		if err != nil {
			return nil, fmt.Errorf("cannot encode changed snapshot operations: %w", err)
		}
		c.encodedOperations = encoded
	}

	document := &opset.Document{UnknownColumns: unknown}
	changeColumns := compressColumns(
		sortColumns(append(
			append([]encodedColumn(nil), c.encodedChanges...),
			retainedColumns(document, changeColumnSpecifications)...,
		)),
		compress,
	)
	operationColumns := compressColumns(
		sortColumns(append(
			append([]encodedColumn(nil), c.encodedOperations...),
			retainedColumns(document, operationColumnSpecifications)...,
		)),
		compress,
	)

	body := make(
		[]byte,
		0,
		snapshotBodyCapacity(
			c.actors,
			c.heads,
			c.headIndexes,
			changeColumns,
			operationColumns,
		),
	)
	body = appendULEB(body, uint64(len(c.actors)))
	for _, actor := range c.actors {
		body = appendLengthPrefixedNative(body, actor.Bytes())
	}
	body = appendULEB(body, uint64(len(c.heads)))
	for _, head := range c.heads {
		body = append(body, head[:]...)
	}
	body = appendColumnMetadata(body, changeColumns)
	body = appendColumnMetadata(body, operationColumns)
	body = appendColumnData(body, changeColumns)
	body = appendColumnData(body, operationColumns)
	for _, index := range c.headIndexes {
		body = appendULEB(body, index)
	}

	return assembleChunk(opset.ChunkDocument, body), nil
}

func snapshotBodyCapacity(
	actors []opset.ActorID,
	heads []opset.ChangeHash,
	headIndexes []uint64,
	changeColumns []encodedColumn,
	operationColumns []encodedColumn,
) int {
	size := ulebSize(uint64(len(actors)))
	for _, actor := range actors {
		size += ulebSize(uint64(len(actor.Bytes()))) + len(actor.Bytes())
	}
	size += ulebSize(uint64(len(heads))) + len(heads)*len(opset.ChangeHash{})
	size += encodedColumnsSize(changeColumns)
	size += encodedColumnsSize(operationColumns)
	for _, index := range headIndexes {
		size += ulebSize(index)
	}

	return size
}

func encodedColumnsSize(columns []encodedColumn) int {
	size := ulebSize(uint64(len(columns)))
	for _, column := range columns {
		size += ulebSize(uint64(column.specification))
		size += ulebSize(uint64(len(column.data)))
		size += len(column.data)
	}

	return size
}

func ulebSize(value uint64) int {
	size := 1
	for value >= 0x80 {
		value >>= 7
		size++
	}

	return size
}

func rawEncodedColumns(columns []opset.RawColumn) []encodedColumn {
	encoded := make([]encodedColumn, len(columns))
	for i, column := range columns {
		encoded[i] = encodedColumn{
			specification: column.Specification &^ 8,
			data:          append([]byte(nil), column.Data...),
		}
	}

	return encoded
}

func cloneEncodedColumns(columns []encodedColumn) []encodedColumn {
	cloned := make([]encodedColumn, len(columns))
	for i, column := range columns {
		cloned[i] = encodedColumn{
			specification: column.specification,
			data:          append([]byte(nil), column.data...),
		}
	}

	return cloned
}
