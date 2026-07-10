# Release `@probo/skills`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@probo/skills/v*`
- **Version source**: `packages/skills/package.json`
- **Version bump**: `npm --workspace @probo/skills version <X.Y.Z> --no-git-tag-version`
- **Validate**: `npm --workspace @probo/skills run validate`
- **Changelog**: `packages/skills/CHANGELOG.md`
- **Files to stage**: `packages/skills/package.json`,
  `packages/skills/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-skills.yaml`
- **Path filter**: `packages/skills`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@probo/skills/v*')..HEAD --oneline \
  -- packages/skills
```

If empty or non-user-facing only, do not release this track.

## Notes

There is no build step. Run `validate` after the version bump to catch manifest
or structural errors before tagging. CI runs the same validation, publishes to
npm with provenance + SBOM, and creates a GitHub Release.
