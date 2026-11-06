# ProseMirror schemas and extensions

Rich text is a ProseMirror document JSON string. Behavior lives on a compiled
`Schema` built from per-node and per-mark extensions, not on package-level
switches. The console editor's full schema is `prosemirror.DocumentSchema`.
Narrower profiles (comments, for example) compose a subset of the same
extensions.

## Packages

| Layer | Package / files | Role |
| --- | --- | --- |
| Schema | `pkg/prosemirror` | Validate, sanitize, render HTML/Markdown, text length |
| Automerge bridge (Go) | `pkg/automerge/prosemirror` | Seed and materialize collaboration CRDTs (`ToSpans` / `Render`) |
| Editor (TS) | `packages/ui/src/RichEditor` | Tiptap extensions; collaborative mapping via `automerge()` |

`pkg/prosemirror` never imports Automerge. Bridge codecs live next to the CRDT
code so the schema package stays independent of the collaboration engine.

## Schema API

```go
schema := prosemirror.NewSchema(nodes, marks)

node, err := schema.Parse(raw)           // unmarshal + root check + validate
err = schema.Validate(node)
sanitized, err := schema.Sanitize(raw)   // one Check(sanitize) walk
html, err := schema.RenderHTML(node)
md, err := schema.RenderMarkdown(node)
n := schema.TextLength(node)
ok := schema.HasVisibleContent(node)
```

Callers pass a schema into the validator:

```go
v.Check(req.Content, "content", validator.ProseMirrorContent(prosemirror.DocumentSchema))
```

`prosemirror.Parse` still exists for read-only traversal with no validation.
Prefer `DocumentSchema.Parse` when the content must match the editor schema.

## Composing a schema

Each extension is a constructor named after its Tiptap counterpart:
`Doc()`, `Paragraph()`, `Text()`, `Heading()`, `Mention(MentionOptions)`,
`Bold()`, `Link()`, and so on. `NewSchema` compiles them once into dense lookup
tables and panics on misconfiguration (duplicate names, a content rule that
references a missing type, more than 64 node types, or a missing `Doc`/`Text`).

```go
var DocumentSchema = NewSchema(
	[]NodeSpec{
		Doc(), Paragraph(), Text(), Heading(), /* ... */
		Mention(MentionOptions{}), // empty EntityTypes: any GID
	},
	[]MarkSpec{Bold(), Italic(), Underline(), Strike(), Code(), Link()},
)
```

`MentionOptions{EntityTypes: []uint16{...}}` restricts mention IDs to those
entity types. An empty list accepts any GID (document behavior today).

### Append-only rule

Extensions are only ever **added** to a schema that already has stored content.
Removing an extension makes old documents fail to validate or render. If a type
must go away, run a data migration first, then shrink the schema.

When rendering stored content, use the schema it was written with, or
`DocumentSchema` (a superset).

## Adding an extension

Do all of the following in one change (or a short stack of commits). Coverage
tests fail if any step is skipped.

### 1. Schema (`pkg/prosemirror`)

1. Add `ext_<name>.go` implementing `NodeSpec` or `MarkSpec`.
2. Put validation and sanitize normalization in a single `Check(n, mode)` that
   decodes attrs once. In `CheckSanitize`, re-encode typed attrs so unknown keys
   are stripped (security boundary; keep byte-stable key order).
3. Implement `RenderHTML` / `RenderMarkdown` (marks: `Open*` / `Close*`).
4. Optionally implement `TextLength(n Node) int` and/or `Visible() bool`.
   `NewSchema` resolves those once; do not type-assert during a walk.
5. Register the constructor in `DocumentSchema` (and any narrower schema that
   should allow it).
6. Keep error messages byte-identical to existing ones when behavior is
   unchanged; clients surface `err.Error()`.

### 2. Automerge Go bridge (`pkg/automerge/prosemirror`)

1. Add `codec_<name>.go` implementing `nodeCodec` or `markCodec`.
2. Pick a `blockRole` (`roleTextblock`, `roleContainer`, `roleStructural`,
   `roleEmbed`, `roleInlineEmbed`, or `roleWrapper` for list shells).
3. Register the codec in the `nodeCodecs` / `markCodecs` slice in `codec.go`.
4. Add a row to `testdata/schema-mapping.json` (shared ledger with the frontend).
5. `TestCodecsCoverDocumentSchema` and `TestSchemaMappingLedger` must stay green.

Inline leaves (hard break, mention) use `roleInlineEmbed` with `isEmbed: true`.
They are emitted from inside a text block and collected back into the parent
block; their `Render` method is not used on the reverse path.

### 3. Frontend (`packages/ui/src/RichEditor`)

1. Declare `automerge: () => ({ ... })` on the Tiptap extension (or wrap a
   built-in in `automergeExtensions.ts`).
2. Include `AutomergeMappingExtension` in collaborative extension lists so the
   mapping is copied onto the ProseMirror spec.
3. Remove the node from any explicit opt-out set in `automergeCoverage.test.ts`.
4. Keep `schemaMappingParity.test.ts` green against `schema-mapping.json`.
5. After ledger changes that affect render fixtures, regenerate with
   `make generate-prosemirror-parity` when the corpus needs updating.

Built-in Tiptap nodes/marks that need a mapping are wrapped once in
`automergeExtensions.ts`. Custom nodes (`ImageExtension`, `FileExtension`,
`MentionExtension`, …) declare `automerge()` directly.

`createSchemaAdapter` is memoized with a `WeakMap` keyed by the extensions
array (and optional target schema). Do not rebuild the adapter on every keystroke
path without that cache.

## Performance rules

This package sits on hot paths (`Sanitize` on every collaborative apply, docgen
renders). Keep these invariants:

- **Compile once.** `NewSchema` builds index tables and bitset content rules.
  Groups are never resolved during a walk.
- **One walk, one attrs decode.** `Check(n, mode)` validates and (in sanitize
  mode) normalizes in a single pass.
- **No per-node dispatch allocations.** Specs are stateless values in the
  compiled table. Writers are concrete structs passed by pointer.
- **Benchmark gate.** `BenchmarkParse`, `BenchmarkValidate`, `BenchmarkSanitize`,
  `BenchmarkRenderHTML`, `BenchmarkRenderMarkdown`, and `BenchmarkTextLength`
  (plus Automerge `BenchmarkToSpans` / `BenchmarkRender`) must not regress more
  than ~5% ns/op or grow allocs/op versus the recorded baseline under
  `pkg/prosemirror/testdata/benchstat/`.

## Tests to keep green

- Restricted schemas reject excluded nodes/marks (`schema_test.go`).
- `FuzzSanitize`: sanitize is idempotent; sanitize output always validates.
- Automerge round-trips and ledger/coverage tests in `pkg/automerge/prosemirror`.
- Frontend parity: `schemaMappingParity.test.ts`, `prosemirrorRenderParity.test.ts`,
  `automergeCoverage.test.ts`, collaboration round-trips including mentions.

## Out of scope (follow-ups)

- Markdown/HTML **parsers** still emit full-schema nodes; they are not schema-
  filtered. Comments use plain text today. A future `Schema.Coerce` would turn
  unsupported blocks into paragraphs if comments accept Markdown.
- Integration traversals such as `ReferencedFiles` and Linear mention sync are
  not schema rules; they walk JSON for a specific purpose.
