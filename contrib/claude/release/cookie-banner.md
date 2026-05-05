# Release `@probo/cookie-banner`

Entrypoint for releasing the cookie banner SDK. Read
[README.md](./README.md) first for overall flow, changelog rules, and the
non-empty-track guarantee.

## Track facts

- **Tag pattern**: `@probo/cookie-banner/v*`
- **Version source**: `packages/cookie-banner/package.json`
- **Changelog**: `packages/cookie-banner/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-npm-cookie-banner.yaml`
- **Path filter** (for log/scoping): `packages/cookie-banner`

## Important: build script bakes the version

`packages/cookie-banner/build.mjs` reads `version` from `package.json`
and exposes it to the bundle as the `__SDK_VERSION__` define. The SDK
uses this value at runtime (e.g. when calling the Probo REST API), so
the build **must** run after the version bump to make sure the new
version is what gets published. The release CI workflow does run the
build, but we still run it locally as part of the release commit so:

- compile errors are caught before tagging,
- any tracked side-effects (`package-lock.json`, etc.) are part of the
  same `Release @probo/cookie-banner/v<version>` commit.

## Steps

1. From a clean `main`, list commits since the last
   `@probo/cookie-banner` tag:

   ```shell
   git log $(git describe --tags --abbrev=0 --match='@probo/cookie-banner/v*')..HEAD --oneline \
     -- packages/cookie-banner
   ```

   If the list is empty (or contains only non-user-facing commits), do
   not release this track.

2. Decide the version bump (PATCH for fixes, MINOR for features).
3. Bump the version using npm so `package.json` and `package-lock.json`
   stay consistent:

   ```shell
   npm --workspace @probo/cookie-banner version <X.Y.Z> --no-git-tag-version
   ```

4. Run the workspace build so `__SDK_VERSION__` is rebuilt from the new
   `package.json` and any compile error surfaces before we tag:

   ```shell
   npm --workspace @probo/cookie-banner run build
   ```

   `dist/` is gitignored, so this step does not produce checked-in build
   artifacts — but it must succeed for the release to be valid.

5. Write the new entry in `packages/cookie-banner/CHANGELOG.md` following
   the rules in [README.md](./README.md#2-writing-a-changelog-entry).
6. Show the user the changelog diff and the new version. Wait for
   confirmation.
7. Stage the files modified by the version bump and changelog edit
   (typically `packages/cookie-banner/package.json`,
   `packages/cookie-banner/CHANGELOG.md`, and `package-lock.json`).
   Commit subject: `Release @probo/cookie-banner/v<version>`. No body.
8. Annotated tag:
   `git tag -a @probo/cookie-banner/v<version> -m "@probo/cookie-banner/v<version>"`.
9. Push: `git push origin main --follow-tags`.

CI workflow `release-npm-cookie-banner.yaml` verifies the tag matches
`package.json`, runs the build again, publishes to npm with provenance +
SBOM, and creates a GitHub Release.
