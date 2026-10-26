---
name: using-prb
description: Call the prb CLI without inventing flags. Use when an agent runs prb, changes a Probo organization from the terminal, or needs JSON output, confirmation, or command discovery.
compatibility: Requires the prb binary on PATH and either an existing login or PROBO_HOST plus PROBO_TOKEN.
---

# Use prb

`prb` is the Probo CLI. Prefer it over hand-written GraphQL. Command help is
the source of truth for flags. This skill is the source of truth for how to
call the CLI.

Before the first call to a command in this session, read its help:

```sh
prb <command> --help
prb <command> <verb> --help
```

Pass only flags that appear there. Copy enum values from the help text.

When this skill is not installed on disk, read a bundled reference with
`prb skills show using-prb references/commands.md`.

## Call shape

1. Export `PROBO_NO_INTERACTIVE=1` so prb does not prompt. In that mode every
   required flag must be present.
2. Results go to stdout. Errors go to stderr and the process exits non-zero.
   Do not treat stderr warnings as failure when the exit code is 0.
3. Request JSON with `--json`. `--output json` and `-o json` are the same
   choice. Combining `--json` with any other `--output` value fails.
4. `list` is also `ls`.
5. Command and flag names are kebab-case. The vendor command is
   `prb third-party`. `tp` is its short alias.
6. View, update, and delete take the resource id printed by create. That id
   is the first token that parses as a Probo id.
7. Update with only the id fails with `at least one field must be specified`.
   Pass the fields that change.
8. Delete in non-interactive mode requires `--yes`. Without it, prb exits
   non-zero and deletes nothing.
9. `--limit` / `-L` caps a list. A line on stderr of the form
   `Showing N of M` means the list was truncated. Raise `--limit` or page
   before concluding the set is complete.

## Organization

Commands that select an organization take `--org`. `--organization` sets the
same value. `prb-auth` covers login and how to choose the organization.
Lists of a child record take the parent id instead: `control` takes
`--framework`, `evidence` takes `--internal-control`, consent records and
cookie categories take the banner, and access-review entries take the
campaign.

`connector connect --organization` is a boolean. The Probo organization id
on that command is `--org`.

## After writing

Confirm the change with `prb <command> view <id> --json` or
`prb <command> list --org <id> --json`. A deleted id is absent from the list.
