# Automerge table prototype

`@automerge/prosemirror` 0.2.0 cannot represent Tiptap tables losslessly in
its linear rich-text schema. A cell marker can name parent block types, but it
cannot identify parent instances. Cells in two adjacent rows therefore both
have `["table", "table-row"]` as parents, and the reader keeps one row open.
Adding an ID to the cell marker does not restore the missing row boundary.

This prototype keeps the linear `body` for ordinary rich text and adds a
`tiptap-table` embed carrying a stable table ID. The referenced state is:

```text
tables[tableID]
  rowIDs: CRDT sequence of stable IDs
  rows[rowID]
    cellIDs: CRDT sequence of stable IDs
    cells[cellID]
      body: Automerge rich text
```

Rows and cells are maps keyed by stable IDs, while their display order remains
a CRDT sequence. Every cell body is a dedicated Automerge rich-text object.
Consequently, edits to separate cells merge independently; no operation writes
the table or its JSON representation to one LWW register.

The hidden `automergeId` ProseMirror attribute carries identity through Tiptap
transactions, including row and column moves. New nodes initially fall back to
their predecessor at the same position and receive an ID when the transaction
is reconciled.

## Prototype boundary

The conversion and storage model are viable and covered by row-boundary,
round-trip, and concurrent-cell tests. The plugin currently rebuilds the
ProseMirror document for a remote table patch. Before production use, the fork
should translate nested table patches to narrow ProseMirror transactions so
selection mapping and large-table performance match the upstream text path.
Concurrent structural edits also need a product decision for dangling rows or
cells when one peer deletes an ancestor that another peer edits.
