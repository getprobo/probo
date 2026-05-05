# Release `probod` (server group)

Entrypoint for releasing the `probod` server group. Read
[README.md](./README.md) first for overall flow, changelog rules, and the
non-empty-track guarantee.

This track ships `probod`, `@probo/console`, `@probo/trust`, and
`@probo/ui` together as the Docker image and accompanying binary archive.
They share the same version.

## Track facts

- **Tag pattern**: `probod/v*`
- **Version source**: `cmd/probod/VERSION` (contains only `X.Y.Z`)
- **Changelog**: `cmd/probod/CHANGELOG.md` (covers all four components)
- **Workflow**: `.github/workflows/release-probod.yaml`
- **Path filter** (for log/scoping):
  `cmd/probod apps/console apps/trust packages/ui pkg`

## Steps

1. From a clean `main`, list commits since the last `probod` tag:

   ```shell
   git log $(git describe --tags --abbrev=0 --match='probod/v*')..HEAD --oneline \
     -- cmd/probod apps/console apps/trust packages/ui pkg
   ```

   If the list is empty (or contains only non-user-facing commits), do
   not release this track.

2. Decide the version bump (PATCH for fixes, MINOR for features).
3. Bump the version in `cmd/probod/VERSION`.
4. Write the new entry in `cmd/probod/CHANGELOG.md` covering changes
   across `probod`, `@probo/console`, `@probo/trust`, and `@probo/ui`.
   Follow the rules in
   [README.md](./README.md#2-writing-a-changelog-entry).
5. Show the user the changelog diff and the new version. Wait for
   confirmation.
6. Stage only `cmd/probod/VERSION` and `cmd/probod/CHANGELOG.md`.
   Commit subject: `Release probod/v<version>`. No body.
7. Annotated tag: `git tag -a probod/v<version> -m "probod/v<version>"`.
8. Push: `git push origin main --follow-tags`.

CI workflow `release-probod.yaml` builds the frontends, builds the Go
binaries, builds and pushes the multi-arch image to
`ghcr.io/getprobo/probo:probod-v<version>` (and `:latest`), runs Trivy +
cosign + attestations, and publishes the GitHub Release.
