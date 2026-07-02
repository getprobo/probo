# Syncing with upstream (getprobo/probo)
<!-- 
This is the playbook for pulling the latest changes from the original
`getprobo/probo` repo into our fork at `Advance-Datasec/probo-platform`,
preserving our Govrly rebrand and deployment customizations.

## Prerequisites — one-time setup

### 1. Two remotes

Make sure both remotes exist:

```bash
git remote -v
```

You need:

| Name              | URL                                                 | Purpose                  |
| ----------------- | --------------------------------------------------- | ------------------------ |
| `origin`          | `git@github-advance:Advance-Datasec/probo-platform.git` | Our team fork (push here) |
| `upstream`        | `https://github.com/getprobo/probo.git`             | The original repo (fetch only) |

If `upstream` is missing:

```bash
git remote add upstream https://github.com/getprobo/probo.git
```

If `origin` uses the bare `git@github.com:...` form, your default SSH key
must have write access. If you have multiple GitHub accounts, use the
`github-advance` host alias from `~/.ssh/config` instead — see
**SSH key setup** below.

### 2. SSH key setup (if you have multiple GitHub accounts)

Verify which user your default key authenticates as:

```bash
ssh -T git@github.com
# Hi <username>! ...
```

If that username doesn't have write access to `Advance-Datasec`, switch
`origin` to use the dedicated host alias. Your `~/.ssh/config` should
have something like:

```
Host github-advance
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519_advance
```

Then point `origin` through it:

```bash
git remote set-url origin git@github-advance:Advance-Datasec/probo-platform.git
ssh -T git@github-advance   # confirm the right user
``` -->

## The sync workflow

### 1. Stash any uncommitted work

```bash
git status
git stash push -m "wip before upstream sync $(date +%F)"
```

### 2. Fetch upstream

```bash
git git rev-list --left-right --count main...upstream/main
# e.g. "5  998" → 5 ahead, 998 behind
```

### 3. Merge — DO NOT rebase

```bash
git checkout main
git merge upstream/main --no-edit
```

**Why merge, not rebase:** Rebase replays every upstream commit (could
be hundreds or thousands) onto our base, and you'd resolve conflicts
multiple times across many of them. Merge resolves everything in a
single commit. We tried rebase once — it was 999 commits with no end
in sight. Don't repeat that mistake.

### 4. Resolve conflicts using these rules

Run `git diff --name-only --diff-filter=U` to list conflicted files.
Apply these defaults — they're what worked for the May 2026 sync.

#### Branding / UI files → take theirs (preserve Govrly rebrand)

```bash
git checkout --theirs <file>
git add <file>
```

Files in this category:
- `packages/ui/src/Atoms/Logo/Logo.tsx`
- `packages/ui/src/Atoms/Sidebar/Sidebar.tsx`
- `packages/ui/src/theme.css`
- `apps/console/src/main.tsx`, `apps/trust/src/main.tsx`
- `apps/console/src/pages/iam/organizations/_components/Sidebar.tsx`
- `apps/console/package.json` (name field rebrand)
- `pkg/webhook/sender.go` (`X-Govrly-Webhook-*` headers — this is a
  public API contract, do not revert)
- `README.md`, `contrib/claude/sandbox.md`
- `.dockerignore`

To take theirs and stage each of them:

```bash
git checkout --theirs packages/ui/src/Atoms/Logo/Logo.tsx && git add packages/ui/src/Atoms/Logo/Logo.tsx
git checkout --theirs packages/ui/src/Atoms/Sidebar/Sidebar.tsx && git add packages/ui/src/Atoms/Sidebar/Sidebar.tsx
git checkout --theirs packages/ui/src/theme.css && git add packages/ui/src/theme.css
git checkout --theirs apps/console/src/main.tsx && git add apps/console/src/main.tsx
git checkout --theirs apps/trust/src/main.tsx && git add apps/trust/src/main.tsx
git checkout --theirs apps/console/src/pages/iam/organizations/_components/Sidebar.tsx && git add apps/console/src/pages/iam/organizations/_components/Sidebar.tsx
git checkout --theirs apps/console/package.json && git add apps/console/package.json
git checkout --theirs pkg/webhook/sender.go && git add pkg/webhook/sender.go
git checkout --theirs README.md && git add README.md
git checkout --theirs contrib/claude/sandbox.md && git add contrib/claude/sandbox.md
git checkout --theirs .dockerignore && git add .dockerignore
```

#### Our deployment files → take ours

```bash
git checkout --ours <file>
git add <file>
```

