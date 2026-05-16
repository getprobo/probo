# Probo — Go Backend — pkg/cli + pkg/cmd

**Purpose.** The `prb` CLI command tree. `pkg/cli` provides the GraphQL
HTTP client, generic cursor-pagination helper, OAuth2 / device-flow
auth, and YAML config (`~/.config/prb/config.yaml`). `pkg/cmd` is the
cobra command tree — one directory per resource, one file per verb.

> See [patterns.md § 9 CLI command shape](../patterns.md#9-cli-command-shape).

**Key files.**

- `pkg/cli/api/client.go` — GraphQL client (Bearer auth, 401
  auto-refresh, multipart uploads).
- `pkg/cli/api/pagination.go` — `api.Paginate[T]` generic cursor walker.
- `pkg/cli/config/config.go` — YAML config (hosts → tokens), env-var
  overrides (`PROBO_HOST`, `PROBO_TOKEN`).
- `pkg/cmd/cmdutil/factory.go` — `Factory` (DI container: IOStreams,
  Version, lazy Config).
- `pkg/cmd/cmdutil/table.go` — `NewTable` (lipgloss/table for list
  output).
- `pkg/cmd/iostreams/iostreams.go` — TTY/CI/NO_COLOR detection.
- `pkg/cmd/<resource>/<resource>.go` — group command wiring its verbs.
- `pkg/cmd/<resource>/<verb>/<verb>.go` — leaf command (one per file).
- `pkg/cmd/risk/list/list.go` — canonical leaf command.
- `pkg/cmd/root/root.go` — root command, registers all groups.

**How to extend (a new verb).**

1. Create `pkg/cmd/<resource>/<verb>/<verb>.go` with:
   - `const <verb>Query/Mutation = "..."` GraphQL string.
   - Unexported `<verb>Response` struct.
   - `func NewCmd<Verb>(f *cmdutil.Factory) *cobra.Command`.
2. Wire it from `pkg/cmd/<resource>/<resource>.go`.
3. Output: tables to `f.IOStreams.Out`, info/truncation to
   `f.IOStreams.ErrOut`.
4. Interactive prompts via `huh`, gated on `IOStreams.IsInteractive()`.

**Top pitfalls.**

- Calling `huh` without checking `IsInteractive()` — breaks CI / pipes /
  `--no-interactive`.
- Hand-rolling pagination — use `api.Paginate[T]` so cursor handling
  stays consistent across verbs.
- Bypassing the `Factory` and reading config directly — breaks tests
  that inject a fake `Config` via `Factory`.
- Coupling the leaf to a specific output format — keep table vs JSON
  selection in the leaf, controlled by the global `--output` flag from
  `cmdutil`.
