# Vendored mcpgen

This is a temporary fork of [getprobo/mcpgen](https://github.com/getprobo/mcpgen)
(module `go.probo.inc/mcpgen`) with tool annotation improvements:

- `title` field on tools (emitted as `Tool.Title` and `ToolAnnotations.Title`)
- When `hints` are present, always emit annotations so write tools get
  `readOnlyHint: false` and `destructiveHint: false`, distinguishing them from
  deletes (`destructiveHint: true`)

`go.mod` replaces `go.probo.inc/mcpgen` with this directory. Once the same
changes land upstream, drop the replace and delete this tree.
