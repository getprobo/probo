# Release `probod-bootstrap`

Entrypoint for releasing `probod-bootstrap`. Read [README.md](./README.md)
first for overall flow, changelog rules, and the non-empty-track
guarantee.

## Track facts

- **Tag pattern**: `probod-bootstrap/v*`
- **Version source**: `cmd/probod-bootstrap/VERSION` (contains only `X.Y.Z`)
- **Changelog**: `cmd/probod-bootstrap/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-probod-bootstrap.yaml`
- **Path filter** (for log/scoping): `cmd/probod-bootstrap`

## Steps

1. From a clean `main`, list commits since the last `probod-bootstrap`
   tag:

   ```shell
   git log $(git describe --tags --abbrev=0 --match='probod-bootstrap/v*')..HEAD --oneline \
     -- cmd/probod-bootstrap
   ```

   If the list is empty (or contains only non-user-facing commits), do
   not release this track.

2. Decide the version bump (PATCH for fixes, MINOR for features).
3. Bump the version in `cmd/probod-bootstrap/VERSION`.
4. Write the new entry in `cmd/probod-bootstrap/CHANGELOG.md` following
   the rules in [README.md](./README.md#2-writing-a-changelog-entry).
5. Show the user the changelog diff and the new version. Wait for
   confirmation.
6. Stage only `cmd/probod-bootstrap/VERSION` and
   `cmd/probod-bootstrap/CHANGELOG.md`. Commit subject:
   `Release probod-bootstrap/v<version>`. No body.
7. Annotated tag:
   `git tag -a probod-bootstrap/v<version> -m "probod-bootstrap/v<version>"`.
8. Push: `git push origin main --follow-tags`.

CI workflow `release-probod-bootstrap.yaml` builds binaries for 9 OS/arch
targets and publishes a GitHub Release. Note: the same binary, built
from the tagged ref, is also bundled into the probod Docker image when
`probod/v*` runs.
