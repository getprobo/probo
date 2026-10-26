# CLI Command Patterns

CLI commands use [cobra](https://github.com/spf13/cobra) with `pkg/cmd/cmdutil.Factory` for shared dependencies. Each resource gets a group command with verb subcommands (`list`, `create`, `view`, `update`, `delete`).

## Directory structure

```
cmd/prb/main.go                           # Binary entry point
pkg/cmd/root/                             # Root command, registers all subcommands
pkg/cmd/<resource>/<resource>.go          # Group command, wires verbs
pkg/cmd/<resource>/list/list.go           # List verb
pkg/cmd/<resource>/create/create.go       # Create verb
pkg/cmd/<resource>/view/view.go           # View verb
pkg/cmd/<resource>/update/update.go       # Update verb
pkg/cmd/<resource>/delete/delete.go       # Delete verb
pkg/cmd/cmdutil/                          # Factory, flags, output helpers
pkg/cmd/iostreams/                        # Terminal I/O abstraction
pkg/cli/api/                              # GraphQL client, pagination
pkg/cli/config/                           # Config file management (hosts, tokens, default org)
```

Register group commands in `pkg/cmd/root/root.go` with `cmd.AddCommand()`.

## Agent skills

`prb skills` ships Agent Skills inside the binary (`pkg/cmd/skill/content`). `list` and `show` read them with no network call. `install` copies each `<skill>/SKILL.md` tree into `--dir`, or into the user skill directories for Cursor, Claude, Codex, and OpenCode (`--target`). `--force` replaces a file whose bytes differ from the copy in the binary. Identical files are left in place. These skills describe how to call `prb`. The MCP workflows in `@probo/skills` stay in that package.

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
			// 2. f.GraphQLClient (config, host, token, refresh)
			// 3. Resolve --org (flag or config default)
			// 4. Call API
			// 5. cmdutil.PrintList
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
```

## Pagination with `api.Paginate[T]`

List commands share the client, the node extract, and the printer. `Factory.GraphQLClient` loads the active host. Pass `cmdutil.ConsoleEndpoint` for organization records and `cmdutil.ConnectEndpoint` for the signed-in user (`org list`, SCIM events).

```go
client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
if err != nil {
	return err
}

resources, totalCount, err := api.Paginate(
	client,
	listQuery,
	variables,
	flagLimit,
	api.NodeConnection[resource](
		"resources",
		"Organization",
		fmt.Sprintf("organization %s not found", flagOrg),
	),
)
```

`api.NodeConnection` reads `data.node.<field>`. A null node returns the not-found string. A different `__typename` returns `expected <type> node, got <actual>`. `api.Paginate` wraps that error as `cannot parse response`.

Use the variant that matches the response:

- `NodeConnectionOrMissing` when a wrong `__typename` is also not-found (risk and treatment-plan lists choose the parent at runtime).
- `NodeConnectionRequired` when a null connection means the list is unavailable (`employee-portal list`).
- `NodeNestedConnection` when the connection sits under a child object (`node.trustCenter.commitmentGroups`).
- `ViewerConnection` for `data.viewer.<field>` (`org list`).
- `RootConnection` for a top-level field (GVL catalog).

`NodeSlice` is a direct call, not an `api.Paginate` extract. It reads a JSON array from `data.node.<field>` for a non-Relay list (`connector list`). Pass `len(items)` as the total to `PrintList` so the truncation line stays absent.

## Output formatting

**JSON output:** `--json` is shorthand for `--output json` / `-o json`. Do not combine `--json` with a different `--output` value.

**List output:** `PrintList` writes JSON (a null slice becomes `[]`), the empty sentence, the table, and `Showing n of m <noun>` on stderr when the server total is larger than the page. The truncation line is written for both JSON and table output so agents still see a truncated page. `Empty` and `Noun` can differ (`No access review campaigns found.` vs `campaigns`).

```go
return cmdutil.PrintList(
	f.IOStreams,
	*flagOutput,
	resources,
	totalCount,
	cmdutil.ListView[resource]{
		Empty:   "No resources found.",
		Noun:    "resources",
		Headers: []string{"ID", "NAME"},
		Row: func(r resource) ([]string, error) {
			return []string{r.ID, r.Name}, nil
		},
	},
)
```

**View detail output with lipgloss:**
```go
bold := lipgloss.NewStyle().Bold(true)
label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(r.Name))
_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), r.ID)
_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(r.CreatedAt))
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
flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
if err != nil {
	return err
}
```

## Flag conventions

- Use kebab-case for command names and flags: `third-party`, `--order-by`, `--inherent-likelihood`. Organization-scoped commands use `--org`. `--organization` is the same flag. Register it with `cmdutil.AddOrganizationFlag`. `connector connect --organization` is a boolean and is the exception.
- Common options: `-L` (limit), `-o` (output), `--json` (JSON output), `-q` (query), `-y` (yes)
- `StringVar`/`IntVar`/`BoolVar` for flags, positional args only for IDs in view/update/delete
- Use `cmd.MarkFlagRequired()` for mandatory flags

## IOStreams

```go
f.IOStreams.Out     // stdout — primary output
f.IOStreams.ErrOut  // stderr — status messages, truncation info
f.IOStreams.IsInteractive()  // true if TTY and not forced non-interactive
```

Environment variables: `PROBO_NO_INTERACTIVE=1`, `CI=true`, `TERM=dumb` (non-interactive), `NO_COLOR` (disable color).

## OAuth device login (`prb auth login`)

The well-known CLI OAuth client (`config.CLIClientID`) requests `config.CLIClientScopes` at device authorization. Keep three places in sync when adding `v1:*` scopes:

1. Scope constants and `OAuth2ScopeSet()` registration in the owning package
2. `iam_oauth2_clients.scopes` for `CLIClientID` (SQL migration)
3. `CLIClientScopes` in `pkg/cli/config/config.go`

After scope changes ship, users must re-authenticate so new tokens carry the updated scopes:

```sh
prb auth logout --hostname <host>
prb auth login --hostname <host>
```

Existing tokens may be backfilled by migration; fresh logins need the client row and `CLIClientScopes` updated.

### Local development against probod

```sh
make stack-up
make build
make dev-config
bin/probod -cfg-file cfg/dev.yaml   # API + OAuth at http://localhost:8080

bin/prb auth login --hostname http://localhost:8080
```

Use `http://` explicitly — hosts without a scheme default to HTTPS. Config is stored under the OS user config dir (`prb/config.yaml`). Override with `PROBO_HOST` and `PROBO_TOKEN` for scripting; see `contrib/seed.sh` for an OAuth access token bootstrap.

## New resource command checklist

1. **Group command** — `pkg/cmd/<resource>/<resource>.go` with `NewCmd<Resource>(f)`, wiring all verb subcommands
2. **Leaf commands** — one file per verb in `pkg/cmd/<resource>/<verb>/<verb>.go`
3. **Each leaf file** — MIT license header, GraphQL const, unexported response struct, `NewCmd<Verb>(f)` function
4. **Register in root** — import and `cmd.AddCommand()` in `pkg/cmd/root/root.go`
5. **API surface** — update GraphQL schema, MCP tools, and e2e tests
