# Automerge parity plan

## Goal

Make the native Go engine a complete behavioral replacement for the public
Rust `automerge` 0.10.0 engine while retaining a small JavaScript 3.4.0 boundary
suite for behavior introduced by the browser binding.

Completion means:

- every required entry in `testdata/upstream-parity.json` is mapped to an
  executable Go test;
- `make audit-automerge-interop` passes with zero pending required entries;
- documents, changes, heads, cursors, marks, blocks, and sync messages can move
  through Go, Rust, and JavaScript in any order without changing transferred
  identities or materialized state;
- race, malformed-input, fuzz, and conformance suites pass; and
- no production fallback silently bypasses the native engine.

This plan excludes Rust-private storage structures and JavaScript
proxy/packaging mechanics. They remain recorded in the manifest but do not
define Go engine behavior.

## Pinned sources

| Surface | Version | Commit |
|---|---|---|
| Rust `automerge` | 0.10.0 with UTF-16 indexing | `a4f584c86358dd07f83f36708573e1c8d1bd8161` |
| JavaScript `@automerge/automerge` | 3.4.0 | `f8b0911dc9d86265dd62934b7dc782571e3a7fcb` |

The generator verifies both Git revisions before updating the ledger.

## Current baseline

| Classification | Covered | Pending |
|---|---:|---:|
| Required Rust + JavaScript boundary behavior | 206 | 156 |
| Non-blocking JavaScript convenience behavior | 44 | 196 |
| Private or language-specific behavior | 110 | — |

The generated manifest is the authoritative leaf-level list. Counts in this
document are informational and must be updated whenever the manifest changes.

## Required source backlog

| Source | Pending | Missing behavior |
|---|---:|---|
| `rust/tests/test.rs` | 8 | patch-log misuse, transactions and isolation, and change-encoding round-trips |
| `rust/src/sync.rs` | 5 | Bloom false positives, reset/data-loss recovery, old-peer fallback, and message encode/decode internals |
| `rust/tests/text.rs` | 4 | block-adjacent marks, isolation patches, property scenarios, and zero-length spans |
| `rust/tests/batch_insert.rs` | 2 | invalid-scalar rejection and transaction integration |
| `rust/src/automerge/current_state.rs` | 1 | loading current-state patches from a stored fixture |
| `rust/tests/test_save_load_orphans.rs` | 2 | preserving orphan changes across a document save and discarding them on request (needs a native document-chunk encoder) |
| `rust/src/sync/v1_compat_test` | 3 | V1→V2, V2→V1, and compressed-change compatibility |
| Rust `AutoCommit` doctests | 3 | commit options, diff, and incremental diff |
| Curated JavaScript scalar boundaries | 3 | immutable strings, raw-string compatibility, and numeric wrappers |
| Curated JavaScript mark boundaries | 2 | patch-visible marks and expansion at text end |
| Rust iterator behavior | 1 | document iteration with conflicts |
| Remaining Rust public/doctest cases | 6 | hydration, autoserde, manual transaction, patch log, document parse, and one active automerge regression |

Total required pending entries: **44**.

## Known native defects (found by parity reproduction, fix pending)

No reproduced native-only defects are currently known. New differential
failures must be recorded here until fixed.

## Optional JavaScript convenience backlog

These do not block Go/Rust engine parity but remain tracked:

| Source | Pending |
|---|---:|
| Legacy JavaScript proxy and mutation API | 105 |
| JavaScript patch application/callback helpers | 31 |
| JavaScript basic convenience APIs | 25 |
| JavaScript sync wrapper duplicates | 22 |
| Fragments and `changeAt` wrappers | 8 |
| Unstable change API | 3 |
| JavaScript conflicted-proxy mutation | 3 |
| Anonymization helper | 1 |

## Workstream 1: storage and change identity

Implement and verify:

- complete document, change, compressed-change, and bundle parsing;
- canonical encoding for every operation and scalar column;
- expanded/compressed change byte and hash stability;
- 64-bit object IDs and actor tables referenced only by deletes;
- partial and incremental loading with corrupted tails;
- orphan preservation/discard rules;
- missing dependency behavior;
- unknown-column and unknown-scalar preservation;
- no-op and empty changes across save/load;
- official malformed fixtures and fuzz crashers; and
- V1 storage compatibility.

Acceptance:

- every storage fixture loads or rejects identically to Rust;
- Go-authored changes retain their hash after Rust and JavaScript relay;
- Rust-authored snapshots can be extended and forwarded by Go; and
- all storage and parser manifest entries are covered.

## Workstream 2: maps, lists, scalars, counters, and conflicts

Complete:

- conflicts involving different scalar/object types;
- nested map/list conflicts;
- updates inside conflicted objects;
- concurrent assignment/deletion;
- updates to concurrently deleted objects;
- counter increments attached only to the values they precede;
- list-counter deletion semantics;
- actor/counter sequence ordering;
- insertions around large deleted runs;
- causality-preserving insertion chunks;
- large-list indexing and regression fixtures;
- wrong-object and invalid-index errors; and
- merge after no-op/equal-value updates.

Acceptance:

- deterministic and randomized Go/Rust histories match values, conflicts,
  heads, and transferred hashes;
- all `rust/tests/test.rs` core-model entries are covered; and
- three-peer forwarding preserves every conflict.

## Workstream 3: text encodings and cursors

Complete:

- UTF-16 length/get/put/insert/delete behavior;
- update-text minimal diff behavior;
- grapheme and combining-character cases;
- cursors at historical heads;
- start/end and movement bias through nested deletions;
- cursor reuse across compatible documents;
- text inside lists/maps;
- string-to-text migration; and
- cursor patch source metadata at the JavaScript boundary.

