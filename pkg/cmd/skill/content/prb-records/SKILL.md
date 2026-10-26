---
name: prb-records
description: Create, list, view, update, and delete Probo records with prb. Use for frameworks, controls, risks, documents, tasks, assets, data, vendors, audits, findings, obligations, processing activities, rights requests, business functions, AI systems, or webhooks.
compatibility: Requires the prb binary, a signed-in host, and an organization id. Read using-prb first.
---

# Change records with prb

Read `using-prb` before calling prb. Read `prb-auth` when no organization is
selected.

Required create flags are listed in `references/create.md`. When this skill is
not installed on disk, print that file with:

```sh
prb skills show prb-records references/create.md
```

Flag help wins when it disagrees with the reference. Copy enum values from
`prb <command> create --help`. Do not invent a value that is not listed there.

## Sequence

1. Resolve the organization (`prb-auth`). Controls are the exception: pass
   `--framework` and omit `--org`.
2. Create with every required flag. The new id is on stdout.
3. `prb <command> view <id> --json` and confirm the field you wrote.
4. Update with `prb <command> update <id>` plus at least one changed flag.
5. Delete with `prb <command> delete <id> --yes` only after the user asked to
   delete that record.

## Ids used as inputs

- `--owner` is a profile id from `prb user list --org <org> --json`.
- `--framework` on audits and controls is a framework id from
  `prb framework create` or `prb framework list --json`.
- `--third-party-ids` takes vendor ids. The kebab-case flag is the one to
  pass.

## Documents and webhooks

Document title, type, and classification are fields of the document.
`prb document view <id> --json` includes `title`.

`prb webhook list` requires `--org`. Create takes `--url` and one or more
`--event` flags. Event names come from `prb webhook create --help`.
