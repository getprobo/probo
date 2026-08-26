# Automerge

This package is Probo's owned, no-CGO boundary for Automerge documents.

## Implementation

The public package uses a clean-room, pure-Go Automerge 0.10 implementation. It:

- decodes document, change, compressed-change, and v1/v2 sync formats;
- validates checksums, actor ownership, causal frontiers, sequences, and limits;
- preserves unknown columns and scalars for forward compatibility;
- supports maps, text, UTF-16 splices, rich-text materialization, cursors,
  changes, heads, merges, and synchronization; and
- emits changes accepted by official Rust and JavaScript implementations.

An internal test-support package retains a first-party WASI adapter around the
official [`automerge`](https://crates.io/crates/automerge) Rust crate as an
independent differential oracle. It is not part of the public API or production
dependency graph. The adapter:

- pins `automerge` 0.11.0 and every transitive crate in `Cargo.lock`;
- compiles with UTF-16 indexing to match the JavaScript editor;
- runs in-process through wazero without CGO or native shared libraries; and
- gives every open document an isolated WASM instance.

Parity tests access the oracle through `internal/testsupport`; production
`New` and `Load` can only use the Go implementation.

The committed `reference.wasm` is reproducible from reviewed Rust source:

```sh
rustup toolchain install 1.90.0 --profile minimal --target wasm32-wasip1
make generate-automerge-reference
```

## Compatibility checks

Ordinary Go tests cover binary round trips, UTF-16 text offsets, concurrent
changes, randomized concurrent text histories, dependency reordering, merge
forwarding, duplicate messages, persisted sync sessions, three-peer relays,
native/reference sync, rich-text spans, cursor movement, and lifecycle
behavior:

```sh
go test -race ./pkg/automerge/...
```

The cross-language suite additionally loads Go documents in the official
JavaScript implementation, JavaScript documents in Go, and verifies that
JavaScript preserves the exact hashes and bytes of Go-generated changes:

```sh
npm install
make test-automerge-conformance
```

The conformance oracle is deliberately separate from the Go implementation so
the two paths do not share adapter code.

Neutral interoperability scenarios live in `testdata/scenarios`. The same JSON
operations are executed independently by Go, Rust/WASM, and
JavaScript. Independently authored changes may have different hashes when an API
chooses a different valid operation order; the gate instead requires every
engine to load, preserve, extend, and semantically materialize every other
engine's output while preserving each transferred change's original hash.

Fuzz targets exercise document and change decoding, sync message parsing and
round trips, and rich-text projection. Every production failure should be
reduced to a deterministic regression seed before its fix is merged.

```sh
make fuzz-automerge AUTOMERGE_FUZZ_TIME=30s
```

Go and Rust/WASM benchmarks cover warm document creation, map mutation,
character-by-character text editing, 10,000-character save/load, and initial
native/reference synchronization:

```sh
make benchmark-automerge
```

For a direct optimized native-Go versus native-Rust comparison, use the shared
worker harness. It executes identical actors, operations, commit metadata,
warmups, and sample counts, and rejects results unless both engines produce the
same document checksum:

```sh
make benchmark-automerge-native
```

## Upstream parity ledger

[`testdata/upstream-parity.json`](testdata/upstream-parity.json) is the
authoritative parity ledger. It inventories every active Rust test reported by
the pinned test harness, every JavaScript leaf test, and every JavaScript
packaging scenario at the pinned upstream revisions:

- Rust `automerge` 0.11.0, tag `rust/automerge-0.11.0`, commit
  `8feb8be8be203e3b878bf2cb5919601094f3c816`; and
- JavaScript `@automerge/automerge` 3.4.1, tag `js/automerge-3.4.1`, commit
  `8d7b12f8da553afbb325e37a6c66942b8dd4d994`.

The ledger also records all 16 upstream JavaScript packaging scenarios as
language-specific coverage so they are never confused with Go CRDT semantics.
The Rust denominator is the 456 active harness tests plus 16 doctests reported
by `cargo test -p automerge --features utf16-indexing -- --list`; dormant and
feature-disabled source code is not counted as parity debt.

Each entry must be mapped to one or more executable local tests or classified
as language-specific with a concrete rationale. Pending entries are visible
debt, not implicit coverage. Wire/state interoperability is gated separately
from language-level convenience APIs:

```sh
make audit-automerge-interop
```

The broader API-parity ledger remains available when working on historical
views, patch callbacks, transaction wrappers, and similar conveniences:

```sh
make audit-automerge-parity
```

Regenerate the inventory from clean checkouts of those exact commits:

```sh
node packages/automerge-conformance/generate-parity-inventory.mjs \
  --rust-root /path/to/rust/automerge \
  --rust-test-list /path/to/cargo-test-list.txt \
  --javascript-root /path/to/javascript/test \
  --mappings packages/automerge-conformance/parity-mappings.json \
  > pkg/automerge/testdata/upstream-parity.json
```
