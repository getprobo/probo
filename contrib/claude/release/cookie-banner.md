# Release `@probo/cookie-banner`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@probo/cookie-banner/v*`
- **Version source**: `packages/cookie-banner/package.json`
- **Version bump**: `npm --workspace @probo/cookie-banner version <X.Y.Z> --no-git-tag-version`
- **Build**: `npm --workspace @probo/cookie-banner run build`
- **Changelog**: `packages/cookie-banner/CHANGELOG.md`
- **Files to stage**: `packages/cookie-banner/package.json`,
  `packages/cookie-banner/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-cookie-banner.yaml`
- **Path filter**: `packages/cookie-banner`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@probo/cookie-banner/v*')..HEAD --oneline \
  -- packages/cookie-banner
```

If empty or non-user-facing only, do not release this track.

## Notes

`packages/cookie-banner/build.mjs` reads `version` from `package.json`
and exposes it as the `__SDK_VERSION__` define. The SDK uses this at
runtime, so the build **must** run after the version bump. The release CI
also runs the build, but running it locally catches compile errors before
tagging and ensures tracked side-effects (`package-lock.json`) are part
of the release commit.

CI verifies the tag matches `package.json`, runs the build, publishes to
npm with provenance + SBOM, and creates a GitHub Release.
