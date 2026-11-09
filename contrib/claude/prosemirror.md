# ProseMirror schemas and extensions

Rich text is a ProseMirror document JSON string. Validation lives on a compiled
`Schema` built from per-node and per-mark specs. Rendering and markdown import
are separate packages that register per-type renderers and rewrite rules — the
same model as Automerge codecs. The console editor's schema is
`richtext.DocumentSchema`. Narrower profiles (comments, for example) compose a
subset of the same extensions.

## Packages

| Layer | Package / files | Role |
| --- | --- | --- |
| Schema | `pkg/prosemirror` | Node tree, specs (`Name`/`Group`/`Content`/`Check`), validate, sanitize, text length |
| HTML output | `pkg/prosemirror/html` | `NodeRenderer` / `MarkRenderer`, `Writer`, `NewRenderer` |
| Markdown | `pkg/prosemirror/markdown` | Import (`Parse`, `Rule`, `ApplyRules`) and export (`NewRenderer`) |
| Attachments | `pkg/richtext/attachment` | Image/File specs, HTML/MD renderers, import rule, `Path` / `ParseID` / `Referenced` / `FileIDs` |
| Mentions | `pkg/richtext/mention` | `Document` / `Profile` specs, HTML/MD renderers, `ProfileIDs` / `NewProfileIDs` |
| Product wiring | `pkg/richtext` | `DocumentSchema`, `CommentSchema`, `DocumentHTML`, `DocumentMarkdown`, `ParseMarkdown` |
| Automerge bridge (Go) | `pkg/automerge/prosemirror` | Seed and materialize collaboration CRDTs (`ToSpans` / `Render`) |
| Editor (TS) | `packages/ui/src/RichEditor` | Tiptap extensions; collaborative mapping via `automerge()` |

`pkg/prosemirror` does not import `pkg/gid`, `pkg/richtext`, or Automerge.
Probo-specific nodes live in their own `pkg/richtext/<name>` package. Bridge
codecs live next to the CRDT code so the schema package stays independent of
the collaboration engine.

## Schema and render API

```go
schema := prosemirror.NewSchema(nodes, marks)

node, err := schema.Parse(raw)           // unmarshal + root type "doc"; does not validate
err = schema.Validate(node)
sanitized, err := schema.Sanitize(raw)   // one Check(sanitize) walk
n := schema.TextLength(node)
ok := schema.HasVisibleContent(node)

htmlOut, err := richtext.DocumentHTML.Render(node)
mdOut, err := richtext.DocumentMarkdown.Render(node)
```

Callers pass a schema into the validator:

```go
v.Check(req.Content, "content", validator.ProseMirrorContent(richtext.DocumentSchema))
```

`prosemirror.Parse` unmarshals JSON and does not validate. `Schema.Parse` only
checks that the root type is `doc`. Use `Validate` or `ValidateJSON` when the
tree must match the schema. `richtext.ParseMarkdown` imports markdown, applies
`attachment.ImportRule()`, and validates against `DocumentSchema`.

## Composing a schema

Generic extensions are constructors in `pkg/prosemirror` named after their
Tiptap counterparts: `Doc()`, `Paragraph()`, `Text()`, `Heading()`, `Bold()`,
`Link()`, and so on. Probo-specific nodes come from their packages:
`attachment.Image()`, `attachment.File()`, `mention.Document()`,
`mention.Profile()`.

`NewSchema` compiles specs once into dense lookup tables and panics on
misconfiguration (duplicate names, a content rule that references a missing
type, more than 64 node types, or a missing `Doc`/`Text`).

```go
var schema = prosemirror.NewSchema(
	[]prosemirror.NodeSpec{
		prosemirror.Doc(), prosemirror.Paragraph(), prosemirror.Text(),
		attachment.Image(), attachment.File(),
		mention.Document(),
	},
	[]prosemirror.MarkSpec{prosemirror.Bold(), prosemirror.Link()},
)

htmlRenderer := html.NewRenderer(
	schema,
	append(html.Nodes(), attachment.HTMLNodes()...),
	html.Marks(),
)
mdRenderer := markdown.NewRenderer(
	schema,
	htmlRenderer,
	append(markdown.Nodes(), attachment.MarkdownNodes()...),
	markdown.Marks(),
)
```

`mention.Document()` accepts any GID. `mention.Profile()` accepts only
membership-profile GIDs (`richtext.CommentSchema`).

`NewRenderer` panics if a schema node or mark has no renderer, or if a type is
registered twice. It compiles a dense table so a walk performs no type
assertions.

### Append-only rule

