# Release `@probo/n8n-nodes-probo`

Entrypoint for releasing the n8n nodes package. Read
[README.md](./README.md) first for overall flow, changelog rules, and the
non-empty-track guarantee.

## Track facts

- **Tag pattern**: `@probo/n8n-nodes-probo/v*` (the `@` and `/` are valid
  in Git tag refs)
- **Version source**: `packages/n8n-node/package.json`
- **Changelog**: `packages/n8n-node/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-npm-n8n-node.yaml`
- **Path filter** (for log/scoping): `packages/n8n-node`

## Steps

1. From a clean `main`, list commits since the last
   `@probo/n8n-nodes-probo` tag:

   ```shell
   git log $(git describe --tags --abbrev=0 --match='@probo/n8n-nodes-probo/v*')..HEAD --oneline \
     -- packages/n8n-node
   ```

   If the list is empty (or contains only non-user-facing commits), do
   not release this track.

2. Decide the version bump (PATCH for fixes, MINOR for features).
3. Bump the version using npm so `package.json` and `package-lock.json`
   stay consistent:

   ```shell
   npm --workspace @probo/n8n-nodes-probo version <X.Y.Z> --no-git-tag-version
   ```

4. Run the workspace build to confirm it succeeds with the new version
   (and to surface any compile errors before we tag):

   ```shell
   npm --workspace @probo/n8n-nodes-probo run build
   ```

5. Write the new entry in `packages/n8n-node/CHANGELOG.md` following the
   rules in [README.md](./README.md#2-writing-a-changelog-entry).
6. Show the user the changelog diff and the new version. Wait for
   confirmation.
7. Stage the files modified by the version bump and changelog edit
   (typically `packages/n8n-node/package.json`,
   `packages/n8n-node/CHANGELOG.md`, and `package-lock.json`). Commit
   subject: `Release @probo/n8n-nodes-probo/v<version>`. No body.
8. Annotated tag:
   `git tag -a @probo/n8n-nodes-probo/v<version> -m "@probo/n8n-nodes-probo/v<version>"`.
9. Push: `git push origin main --follow-tags`.

CI workflow `release-npm-n8n-node.yaml` verifies the tag matches
`package.json`, publishes to npm with provenance + SBOM, and creates a
GitHub Release.
