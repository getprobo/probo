# Release `proboctl`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `proboctl/v*`
- **Version source**: `cmd/proboctl/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/proboctl/VERSION` directly
- **Changelog**: `cmd/proboctl/CHANGELOG.md`
- **Files to stage**: `cmd/proboctl/VERSION`, `cmd/proboctl/CHANGELOG.md`
- **Path filter**: `cmd/proboctl pkg/proboctl`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='proboctl/v*')..HEAD --oneline \
  -- cmd/proboctl pkg/proboctl
```

If empty or non-user-facing only, do not release this track.
