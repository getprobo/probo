# Automerge

This package is Probo's owned, no-CGO boundary for Automerge documents.

## Engines

The default backend is a clean-room, pure-Go Automerge 0.10 engine. It:

- decodes document, change, compressed-change, and v1/v2 sync formats;
- validates checksums, actor ownership, causal frontiers, sequences, and limits;
- preserves unknown columns and scalars for forward compatibility;
- supports maps, text, UTF-16 splices, rich-text materialization, cursors,
  changes, heads, merges, and synchronization; and
- emits changes accepted by official Rust and JavaScript implementations.

The package also retains a first-party WASI adapter around the official
[`automerge`](https://crates.io/crates/automerge) Rust crate as an independent
differential oracle. The adapter:

- pins `automerge` 0.10.0 and every transitive crate in `Cargo.lock`;
- compiles with UTF-16 indexing to match the JavaScript editor;
- runs in-process through wazero without CGO or native shared libraries; and
- gives every open document an isolated WASM instance.

Use `NewReference` and `LoadReference` only in conformance tests or when
diagnosing native parity. Production `New` and `Load` use the Go engine.

The committed `reference.wasm` is reproducible from reviewed Rust source:

```sh
rustup toolchain install 1.89.0 --profile minimal --target wasm32-wasip1
make generate-automerge-reference
```

## Compatibility checks

Ordinary Go tests cover binary round trips, UTF-16 text offsets, concurrent
changes, randomized text histories, merge convergence, native/reference sync,
rich-text spans, cursor movement, and lifecycle behavior:

```sh
go test -race ./pkg/automerge/...
```

The cross-language suite additionally loads Go documents in the official
JavaScript implementation and JavaScript documents in Go:

```sh
npm install
make test-automerge-conformance
```

The conformance oracle is deliberately separate from the Go implementation so
the two paths do not share adapter code.
