---
name: using-prb
description: Call the prb CLI without inventing flags. Use when an agent runs prb, signs in, chooses an organization, or creates, lists, views, updates, or deletes Probo records such as frameworks, controls, risks, documents, tasks, assets, data, vendors, audits, findings, obligations, processing activities, rights requests, business functions, AI systems, or webhooks.
compatibility: Requires the prb binary on PATH. Setting PROBO_HOST and PROBO_TOKEN lets every command skip login entirely. Otherwise `prb auth login` runs the device flow; it opens a browser when the session is interactive, and prints a URL and code to approve manually when it is not.
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
Read `references/commands.md` in this skill for the command groups. Required
create flags are listed in `references/create.md`. Flag help wins when it
disagrees with a reference.

## Call shape

1. Export `PROBO_NO_INTERACTIVE=1` so prb does not prompt. In that mode every
   required flag must be present.
2. Results go to stdout. Errors go to stderr and the process exits non-zero.
   Do not treat stderr warnings as failure when the exit code is 0.
3. Request JSON with `--json`. `--output json` and `-o json` are the same
   choice. Combining `--json` with any other `--output` value fails.
   `--json` is registered on list and view commands; a few create and
   update commands have it too. Check `--help` to confirm it exists before
   relying on it elsewhere.
4. `list` is also `ls`.
5. Command and flag names are kebab-case. The vendor command is
   `prb third-party`. `tp` is its short alias.
6. View, update, and delete take the resource id printed by create. That id
   is the first token that parses as a Probo id.
7. Update with only the id fails with `at least one field must be specified`.
   Pass the fields that change.
8. Delete in non-interactive mode requires `--yes`. Without it, prb exits
   non-zero and deletes nothing.
9. `--limit` / `-L` is the maximum number of rows to return; a list with
   fewer matches returns fewer rows. It must be greater than 0.
   stderr prints `Showing N of M` when more rows exist, and
   `Next cursor: <cursor>` when another page exists. Repeat the same
   command and filters with `--after <cursor>`. Do not treat one page as
   the full set. On commands that expose it, `--order-direction` accepts
   `ASC` or `DESC`.

## Sign in

Settle the host and the organization before any other `prb` command.

1. Run `prb auth status`.
2. An active host prints `(active)` and `Token: set` or `Token: not set`.
   The command never prints the token. Do not ask the user to paste a token
   into the chat when status already says `Token: set`.
3. `PROBO_HOST` selects the host for this process. `PROBO_TOKEN` supplies the
   bearer token. A token without a host is an error.
4. Hosts without a scheme are HTTPS. A local server needs an explicit
   `http://` URL, for example `http://localhost:8080`.
5. When status shows no active host, or an active host with `Token: not
   set`, run `prb auth login`. Pass `--hostname` when the user names a
   region or a self-hosted instance (`eu.probo.com`, `us.probo.com`, or the
   instance host). Pass `--org` when the user already knows the default
   organization id.

Then choose the organization:

1. Run `prb org list --json`.
2. Use the organization the user named, or the only organization in the list.
3. When several organizations match, print the id and name of each and ask
   which one to use.
4. Pass that id as `--org` on later commands. `--organization` is the same
   flag. Pass it once to `prb auth login --org` so it becomes the saved
   default.
5. A missing organization fails with `organization is required; pass --org or
   --organization, or set a default with 'prb auth login'`.

## Organization

Commands that select an organization take `--org`. `--organization` sets the
same value. Lists of a child record take the parent id instead: `control`
takes `--framework`, `evidence` takes `--internal-control`, consent records
and cookie categories take the banner, and access-review entries take the
campaign.

`connector connect --organization` is a boolean. The Probo organization id
on that command is `--org`.

## Change a record

1. Pass the organization as `--org` or `--organization`. Controls are the
   exception: pass `--framework` and omit both.
2. Create with every required flag from `references/create.md`. The new id
   is on stdout.
3. `prb <command> view <id> --json` and confirm the field you wrote.
4. Update with `prb <command> update <id>` plus at least one changed flag.
5. Delete with `prb <command> delete <id> --yes` only after the user asked to
   delete that record.

Ids used as inputs:

- `--owner` is a profile id from `prb user list --org <org> --json`.
- `--framework` on audits and controls is a framework id from
  `prb framework create` or `prb framework list --json`.
- `--third-party-ids` takes vendor ids. The kebab-case flag is the one to
  pass.

Document title, type, and classification are fields of the document.
`prb document view <id> --json` includes `title`.

`prb webhook list` requires `--org`. Create takes `--url` and one or more
`--event` flags. `--help` only shows example event names, not the full set.
An invalid `--event` value returns an error listing every valid name; use
that to confirm a name before retrying.

## After writing

Confirm the change with `prb <command> view <id> --json` or
`prb <command> list --org <id> --json`. A deleted id is absent from the list.
Controls are the exception from "Change a record": verify with
`prb control list --framework <framework-id> --json`, not `--org`.
