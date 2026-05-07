# Release `prb` (CLI)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `prb/v*`
- **Version source**: `cmd/prb/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/prb/VERSION` directly
- **Changelog**: `cmd/prb/CHANGELOG.md`
- **Files to stage**: `cmd/prb/VERSION`, `cmd/prb/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-prb.yaml`
- **Path filter**: `cmd/prb pkg/cli pkg/cmd`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='prb/v*')..HEAD --oneline \
  -- cmd/prb pkg/cli pkg/cmd
```

If empty or non-user-facing only, do not release this track.

## Notes

CI builds binaries for 9 OS/arch targets, publishes a GitHub Release,
and updates the Homebrew formula at `getprobo/homebrew-tap`.
