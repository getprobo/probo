# Release `probod` (server group)

This track ships `probod`, `@probo/console`, `@probo/trust`, and
`@probo/ui` together as the Docker image and accompanying binary archive.
They share the same version.

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `probod/v*`
- **Version source**: `cmd/probod/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/probod/VERSION` directly
- **Changelog**: `cmd/probod/CHANGELOG.md` (covers all four components)
- **Files to stage**: `cmd/probod/VERSION`, `cmd/probod/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-probod.yaml`
- **Path filter**: `cmd/probod apps/console apps/trust packages/ui pkg`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='probod/v*')..HEAD --oneline \
  -- cmd/probod apps/console apps/trust packages/ui pkg
```

If empty or non-user-facing only, do not release this track.

## Notes

The changelog covers changes across all four components (`probod`,
`@probo/console`, `@probo/trust`, `@probo/ui`).

CI builds the frontends and Go binaries, builds and pushes the
multi-arch image to `ghcr.io/getprobo/probo:probod-v<version>` (and
`:latest`), runs Trivy + cosign + attestations, and publishes the GitHub
Release.
