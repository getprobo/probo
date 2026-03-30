# Probo -- Go Backend -- pkg/cmd (CLI)

> Module-specific patterns that differ from stack-wide conventions.
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md).

## Purpose

Implements the `prb` CLI tool using cobra. All commands communicate with the Probo backend exclusively via GraphQL over HTTPS. No direct database access.

## Directory Structure (feature-sliced)

Unlike the flat package pattern used elsewhere in the backend, the CLI uses a feature-slice layout:

```
pkg/cmd/
  root/root.go              # Registers all top-level resource groups
  <resource>/
    <resource>.go            # Group command, aggregates verb sub-commands
    <verb>/
      <verb>.go              # Leaf command, owns RunE logic
```

Example: `pkg/cmd/vendor/vendor.go` groups `list/`, `create/`, `view/`, `update/`, `delete/`.

## Command Construction Pattern

Every leaf command receives `*cmdutil.Factory` as its sole dependency:

```go
// See pattern from: contrib/claude/cli.md
func NewCmdCreateVendor(f *cmdutil.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a vendor",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Load config
            cfg, err := f.Config()
            if err != nil { return err }

            // 2. Resolve host + token
            hc, err := cfg.DefaultHost()
            if err != nil { return err }

            // 3. Resolve organization
            orgID := cmd.Flag("org").Value.String()
            if orgID == "" { orgID = hc.Organization }

            // 4. Create API client
            client := api.NewClient(hc.Host, hc.Token, "/api/console/v1/graphql", 0)

            // 5. Execute GraphQL
            result, err := client.Do(cmd.Context(), query, variables)

            // 6. Print output
            fmt.Fprintln(f.IOStreams.Out, "Vendor created:", id)
            return nil
        },
    }
    return cmd
}
```

## GraphQL Queries as Constants

GraphQL queries are defined as `const` strings at the package level in leaf command files. Response types are unexported structs local to the leaf package:

```go
// See pattern from: contrib/claude/cli.md
const createVendorMutation = `
    mutation CreateVendor($input: CreateVendorInput!) {
        createVendor(input: $input) {
            vendor { id name }
        }
    }
`

type createVendorResponse struct {
    CreateVendor struct {
        Vendor struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        } `json:"vendor"`
    } `json:"createVendor"`
}
```

## Flag Conventions

- Flag naming: kebab-case (`--order-by`, `--inherent-likelihood`)
- Standard short flags: `-L` (limit), `-o` (output), `-q` (query), `-y` (yes)
- Organization resolution: `--org` flag first, then `hc.Organization` from config
- Update commands: only include fields where `cmd.Flags().Changed()` is true
- Delete commands: require `--yes` flag or interactive `huh.NewConfirm()` prompt
- List commands: support `--output json|table` (default is table)

## Output Formatting

- Table output via `cmdutil.NewTable` (pre-styled lipgloss/table)
- JSON output via `cmdutil.PrintJSON`
- Truncation message goes to stderr, not stdout
- Single confirmation line for create/update/delete

## Pagination

List commands use `api.Paginate[T]` with `--limit` / `-L` flag (default 30):

```go
// See: pkg/cli/api/pagination.go
nodes, totalCount, err := api.Paginate[vendorNode](
    cmd.Context(),
    client,
    listVendorsQuery,
    variables,
    func(raw json.RawMessage) (*api.Connection[vendorNode], error) { ... },
)
```

## Interactive Prompts

When `f.IOStreams.IsInteractive()` is true, commands use `charmbracelet/huh` for prompts. Non-interactive mode (piped input) skips prompts and requires all data via flags.

## Registration

New resource group commands must be registered in `pkg/cmd/root/root.go`.
