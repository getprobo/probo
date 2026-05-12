# Release `@probo/n8n-nodes-probo`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@probo/n8n-nodes-probo/v*` (the `@` and `/` are
  valid in Git tag refs)
- **Version source**: `packages/n8n-node/package.json`
- **Version bump**: `npm --workspace @probo/n8n-nodes-probo version <X.Y.Z> --no-git-tag-version`
- **Build**: `npm --workspace @probo/n8n-nodes-probo run build`
- **Changelog**: `packages/n8n-node/CHANGELOG.md`
- **Files to stage**: `packages/n8n-node/package.json`,
  `packages/n8n-node/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-n8n-node.yaml`
- **Path filter**: `packages/n8n-node`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@probo/n8n-nodes-probo/v*')..HEAD --oneline \
  -- packages/n8n-node
```

If empty or non-user-facing only, do not release this track.

## Notes

CI verifies the tag matches `package.json`, publishes to npm with
provenance + SBOM, and creates a GitHub Release.
