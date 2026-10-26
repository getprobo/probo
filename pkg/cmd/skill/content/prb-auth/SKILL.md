---
name: prb-auth
description: Sign prb in to Probo and choose the organization before other prb commands. Use when prb auth status is not active, PROBO_HOST or PROBO_TOKEN is unset, or more than one organization is available.
compatibility: Requires the prb binary on PATH. Login opens a browser for the device flow unless PROBO_HOST and PROBO_TOKEN are already set.
---

# Sign in with prb

Settle the host and the organization before any other `prb` command. Read
`using-prb` for output, confirmation, and the rule against inventing flags.

## Host

1. Run `prb auth status`.
2. An active host prints `(active)` and `Token: set` or `Token: not set`.
   The command never prints the token. Do not ask the user to paste a token
   into the chat when status already says `Token: set`.
3. `PROBO_HOST` selects the host for this process. `PROBO_TOKEN` supplies the
   bearer token. A token without a host is an error.
4. Hosts without a scheme are HTTPS. A local server needs an explicit
   `http://` URL, for example `http://localhost:8080`.
5. When status shows no active host, run `prb auth login`. Pass `--hostname`
   when the user names a region or a self-hosted instance (`eu.probo.com`,
   `us.probo.com`, or the instance host). Pass `--org` when the user already
   knows the default organization id.

## Organization

1. Run `prb org list --json`.
2. Use the organization the user named, or the only organization in the list.
3. When several organizations match, print the id and name of each and ask
   which one to use.
4. Pass that id as `--org` on later commands. `--organization` is the same
   flag. Pass it once to `prb auth login --org` so it becomes the saved
   default.
5. A missing organization fails with `organization is required; pass --org or
   --organization, or set a default with 'prb auth login'`.
