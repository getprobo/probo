# Probo — TypeScript Frontend — @probo/prosemirror

Markdown ↔ ProseMirror node-tree bridge built on **`@tiptap/pm`** (the ProseMirror modules
re-exported by Tiptap). Used by the rich-text editor in `apps/console` for fields like finding
descriptions, control descriptions, and freeform notes.

The package is intentionally minimal: it owns the **schema** (which nodes/marks are allowed), the
**markdown serializer** (ProseMirror → Markdown for storage), and the **markdown parser**
(Markdown → ProseMirror for editor mount). The editor itself (Tiptap React component) lives in
the consuming app.

## Key files

- `packages/prosemirror/src/schema.ts` — node + mark schema (paragraph, heading, lists, code,
  bold, italic, link, …).
- `packages/prosemirror/src/serializer.ts` — ProseMirror node tree → Markdown string.
- `packages/prosemirror/src/parser.ts` — Markdown string → ProseMirror node tree.

## How to extend

To add a new node (e.g. callout, mention):

1. Add the schema entry in `schema.ts`.
2. Extend serializer + parser in lockstep — round-trip must be lossless.
3. On the consuming side, add the matching Tiptap extension that renders it.
4. Add unit tests for round-tripping a sample Markdown string (currently the package has none —
   see [testing.md](../testing.md#2-current-state--coverage-gaps)).

## Top pitfalls

1. **Lossy round-trip** — adding a node to the schema without matching serializer/parser code
   will silently drop content on save.
2. **No tests today** — be especially careful when refactoring the schema; a regression won't be
   caught by CI.
