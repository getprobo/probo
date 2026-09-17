# Release `@probo/cookie-banner-tcf`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@probo/cookie-banner-tcf/v*`
- **Version source**: `packages/cookie-banner-tcf/package.json`
- **Version bump**: `npm --workspace @probo/cookie-banner-tcf version <X.Y.Z> --no-git-tag-version`
- **Build**: `npm --workspace @probo/cookie-banner-tcf run build`
- **Changelog**: `packages/cookie-banner-tcf/CHANGELOG.md`
- **Files to stage**: `packages/cookie-banner-tcf/package.json`,
  `packages/cookie-banner-tcf/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-cookie-banner-tcf.yaml`
- **Path filter**: `packages/cookie-banner-tcf`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@probo/cookie-banner-tcf/v*')..HEAD --oneline \
  -- packages/cookie-banner-tcf
```

If empty or non-user-facing only, do not release this track.

## Notes

`packages/cookie-banner-tcf/build.mjs` reads `version` from `package.json`
and exposes it as the `__SDK_VERSION__` define. The TCF IIFE bundles
`@probo/cookie-banner`, so that package must be built first (`turbo` already
depends on `^build`). Run the TCF build after the version bump.

CI verifies the tag matches `package.json`, runs the build, publishes to
npm with provenance + SBOM, and creates a GitHub Release.
