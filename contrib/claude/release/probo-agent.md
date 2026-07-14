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
--enrollment-token …` from the shell, or the curl-to-sh installer
documented below. Windows release binaries are
cross-compiled from Linux with MinGW (CGO). macOS release binaries and
`.pkg` installers must be built on macOS with `CGO_ENABLED=1`.

macOS `.pkg` installers are built locally with
`cmd/probo-agent/installer/macos/build.sh` (requires macOS, a
pre-built binary, and the Swift toolchain). The script also compiles
`Probo Agent.app` — the headless `probo://` URL handler installed to
`/Applications` — from `cmd/probo-agent/installer/macos/enroll-ui/`.
They are not part of the GitHub Release workflow yet.

Windows enrollment is browser-driven: the console issues a
`probo://enroll?server=...&token=...` deep link handled by
`probo-agent enroll-url`. After install, register the protocol for the
current user with
`cmd/probo-agent/installer/windows/register-protocol.ps1` (per-user
`HKCU` handler pointing at `probo-agent.exe`). The system tray helper
(`probo-agent tray`, CGO build on local Windows hosts) shows enrollment
status; enrollment itself happens in the browser.

Region labels and console URLs for the macOS installer HTML live in
`cmd/probo-agent/installer/regions.json`. A Go test keeps US/EU URLs in
sync with `pkg/deviceagent/server_url.go`.

## Install script

`cmd/probo-agent/installer/install.sh` is published as `install.sh` on each
`probo-agent/v*` GitHub Release. It supports Darwin, Linux, and FreeBSD.
Each published `install.sh` pins one release: the release workflow injects
the tag and archive SHA-256 checksums into the script before upload.

```shell
# Interactive — curl install.sh from the target release
curl -fsSL "https://github.com/getprobo/probo/releases/download/probo-agent/vX.Y.Z/install.sh" | sudo sh

# Unattended / MDM
curl -fsSL "…/install.sh" | sudo \
  PROBO_SERVER_URL=https://us.probo.com \
  PROBO_ENROLLMENT_TOKEN='…' sh

# Mirror release assets (PROBO_AGENT_RELEASE_BASE must end with the embedded tag)
PROBO_AGENT_RELEASE_BASE="https://mirror.example/probo-agent/vX.Y.Z" \
  curl -fsSL "…/install.sh" | sudo sh
```

The script downloads only the matching platform archive and verifies its
SHA-256 against checksums embedded in `install.sh` at release time. This
anchors trust to the script the user already obtained, rather than a
co-downloaded `checksums.txt`. Post-install upgrades still use cosign
bundle verification via the agent auto-update path.

The script installs the binary to `/usr/local/bin/probo-agent`, then runs
`probo-agent install` to enroll and start the OS service. Agent state defaults
to `/var/lib/probo-agent` (override with `--dir` or `PROBO_AGENT_STATE_DIR`).

Environment variables:

| Variable | Purpose |
|----------|---------|
| `PROBO_AGENT_RELEASE_BASE` | Override release download base URL (must match embedded tag) |
| `PROBO_AGENT_RELEASE_TAG` | Override embedded release tag (local dev) |
| `PROBO_AGENT_SKIP_CHECKSUM_VERIFY` | Skip SHA-256 verification (local dev only) |
| `PROBO_AGENT_STATE_DIR` | Agent state directory (`--dir`; default `/var/lib/probo-agent`) |
| `PROBO_SERVER_URL` | Probo server base URL (skip interactive prompt) |
| `PROBO_ENROLLMENT_TOKEN` | One-shot enrollment token (skip interactive prompt) |
| `PROBO_NO_AUTO_UPDATE` | Set to `true` to pass `--no-auto-update` |

Never pass the enrollment token in the curl URL.

## Local dev install

Build, package, and install a dev binary with `sudo`. Agent state is stored
under `~/.local/share/probo-agent-dev` (`--dir`); release archives are staged
under `~/.cache/probo-agent-dev` so the two trees do not overlap.

```shell
make -C cmd/probo-agent install \
  PROBO_SERVER_URL=https://us.probo.com \
  PROBO_ENROLLMENT_TOKEN='…'
```

This does not compile Go in `cmd/probo-agent`; it reuses `bin/probo-agent` from
the root `GNUmakefile` (via `make bin/probo-agent`, invoked automatically when
needed), stages it into a local release archive, then runs `installer/install.sh`
against that `file://` archive with `PROBO_AGENT_RELEASE_TAG=probo-agent/dev`,
skips checksum verification, installs the binary to `/usr/local/bin/probo-agent`,
and passes `--skip-service` and `--dir ~/.local/share/probo-agent-dev` by default
(`INSTALL_ARGS` overrides).

Run the agent in the foreground against the same dev state directory:

```shell
make -C cmd/probo-agent run
```

Remove dev artifacts:

```shell
make -C cmd/probo-agent clean
```
