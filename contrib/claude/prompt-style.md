# Agent prompt style

Agent prompt templates (the `//go:embed`-ed `*.txt.tmpl` files that back an
agent's instructions) follow a consistent XML-tag structure. This keeps prompts
scannable, makes the contract between the prompt and the typed result struct
explicit, and matches what the models in use respond to best.

## Structure

Three top-level sections, always in this order:

```
<role>
One short paragraph: who the agent is and its single objective.
</role>

<task>
What the agent is given and the fields it must return. Describe the structured
output as a bullet list, one bullet per field, mirroring the typed result.
</task>

<instructions>
1. A numbered list of directives, ordered most-decisive first.
2. Each rule is a single, actionable directive.
3. ...
</instructions>
```

## Rules

- Use `<role>`, `<task>`, `<instructions>` in that order. Do not invent other
  top-level sections (no loose `Method:` / `Rules:` prose blocks).
- `<role>` is one short paragraph stating who the agent is and its objective.
- `<task>` enumerates the inputs and the output fields. List one bullet per
  field so the prompt stays in sync with the typed result struct.
- `<instructions>` is an ordered (numbered) list. Put the most important rule
  first. Keep each item to a single directive.
- Never hardcode enum values or source-of-truth lists. Use a `{{.Placeholder}}`
  and substitute at runtime. See [`file-naming.md`](file-naming.md) and the
  `template-files` cursor rule.
- No commentary outside the tags; the whole body is the system instruction.

## Per-row input

The embedded template is the static system instruction. The dynamic per-row
input (the specific entity being processed) is built separately in Go and sent
as the user message, wrapped in its own descriptive tags:

```go
fmt.Fprintf(&b, "<name> %s </name>\n", party.Name)
fmt.Fprintf(&b, "<website> %s </website>\n", website)
```

## Examples

- `pkg/thirdparty/prompts/disambiguation.txt.tmpl`
- `pkg/cookiebanner/prompts/tracker_identification.txt.tmpl`
- `pkg/cookiebanner/prompts/tracker_enrichment.txt.tmpl`
- `pkg/thirdparty/prompts/common_third_party_company_profile.txt.tmpl`
- `pkg/thirdparty/prompts/common_third_party_compliance_docs.txt.tmpl`
