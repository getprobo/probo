# Release `probod-bootstrap`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `probod-bootstrap/v*`
- **Version source**: `cmd/probod-bootstrap/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/probod-bootstrap/VERSION` directly
- **Changelog**: `cmd/probod-bootstrap/CHANGELOG.md`
- **Files to stage**: `cmd/probod-bootstrap/VERSION`,
  `cmd/probod-bootstrap/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-probod-bootstrap.yaml`
- **Path filter**: `cmd/probod-bootstrap`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='probod-bootstrap/v*')..HEAD --oneline \
  -- cmd/probod-bootstrap
```

If empty or non-user-facing only, do not release this track.

## Notes

CI builds binaries for 9 OS/arch targets and publishes a GitHub Release.
The same binary, built from the tagged ref, is also bundled into the
probod Docker image when `probod/v*` runs.
