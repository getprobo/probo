# Release `@probo/claude-plugin`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@probo/claude-plugin/v*`
- **Version source**: `packages/claude-plugin/package.json`
- **Version bump**: `npm --workspace @probo/claude-plugin version <X.Y.Z> --no-git-tag-version`
- **Validate**: `npm --workspace @probo/claude-plugin run validate`
- **Changelog**: `packages/claude-plugin/CHANGELOG.md`
- **Files to stage**: `packages/claude-plugin/package.json`,
  `packages/claude-plugin/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-claude-plugin.yaml`
- **Path filter**: `packages/claude-plugin`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@probo/claude-plugin/v*')..HEAD --oneline \
  -- packages/claude-plugin
```

If empty or non-user-facing only, do not release this track.

## Notes

There is no build step. Run `validate` after the version bump to catch manifest
or structural errors before tagging. CI runs the same validation, publishes to
npm with provenance + SBOM, and creates a GitHub Release.
