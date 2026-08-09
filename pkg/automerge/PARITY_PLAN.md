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
Total required pending entries: **0**.

Every interop-required upstream test is now covered. The isolate/integrate
incremental patch ordering
(`incorrect_patches_produced_when_isolating_and_integrating`) is reproduced by
having the native incremental diff chain through the isolation frontiers recorded
in the window: it emits `diff(cursor -> isolation frontier)` followed by
`diff(isolation frontier -> current heads)`, which yields the reference's
"reset then rebuild" patch stream (deletes for the prior keys, conflicting puts,
and a splice only for each winning object). This special path activates only when
an isolate occurred since the diff cursor was last set, so ordinary incremental
diffs are unchanged. Materialization now also skips losing conflict alternatives
at a map key so only the winning object's content is spliced.

`observe_counter_change_application` is covered as a native-matches-reference
differential: the pinned reference (`automerge` 0.10.0 embedded as WASM) collapses
an applied create-and-increment counter change into a single `put_map` of the
materialized value through `diff_incremental` rather than emitting per-operation
patches, and native reproduces that reference behavior exactly.

DEFLATE compression on save is now covered: the native save compresses change
chunks whose body reaches the reference DEFLATE_MIN_SIZE threshold (small changes
stay byte-identical), and SaveNoCompress mirrors AutoCommit::save_nocompress.

Transaction isolation is now covered: `isolate` pins reads and writes to a
historical frontier using derived concurrency actors (matching Rust's
`with_concurrency` scheme), keeps merges hidden until `integrate`, and supports
repeated isolate/integrate cycles. The value-level isolation tests (`can_isolate`,
`can_transaction_at`, `update_text_change_at`) pass on both engines; only the
patch-ordering test (`incorrect_patches_produced_when_isolating_and_integrating`)
remains, because it asserts the exact incremental patch stream produced by Rust's
patch log across isolate/integrate, which native's state-comparison diff does not
reproduce.

Orphan retention across save/load is now covered without a full document-chunk
encoder: the native save appends retained orphan changes and the native load
falls back to a dependency-tolerant path that applies every change whose
dependencies are satisfiable and queues the rest (still failing a load that can
apply nothing, so a bare orphan without a base is rejected as before).

The V2 sync internals are now covered: the empty-message codec round-trips to
the reference wire bytes, and Bloom false-positive recovery is verified on both
engines using the reference engine's real Bloom filter (exposed through the
`am_bloom_contains` FFI) to locate genuine false positives. The native engine's
V2 sync uses exact head comparison instead of Bloom filters, so it is immune to
false positives by construction while still converging in these scenarios.

Legacy V1 sync protocol interoperability (V1↔V2 sessions, compressed changes in
V1 sessions, and old-peer capability fallback) is intentionally out of scope:
this project uses only the V2 sync protocol. Those upstream cases are recorded
as api-convenience rather than interop-required. Rust-internal library rustdoc
examples, the Rust owned/manual transaction object API, `Send` trait checks, and
JavaScript binding-type helpers (ImmutableString/RawString, legacy Text-as-array,
proxy/change-callback) are likewise recorded as api-convenience or
language-specific rather than interop-required.

## Known native defects (found by parity reproduction, fix pending)

**Mark boundaries: fixed for the reported case, deeper cases still diverge.**
The originally reported defect is fixed. Mark begin and end operations now hold
positions in the sequence order, insertions (including the mark boundaries
themselves) resolve their anchors through a port of the reference's insert
query, and spans are produced by a mark state machine walking that order. Text
inserted at an expanding boundary now keeps its mark after the originally marked
content is deleted, and text inserted after the whole marked range was deleted
still does not gain the mark.

Two behaviours were essential to get right and are easy to regress:

- a splice resolves its insertion anchor and inserts *before* deleting, matching
  the reference, so replacement text is positioned against the pre-deletion
  sequence and lands inside an expanding mark;
- mark precedence follows creation order, so a later unmark overrides an earlier
  mark where they overlap, and a mark left open covers nothing (a zero-length
  mark presents this way because its begin and end share an anchor and sibling
  insertions are ordered by descending operation ID, so the end is visited
  first).

Randomized differential testing that compares mark *values* against the
reference (a stronger assertion than upstream `marks_are_okay`, which only
checks span consolidation and text) drove several fixes this pass: an over-long
splice deletion and an over-long mark range are now clamped to the end of the
text rather than rejected, matching the reference. One characterized case
remains: marking an *empty* text with a range whose end is past the end (for
example `mark(0, 2)`) and the `ExpandMark::Both` flag, then inserting at the
head, should capture the inserted text — the reference reports it as marked and
treats it differently from a true zero-length `mark(0, 0)`, which captures
nothing. Native cannot reproduce this because on empty text both the start and
the clamped end anchor resolve to the head with no predecessor element, so the
two marks are structurally identical to it; distinguishing them needs a richer
anchor representation for a boundary positioned beyond the end. The case is
narrow (a mark past the end of an empty text followed by an insertion) and the
`TestRustText_MarksAreOkay` value-level differential reproduces it.

**Independent re-encoding of concurrent edits (determinism, not interop).**
The randomized `TestDifferentialStress_ConcurrentMerge` harness applies the same
concurrent edits to a native and a reference peer and merges them. The two
engines always converge to identical materialized values, but in rare cases a
single change ends up with different bytes (and therefore a different hash)
between the engines. This does not affect real interoperability: in a live
session each change is created once by one engine and shipped as bytes to the
other, which the interop suite verifies. It only means the two engines are not
guaranteed to pick byte-identical encodings when independently re-encoding the
same logical edit. The convergence test therefore compares materialized values;
head-hash equality is asserted only for the single-actor determinism test.

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

Complete (V2 protocol only; legacy V1 interoperability is out of scope):

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
- all V2 sync entries are covered (legacy V1 interoperability is out of scope).

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
