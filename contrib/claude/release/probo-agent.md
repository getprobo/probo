# Release `probo-agent`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `probo-agent/v*`
- **Version source**: `cmd/probo-agent/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/probo-agent/VERSION` directly
- **Changelog**: `cmd/probo-agent/CHANGELOG.md`
- **Files to stage**: `cmd/probo-agent/VERSION`, `cmd/probo-agent/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-probo-agent.yaml`
- **Path filter**: `cmd/probo-agent pkg/deviceagent`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='probo-agent/v*')..HEAD --oneline \
  -- cmd/probo-agent pkg/deviceagent
```

If empty or non-user-facing only, do not release this track.

## Build

```shell
make bin/probo-agent
```

On macOS and Windows hosts, `make` enables CGO automatically so the
menu bar / tray enrollment helper is included. Linux and FreeBSD builds
stay pure Go (no tray).

## Notes

CI builds binaries for 8 OS/arch targets (linux, darwin, and windows on
amd64 and arm64; freebsd on amd64 and arm64), publishes a GitHub
Release with signed checksums, SBOM, and build attestations. The agent
auto-update path downloads the matching archive plus `checksums.txt` and
verifies the cosign bundle before installing.

The menu bar / tray enrollment flow is **macOS and Windows only**.
Linux and FreeBSD use `probo-agent install --server …
--enrollment-token …` from the shell. Windows release binaries are
cross-compiled from Linux with MinGW (CGO). macOS release binaries and
`.pkg` installers must be built on macOS with `CGO_ENABLED=1`.

macOS `.pkg` installers are built locally with
`cmd/probo-agent/installer/macos/build.sh` (requires macOS and a
pre-built binary). They are not part of the GitHub Release workflow yet.

Windows enrollment UI (`probo-agent-enroll-ui.exe`) is built with
`cmd/probo-agent/installer/windows/build-enroll-ui.sh` (requires the
.NET 8 SDK). CI builds it for Windows release archives; ship it next to
`probo-agent.exe` and `regions.json` under `C:\Program Files\Probo\`.

Region labels and console URLs for both enrollment UIs live in
`cmd/probo-agent/installer/regions.json`. A Go test keeps US/EU URLs in
sync with `pkg/deviceagent/server_url.go`.
