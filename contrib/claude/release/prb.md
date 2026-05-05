# Release `prb` (CLI)

Entrypoint for releasing the `prb` CLI. Read [README.md](./README.md)
first for overall flow, changelog rules, and the
non-empty-track guarantee.

## Track facts

- **Tag pattern**: `prb/v*`
- **Version source**: `cmd/prb/VERSION` (contains only `X.Y.Z`)
- **Changelog**: `cmd/prb/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-prb.yaml`
- **Path filter** (for log/scoping): `cmd/prb pkg/cli pkg/cmd`

## Steps

1. From a clean `main`, list commits since the last `prb` tag:

   ```shell
   git log $(git describe --tags --abbrev=0 --match='prb/v*')..HEAD --oneline \
     -- cmd/prb pkg/cli pkg/cmd
   ```

   If the list is empty (or contains only non-user-facing commits), do
   not release this track.

2. Decide the version bump (PATCH for fixes, MINOR for features).
3. Bump the version in `cmd/prb/VERSION` (the file contains a single
   `X.Y.Z` line — no trailing newline conventions beyond what is already
   there).
4. Write the new entry in `cmd/prb/CHANGELOG.md` following the rules in
   [README.md](./README.md#2-writing-a-changelog-entry).
5. Show the user the changelog diff and the new version. Wait for
   confirmation.
6. Stage only `cmd/prb/VERSION` and `cmd/prb/CHANGELOG.md`. Commit
   subject: `Release prb/v<version>`. No body.
7. Annotated tag: `git tag -a prb/v<version> -m "prb/v<version>"`.
8. Push: `git push origin main --follow-tags`.

CI workflow `release-prb.yaml` builds binaries for 9 OS/arch targets,
publishes a GitHub Release, and updates the Homebrew formula at
`getprobo/homebrew-tap`.
