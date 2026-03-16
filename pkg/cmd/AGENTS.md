# AGENTS.md — prb CLI

## Overview

`prb` is the Probo CLI built with [cobra](https://github.com/spf13/cobra). Entry point: `cmd/prb/main.go`.

## Package layout

| Package | Purpose |
|---|---|
| `cmd/prb` | Binary entry point — creates `Factory`, root command, and runs it |
| `pkg/cmd/root` | Root command — registers all top-level subcommands |
| `pkg/cmd/<resource>` | Command group (e.g. `risk`, `framework`, `webhook`) — wires subcommands |
| `pkg/cmd/<resource>/<verb>` | Leaf command (e.g. `risk/create`, `risk/list`) — owns the `RunE` |
| `pkg/cmd/cmdutil` | Shared helpers: `Factory`, flag validators, table/JSON output, time formatting |
| `pkg/cmd/iostreams` | Terminal I/O abstraction (stdout, stderr, color, interactivity) |
| `pkg/cli/api` | GraphQL client (`Client`) and generic pagination (`Paginate[T]`) |
| `pkg/cli/config` | Config file management (hosts, tokens, default org) |

## Adding a new resource command

1. Create `pkg/cmd/<resource>/<resource>.go` with a `NewCmd<Resource>(f *cmdutil.Factory) *cobra.Command` that groups the subcommands.
2. Create a subpackage per verb (`list`, `create`, `view`, `update`, `delete`) each exporting `NewCmd<Verb>(f *cmdutil.Factory) *cobra.Command`.
3. Register the group command in `pkg/cmd/root/root.go`.

## Command structure pattern

Every leaf command follows this pattern:

```go
package verb

func NewCmd<Verb>(f *cmdutil.Factory) *cobra.Command {
    var (
        flagOrg  string
        flagFoo  string
        // ...
    )

    cmd := &cobra.Command{
        Use:     "<verb>",
        Short:   "One-line description",
        Aliases: []string{"..."},   // optional, e.g. "ls" for list
        Example: `  prb <resource> <verb> ...`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Validate output flags (for list commands)
            // 2. Load config, get host + token
            // 3. Create api.Client
            // 4. Resolve --org (flag → config default)
            // 5. Interactive prompts if IOStreams.IsInteractive() and flags are missing
            // 6. Call API via client.Do(query, variables)
            // 7. Output: JSON via cmdutil.PrintJSON or table via cmdutil.NewTable
        },
    }

    cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
    // ... more flags ...

    return cmd
}
```

## Key conventions

- **GraphQL queries/mutations** are `const` strings declared at package level in the leaf command file.
- **Response types** are unexported structs in the leaf command file, shaped to match the GraphQL response.
- **Organization resolution**: every command that needs an org checks `--org` flag first, then falls back to `hc.Organization` from config. If both are empty, return an error telling the user to pass `--org` or run `prb auth login`.
- **Interactive prompts** use `github.com/charmbracelet/huh`. Gate them behind `f.IOStreams.IsInteractive()`. Always support full non-interactive use via flags.
- **Output format**: list commands support `--output json|table` via `cmdutil.AddOutputFlag` / `cmdutil.ValidateOutputFlag`. Default is table.
- **Pagination**: list commands use `api.Paginate[T]` with a `--limit` / `-L` flag (default 30). Show "Showing X of Y" on stderr when results are truncated.
- **Table output**: use `cmdutil.NewTable("COL", ...).Rows(rows...)`.
- **View commands** print detailed formatted output with lipgloss-styled labels and sections. They support `--output json|table` like list commands. Use `lipgloss.NewStyle()` for bold titles and dimmed labels.
- **Create/update/delete** commands print a single confirmation line to stdout (e.g. `"Created risk %s (%s)\n"`).
- **Delete commands** prompt for confirmation interactively; skip the prompt when `--yes` / `-y` is passed.
- **Flag naming**: use kebab-case (`--order-by`, `--inherent-likelihood`). Use `StringVar` / `IntVar` (not positional args) for all inputs.

## Dependencies

- CLI framework: `github.com/spf13/cobra`
- Interactive prompts: `github.com/charmbracelet/huh`
- Styled terminal output: `github.com/charmbracelet/lipgloss`
- All other dependencies follow the root AGENTS.md (same module, same style rules)
