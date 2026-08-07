# Automerge

This package is Probo's owned, no-CGO boundary for Automerge documents.

## Trust model

The first backend is a small first-party WASI adapter around the official
[`automerge`](https://crates.io/crates/automerge) Rust crate. The adapter:

- pins `automerge` 0.10.0 and every transitive crate in `Cargo.lock`;
- compiles with UTF-16 indexing to match the JavaScript editor;
- runs in-process through wazero without CGO or native shared libraries; and
- gives every open document an isolated WASM instance.

The Go API depends on a private backend interface. A native Go implementation
can be added behind that interface and compared with the embedded reference
engine before becoming the default.

The committed `reference.wasm` is reproducible from reviewed Rust source:

```sh
rustup toolchain install 1.89.0 --profile minimal --target wasm32-wasip1
make generate-automerge-reference
```

## Compatibility checks

Ordinary Go tests cover binary round trips, UTF-16 text offsets, concurrent
changes, merge convergence, and lifecycle behavior:

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