Files in this category:
- `rds-ca-bundle.pem` (RDS TLS cert for production)
- `pkg/bootstrap/builder.go` (`SamplingRatio` and any other env-var
  wiring we added)
- `entrypoint.sh` (if we modified the bootstrap startup logic)

#### Modify/delete (DU) conflicts → keep team's file

These show up as "deleted in HEAD and modified in advance-datasec/main"
(or vice versa after a merge). Upstream removed a file we still use:

```bash
git add <file>   # the team's version is already on disk
```

Files seen so far:
- `.goreleaser.yaml` (upstream split releases into per-component workflows)
- `cfg/dev.yaml`
- `e2e/console/testdata/config.yaml`
- `apps/trust/CLAUDE.md`
- `apps/console/src/pages/organizations/documents/_components/UpdateVersionDialog.tsx`

If a file references types or functions that no longer exist after the
sync, you'll need to either port it forward or delete it — verify
compilation before assuming "keep" is safe.

#### `package-lock.json` → regenerate, never manually merge

```bash
git checkout --theirs package-lock.json
git add package-lock.json
# After committing the merge, regenerate:
npm install --package-lock-only
git add package-lock.json
git commit -m "Regenerate package-lock.json"
```

#### `Dockerfile.render` → merge by hand

This file has been an active source of bugs. Watch for:
- **Keep `ENV GOPRIVATE=go.probo.inc`** in the backend stage — without
  it, `go mod download` fails on private modules.
- **Schema paths:** upstream split single `pkg/server/api/*/v1/schema.graphql`
  files into multiple files under `graphql/`. The combined files are
  generated by `contrib/merge-graphql-schema.sh` and gitignored.
  Dockerfile.render must:
  1. Copy the `graphql/` source directories
  2. Copy `contrib/merge-graphql-schema.sh`
  3. Run the merge script to produce `schema.graphql` files
  4. Then run `npm ci` and `npm run relay`

  Reference the current `Dockerfile.render` for the exact sequence.

#### `.dockerignore` → must allow the merge script

If you take theirs blindly, `contrib/` is excluded, which breaks the
Docker build. Add the exception:

```
contrib
!contrib/merge-graphql-schema.sh
```

### 5. Commit the merge

```bash
git diff --name-only --diff-filter=U   # should be empty
git grep -l '^<<<<<<<\|^>>>>>>>'       # should find nothing
git commit --no-edit
```

### 6. Verify locally before pushing

```bash
# Go build (frontend dist folders may be missing — that's expected)
go build ./pkg/bootstrap/ ./pkg/probodconfig/

# Regenerate lockfile if you haven't already
npm install --package-lock-only

# Try a Docker build to catch path issues
docker build -f Dockerfile.render -t probo-test .
```

### 7. Push

**First time syncing (high-risk):** push to a branch and open a PR so
the team can review.

```bash
git push origin main:upstream-sync-$(date +%F)
# Open PR at:
# https://github.com/Advance-Datasec/probo-platform/pull/new/upstream-sync-<date>
```

**Routine sync (low-risk, you're confident):** push directly to main
once it's a clean fast-forward.

```bash
git fetch origin main
git rev-list --left-right --count main...origin/main
# Should be "N  0" — N ahead, 0 behind.
# If 0 behind, push:
git push origin main:main
# If behind, rebase your local work on top first:
git pull --rebase origin main
git push origin main:main
```

### 8. Restore your stashed work

```bash
git stash list
git stash pop
# Resolve any conflicts with the new upstream code
```

## Common pitfalls

- **Don't run `git pull --rebase`** on main when behind upstream — it
  silently switches the strategy to rebase. Always use explicit
  `git fetch` + `git merge`.
- **`schema.graphql` files don't exist in git** — they're generated.
  If Docker complains about missing schema files, the build is using
  the old paths instead of the `graphql/` dirs.
- **Lockfile drift bites every time.** Always run
  `npm install --package-lock-only` after merging `package.json`
  changes from upstream.
- **`go.mod` may pick up the upstream module path.** If imports go from
  `go.probo.inc/probo/...` to something else (or vice versa), watch for
  build errors and update Dockerfile `GOPRIVATE` accordingly.
- **Render deploys from `main`.** Each push to `main` triggers a build.
  If you're iterating on a Docker fix, use a branch first and only
  fast-forward main once it's green.

## When in doubt

Compare with the `getprobo/probo` README and Dockerfile to see how
they're building things today. Their CI workflows under `.github/workflows/`
are the source of truth for build steps.
