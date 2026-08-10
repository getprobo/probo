# Native Automerge engine

This package is the pure-Go Automerge engine. It is intentionally internal:
the stable API lives in `pkg/automerge`, and callers must not depend on engine
types or storage details.

## File boundaries

| Area | Files | Responsibility |
|---|---|---|
| Engine lifecycle | `engine.go` | Engine construction, load/save, incremental persistence and actor setup |
| Public object operations | `object.go` | Map/list scalar and object operations, counters, deletion, keys and length |
| Rich text API | `rich_text.go`, `text_diff.go` | Text splice, blocks, spans, marks, cursors and update-spans reconciliation |
| Transactions | `transaction.go` | Commit, rollback, isolate and integrate |
| Patches | `patch.go` | Current state, historical/current diffs and incremental patch generation |
| History and merge | `history.go` | Heads, dependency queries, changes, apply and merge, including orphan queues |
| Sync engine | `sync_engine.go` | Per-peer V2 synchronization state machine |
| Engine helpers | `engine_helpers.go` | Handle validation, sequence index resolution, scalar wire values and cursors |
| Materialized state | `state.go` | Change graph, pending/applied changes, map indexes and historical state |
| Sequence state | `sequence_state.go` | RGA order, sequence caches, conflicts and visible winners |
| Rich-text state | `rich_text_state.go` | Span/mark state machines, mark anchors and UTF-16 ranges |
| Hydration | `hydrate_state.go` | Recursive map/list materialization |
| Storage facade | `storage.go` | Delegation to the independent storage and encoding packages |
| Shared model facade | `types.go` | Aliases to the independent shared types package |

## Internal package boundaries

| Package | Responsibility |
|---|---|
| `internal/types` | Dependency-free actor, operation, change, object, scalar and chunk model |
| `internal/encoding` | Bounded binary reader, ULEB128 and length-prefixed primitives |
| `internal/storage` | Automerge chunk/column encoding, decoding and graph validation |
| `internal/sync` | V1/V2 sync message wire codec and resource limits |
| `internal/native` | Mutable CRDT engine, materialized state, rich text, patches and sync orchestration |
| `internal/reference` | Rust/WASM differential oracle used only by parity tests |

## Dependency direction

The engine follows this direction:

```text
pkg/automerge public API
        ↓
native Engine methods
        ↓
State / sequence / rich-text state
        ↓
internal/types ← internal/encoding ← internal/storage
        ↑
 internal/sync
```

Sync, patches and rich text may use common engine/state helpers, but storage
code must not depend on those higher-level features. Keep protocol state out of
the CRDT state graph, and keep materialization/diff concerns out of the storage
codec.

## Refactoring rule

Changes that move code between these files must be behavior-preserving and run
the strict gates:

```sh
make audit-automerge-interop
make test-automerge-conformance
go test -race ./pkg/automerge/...
```

Byte identity is part of behavior: concurrent native/reference edits must
produce identical changes and heads, not merely converge to equal values.