Acceptance:

- every valid JavaScript UTF-16 position resolves identically in Go and Rust;
- every official cursor byte sequence round-trips;
- cursor-addressed edits converge after arbitrary concurrent changes; and
- all text-encoding and curated JavaScript cursor entries are covered.

## Workstream 4: marks and blocks

Complete:

- all mark expansion modes at both boundaries;
- adjacent, nested, overlapping, empty, and zero-length marks;
- marks with changing names, values, and scalar types;
- marks on emoji, combining characters, whitespace, and deleted text;
- marks crossing block markers;
- block split/join/replace and block attribute updates;
- simultaneous text, mark, and block updates;
- historical marks and blocks;
- `updateSpans` semantic equivalent;
- list/table/divider/code block marker values; and
- mark/block patches and merge diffs.

Acceptance:

- `spans` output matches Rust and curated JavaScript for every fixture;
- Go-authored marks/blocks load and remain editable everywhere;
- Rust/JS-authored rich text can be edited by Go without normalization loss; and
- all mark, block, and rich-text manifest entries are covered.

## Workstream 5: synchronization

Complete:

- V1/V2 cross-version sessions;
- compressed changes in legacy sessions;
- Bloom filter false positives and chains;
- explicit requested changes and nonexistent requests;
- branching/merging histories;
- simultaneous messages and edits while in flight;
- persisted sessions and process/data-loss recovery;
- all read-only/reset transitions;
- publishers with multiple consumers;
- fully connected and relay topologies; and
- stale shared heads and duplicate paths.

Acceptance:

- every official sync test quiesces within a fixed bound;
- all peer combinations (Go/Go, Go/Rust, Rust/Go) converge;
- duplicate, delayed, reordered, and replayed messages are safe; and
- all sync and V1 compatibility entries are covered.

## Workstream 6: transactions, history, patches, and current state

Complete:

- owned/manual transactions;
- pending reads and writes;
- commit options and empty transactions;
- rollback with multiple actors;
- isolation and transactions at heads;
- historical map/list/text/mark reads;
- reverse diffs after object/block deletion;
- map/list/text/increment/mark/block patches;
- large list patches;
- patch-log ownership errors without panics;
- current-state materialization with conflicts; and
- incremental diff and patch callback semantics.

Go APIs can be idiomatic; they must expose the same observable capability.

Acceptance:

- transaction and historical outcomes match Rust;
- patch sequences are semantically equivalent and use UTF-16 indexes;
- applying patches reconstructs the expected hydrated state; and
- all transaction/current-state/patch entries are covered.

## Workstream 7: batch and hydration completion

Complete the remaining batch behaviors:

- patch generation;
- merge after batch insertion;
- scalar-target rejection;
- transaction integration;
- repeated batches;
- replacement of existing nested maps; and
- parity between batch and individual operation output.

Acceptance:

- batch output loads and merges identically in Go and Rust;
- rollback leaves no batch-created objects;
- nested values and text preserve object identity; and
- all active batch entries are covered.

## Workstream 8: curated JavaScript boundary

Retain only:

- UTF-16 and cursor semantics;
- immutable strings, `Date`, bytes, `BigInt`, integer, unsigned, and float
  conversion;
- marks, blocks, tables, and `updateSpans`;
- default/no/provided timestamps;
- JavaScript-generated documents/changes loading in Go;
- Go changes relayed through JavaScript without hash changes; and
- browser sync message compatibility.

Do not recreate JavaScript proxy syntax, packaging, export aliases, or callback
ergonomics in Go.

## Workstream 9: state-machine differential and fuzzing

Expand the neutral scenario runner to support:

- fork, merge, apply changes, and sync schedules;
- every scalar/object operation;
- marks, blocks, cursors, and historical reads;
- transactions and rollback;
- malformed and partial bytes; and
- deterministic network faults.

Run every scenario independently through native Go and native Rust, then route
each serialized result through JavaScript boundary checks where relevant.

Fuzz:

- documents, changes, compressed chunks, sync messages, cursors, and patches;
- stateful map/list/text/mark/block histories;
- cross-engine generated bytes;
- dependency reordering and duplicate delivery; and
- load/save/merge cycles under memory and size limits.

Every discovered failure becomes a deterministic regression before its fix.

## Workstream 10: performance and production readiness

Benchmark identical native Go/Rust workloads for:

- long-lived map/list histories;
- text typing and large pastes;
- marks, blocks, and tables;
- save/load with compacted and uncompacted histories;
- two- and three-peer sync;
- merge-heavy branching histories; and
- snapshot compaction.

Performance does not waive correctness. Optimization changes must pass the
entire parity and fuzz suite.

Before enabling the native engine without fallback:

1. `make audit-automerge-interop` passes.
2. Go race tests pass.
3. Rust/JavaScript conformance passes.
4. Fuzz smoke and retained corpus pass.
5. Benchmarks show no unbounded regression.
6. The exported Go surface contains only supported, tested capabilities.

## Implementation rules

- Do not mark an upstream test covered because a nearby test looks similar.
- Every mapping names the exact local test and explains the equivalent
  assertion.
- Reproduce a failing upstream behavior before modifying the engine.
- Keep Rust/WASM and JavaScript oracle code independent from native Go logic.
- Preserve original bytes for transferred immutable changes.
- Never silently normalize unknown protocol data.
- Keep all bounds explicit for untrusted documents and sync messages.
- Do not claim completion while any required manifest entry remains pending.
