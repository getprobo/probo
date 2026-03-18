# CLI Command Patterns

CLI commands use [cobra](https://github.com/spf13/cobra) with `pkg/cmd/cmdutil.Factory` for shared dependencies. Each resource gets a group command with verb subcommands (`list`, `create`, `view`, `update`, `delete`).

## Directory structure

```
pkg/cmd/<resource>/<resource>.go          # Group command, wires verbs
pkg/cmd/<resource>/list/list.go           # List verb
pkg/cmd/<resource>/create/create.go       # Create verb
pkg/cmd/<resource>/view/view.go           # View verb
pkg/cmd/<resource>/update/update.go       # Update verb
pkg/cmd/<resource>/delete/delete.go       # Delete verb
pkg/cmd/cmdutil/                          # Factory, flags, output helpers
pkg/cmd/iostreams/                        # Terminal I/O abstraction
pkg/cli/api/                              # GraphQL client, pagination
```

Register group commands in `pkg/cmd/root/root.go` with `cmd.AddCommand()`.

## Leaf command pattern

Every leaf command follows this structure:

```go
package list

const listQuery = `query($id: ID!, $first: Int, $after: CursorKey) { ... }`

type listResponse struct { ... } // unexported, shaped to match GraphQL response

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List resources",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Validate output flag
			// 2. Load config, get host + token
			// 3. Create api.Client
			// 4. Resolve --org (flag or config default)
			// 5. Call API
			// 6. Output results
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
```

## Pagination with `api.Paginate[T]`

```go
resources, totalCount, err := api.Paginate(
	client,
	listQuery,
	variables,
	flagLimit,
	func(data json.RawMessage) (*api.Connection[resource], error) {
		var resp struct {
			Node *struct {
				Typename  string                   `json:"__typename"`
				Resources api.Connection[resource]  `json:"resources"`
			} `json:"node"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		if resp.Node == nil {
			return nil, fmt.Errorf("organization %s not found", flagOrg)
		}
		return &resp.Node.Resources, nil
	},
)
```

## Output formatting

**JSON output:**
```go
if *flagOutput == cmdutil.OutputJSON {
	return cmdutil.PrintJSON(f.IOStreams.Out, resources)
}
```

**Table output:**
```go
rows := make([][]string, 0, len(resources))
for _, r := range resources {
	rows = append(rows, []string{r.ID, r.Name})
}
t := cmdutil.NewTable("ID", "NAME").Rows(rows...)
_, _ = fmt.Fprintln(f.IOStreams.Out, t)
```

**View detail output with lipgloss:**
```go
bold := lipgloss.NewStyle().Bold(true)
label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(r.Name))
_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), r.ID)
_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(r.CreatedAt))
```

**Truncation message** (stderr, not stdout):
```go
if totalCount > len(resources) {
	_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d resources\n", len(resources), totalCount)
}
```

## Interactive prompts with `charmbracelet/huh`

Gate all prompts behind interactivity check, then validate:

```go
if f.IOStreams.IsInteractive() {
	if flagName == "" {
		err := huh.NewInput().Title("Resource name").Value(&flagName).Run()
		if err != nil {
			return err
		}
	}

	if flagCategory == "" {
		err := huh.NewSelect[string]().
			Title("Category").
			Options(
				huh.NewOption("Cloud Provider", "CLOUD_PROVIDER"),
				huh.NewOption("SaaS", "SAAS"),
			).
			Value(&flagCategory).Run()
		if err != nil {
			return err
		}
	}
}

if flagName == "" {
	return fmt.Errorf("name is required; pass --name or run interactively")
}
```

Available prompt types: `huh.NewInput()` (text), `huh.NewText()` (multiline), `huh.NewSelect[T]()` (dropdown), `huh.NewConfirm()` (yes/no).

## Update commands

Only include fields that were explicitly changed:

```go
input := map[string]any{"id": args[0]}

if cmd.Flags().Changed("name") {
	input["name"] = flagName
}
if cmd.Flags().Changed("description") {
	input["description"] = flagDescription
}

if len(input) == 1 {
	return fmt.Errorf("at least one field must be specified for update")
}
```

## Delete commands

Require confirmation via `--yes` flag or interactive prompt:

```go
if !flagYes {
	if !f.IOStreams.IsInteractive() {
		return fmt.Errorf("cannot delete resource: confirmation required, use --yes to confirm")
	}
	var confirmed bool
	err := huh.NewConfirm().Title(fmt.Sprintf("Delete %s?", args[0])).Value(&confirmed).Run()
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}
}
```

## Organization resolution

```go
if flagOrg == "" {
	flagOrg = hc.Organization
}
if flagOrg == "" {
	return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
}
```

## Flag conventions

- Use kebab-case: `--order-by`, `--inherent-likelihood`
- Short flags for common options: `-L` (limit), `-o` (output), `-q` (query), `-y` (yes)
- `StringVar`/`IntVar`/`BoolVar` for flags, positional args only for IDs in view/update/delete
- Use `cmd.MarkFlagRequired()` for mandatory flags

## IOStreams

```go
f.IOStreams.Out     // stdout — primary output
f.IOStreams.ErrOut  // stderr — status messages, truncation info
f.IOStreams.IsInteractive()  // true if TTY and not forced non-interactive
```

Environment variables: `PROBO_NO_INTERACTIVE=1`, `CI=true`, `TERM=dumb` (non-interactive), `NO_COLOR` (disable color).

## New resource command checklist

1. **Group command** — `pkg/cmd/<resource>/<resource>.go` with `NewCmd<Resource>(f)`, wiring all verb subcommands
2. **Leaf commands** — one file per verb in `pkg/cmd/<resource>/<verb>/<verb>.go`
3. **Each leaf file** — ISC license header, GraphQL const, unexported response struct, `NewCmd<Verb>(f)` function
4. **Register in root** — import and `cmd.AddCommand()` in `pkg/cmd/root/root.go`
5. **API surface** — update GraphQL schema, MCP tools, and e2e tests