Extensions are only ever **added** to a schema that already has stored content.
Removing an extension makes old documents fail to validate or render. If a type
must go away, run a data migration first, then shrink the schema.

When rendering stored content, use the schema it was written with, or
`DocumentSchema` (a superset).

## Adding an extension

Do all of the following in one change (or a short stack of commits). Coverage
tests fail if any step is skipped.

### 1. Spec

- **Generic node/mark:** add `ext_<name>.go` in `pkg/prosemirror` implementing
  `NodeSpec` or `MarkSpec` (`Name`, `Group`, `Content`, `Check` only).
- **Probo-specific node:** add a dedicated package under `pkg/richtext/<name>`
  with the spec, HTML/Markdown renderers, optional import `Rule`, and helpers.

Put validation and sanitize normalization in a single `Check(n, mode)` that
decodes attrs once. In `CheckSanitize`, re-encode typed attrs so unknown keys
are stripped (security boundary; keep byte-stable key order). Optionally
implement `TextLength(n Node) int` and/or `Visible() bool`. `NewSchema`
resolves those once; do not type-assert during a walk.

Register the constructor on `richtext.DocumentSchema` (and any narrower schema
that should allow it). Keep error messages byte-identical to existing ones when
behavior is unchanged; clients surface `err.Error()`.

### 2. HTML and Markdown renderers

1. Add a `NodeRenderer` / `MarkRenderer` in `pkg/prosemirror/html` (or the
   product package for Probo-specific nodes) and register it via `Nodes()` /
   `Marks()` / `HTMLNodes()`.
2. Do the same for `pkg/prosemirror/markdown`.
3. Optionally add a `markdown.Rule` for import-time rewrites (for example
   `attachment.ImportRule()` turns `/api/files/v1/<gid>` links into file
   nodes). Wire it in `richtext.ParseMarkdown` through `markdown.ApplyRules`.

### 3. Automerge Go bridge (`pkg/automerge/prosemirror`)

1. Add `codec_<name>.go` implementing `nodeCodec` or `markCodec`.
2. Pick a `blockRole` (`roleTextblock`, `roleContainer`, `roleStructural`,
   `roleEmbed`, `roleInlineEmbed`, or `roleWrapper` for list shells).
3. Register the codec in the `nodeCodecs` / `markCodecs` slice in `codec.go`.
4. Add a row to `testdata/schema-mapping.json` (shared ledger with the frontend).
5. `TestCodecsCoverDocumentSchema` and `TestSchemaMappingLedger` must stay green.

Inline leaves (hard break, mention) use `roleInlineEmbed` with `isEmbed: true`.
They are emitted from inside a text block and collected back into the parent
block; their `Render` method is not used on the reverse path.

### 4. Frontend (`packages/ui/src/RichEditor`)

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

- **Compile once.** `NewSchema` and `NewRenderer` build index tables and bitset
  content rules. Groups and renderer lookups are never resolved during a walk.
- **One walk, one attrs decode.** `Check(n, mode)` validates and (in sanitize
  mode) normalizes in a single pass.
- **No per-node dispatch allocations.** Specs and renderers are stateless values
  in the compiled table. Writers are concrete structs passed by pointer.
- **Benchmark gate.** `BenchmarkParse`, `BenchmarkValidate`, `BenchmarkSanitize`,
  `BenchmarkRenderHTML`, `BenchmarkRenderMarkdown`, and `BenchmarkTextLength`
  (plus Automerge `BenchmarkToSpans` / `BenchmarkRender`) must not regress more
  than ~5% ns/op or grow allocs/op versus the recorded baseline under
  `pkg/prosemirror/testdata/benchstat/`.

## Tests to keep green

- Restricted schemas reject excluded nodes/marks (`schema_test.go`).
- `FuzzParse` round-trips JSON that parses. `FuzzSanitize`: sanitize is
  idempotent and its output always validates. `richtext.FuzzParseMarkdown`:
  markdown that imports also validates against `DocumentSchema`.
- Automerge round-trips and ledger/coverage tests in `pkg/automerge/prosemirror`.
- Frontend parity: `schemaMappingParity.test.ts`, `prosemirrorRenderParity.test.ts`,
  `automergeCoverage.test.ts`, collaboration round-trips including mentions.

## Out of scope (follow-ups)

- Markdown/HTML **parsers** still emit full-schema nodes; they are not schema-
  filtered. A future `Schema.Coerce` would turn unsupported blocks into
  paragraphs if a narrower schema accepts Markdown.
- Integration traversals such as `attachment.Referenced` and Linear mention sync
  are not schema rules; they walk JSON for a specific purpose.
