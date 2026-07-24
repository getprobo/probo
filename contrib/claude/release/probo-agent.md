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

On macOS hosts, `make` enables CGO so the menu bar enrollment helper
is included. Windows tray support is pure Go (`CGO_ENABLED=0`). Linux
and FreeBSD builds stay pure Go (no tray).

## Notes

CI builds binaries for linux, windows, and freebsd (amd64/arm64) on
Linux runners, and builds **CGO-enabled** darwin archives plus a
signed/notarized fat `.pkg` on a macOS runner. The GitHub
Release includes those archives, `probo-agent_*_darwin.pkg`,
`install.sh`, signed checksums, SBOM, and build attestations. The agent
auto-update path downloads the matching archive plus `checksums.txt` and
verifies the cosign bundle before installing.

The menu bar / tray enrollment flow is **macOS and Windows only**.
Linux and FreeBSD use `probo-agent install --server …
--enrollment-token …` from the shell, or the curl-to-sh installer
documented below. Windows release binaries are cross-compiled from
Linux with `CGO_ENABLED=0` (tray is pure Go). macOS release binaries
and the `.pkg` are built on macOS with `CGO_ENABLED=1`.

### macOS `.pkg` (MDM / GUI install)

Release and local builds use
`cmd/probo-agent/installer/macos/build.sh` (requires macOS, a
pre-built fat binary via `lipo`, and the Swift toolchain).
**Signing is mandatory:** `CODESIGN_IDENTITY` and `APPLE_TEAM_ID`
must be set. The script compiles `Probo Agent.app` (the headless
`probo://` URL handler + privileged helper) from
`cmd/probo-agent/installer/macos/enroll-ui/`, signs nested Mach-Os
then the app bundle, optionally signs the product with
`INSTALLER_IDENTITY`, and notarizes/staples when `APPLE_ID` and
`APPLE_ID_PASSWORD` are set (password is stored into a keychain
profile; submits use `--keychain-profile` so the secret is not on
`notarytool submit` argv).

The Finder/Dock icon for `Probo Agent.app` comes from a single master
PNG, `cmd/probo-agent/installer/macos/enroll-ui/Resources/icon-original.png`
(Probo square mark). At PKG build time `build.sh` pads/resizes it with
`sips`, compiles `AppIcon.icns` with `iconutil` under the build stage
directory, and installs it into `Contents/Resources/`. `Info.plist`
sets `CFBundleIconFile` to `AppIcon`. Do not commit generated
`.icns` / iconset PNGs.

There is no unsigned PKG path and no osascript elevation fallback.
Local testing of browser enrollment requires a Developer ID–signed
build. CLI enrollment without the app uses `sudo probo-agent install`.

```shell
# Local signed pkg (example)
export CODESIGN_IDENTITY="Developer ID Application: Probo Inc (TEAMID)"
export INSTALLER_IDENTITY="Developer ID Installer: Probo Inc (TEAMID)"
export APPLE_TEAM_ID="TEAMID"

GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o dist/probo-agent_arm64 ./cmd/probo-agent
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o dist/probo-agent_amd64 ./cmd/probo-agent
lipo -create dist/probo-agent_arm64 dist/probo-agent_amd64 -output dist/probo-agent_universal
cmd/probo-agent/installer/macos/build.sh \
  --binary dist/probo-agent_universal \
  --version "$(cat cmd/probo-agent/VERSION)"
```

PKG postinstall always installs the global tray LaunchAgent, registers
`probo://`, and installs the privileged helper
(`com.probo.agent.helper`) under `/Library/PrivilegedHelperTools` as
root. The only admin authentication is the normal macOS Installer
prompt for the PKG itself. The LaunchDaemon for `probo-agent run` is
created only after enrollment (`probo-agent install`, deep link, or MDM
`/tmp/probo-agent.conf`).

Browser enrollment uses `Probo Agent.app` over XPC to the
PKG-installed helper — no SMJobBless and no admin prompt on the enroll
path. A missing or dead helper surfaces an error asking to reinstall
the package.

Manual QA checklist (macOS PKG):

1. Fresh signed PKG: Installer may ask for admin once; deep link enrolls with no further prompt.
2. Repeat deep link on an enrolled device: no prompt, immediate success.
3. After `sudo make -C cmd/probo-agent uninstall`, deep link fails with a clear “reinstall package” error (no osascript).
4. `sudo make -C cmd/probo-agent uninstall` removes daemon, tray, helper, app, and state.
5. MDM `/tmp/probo-agent.conf` postinstall still enrolls without a browser prompt.
6. Notarized PKG passes Gatekeeper; `codesign --verify --deep` succeeds on the app bundle.
7. Unsigned `build.sh` exits with an error requiring `CODESIGN_IDENTITY`.

### Apple signing secrets (GitHub)

The `build-macos` job in `release-probo-agent.yaml` expects the same
secret names as the auditor-mode release workflow. Configure these on
the probo GitHub repository (or org) before tagging a release:

| Secret | Purpose |
|--------|---------|
| `APPLE_CERTIFICATE` | Base64-encoded `.p12` (Developer ID) |
| `APPLE_CERTIFICATE_PASSWORD` | `.p12` password |
| `KEYCHAIN_PASSWORD` | Ephemeral CI keychain password |
| `CODESIGN_IDENTITY` | e.g. `Developer ID Application: Probo Inc (TEAMID)` |
| `INSTALLER_IDENTITY` | e.g. `Developer ID Installer: Probo Inc (TEAMID)` |
| `APPLE_ID` | Apple ID email for `notarytool store-credentials` |
| `APPLE_ID_PASSWORD` | App-specific password (stored into a keychain profile; not passed to `submit`) |
| `APPLE_TEAM_ID` | 10-character Team ID |

Local notarization uses the same env vars as CI:

```shell
export APPLE_ID="you@example.com"
export APPLE_ID_PASSWORD="app-specific-password"
# optional: NOTARYTOOL_KEYCHAIN_PROFILE=probo-agent-notary (default)
```

Windows enrollment is browser-driven: the console issues a
`probo://enroll?server=...&token=...` deep link handled by
`Probo Agent.app` on macOS (PKG-installed helper + XPC) or
`probo-agent enroll-url` on Windows. After install, register the protocol for the
current user with
`cmd/probo-agent/installer/windows/register-protocol.ps1` (per-user
`HKCU` handler pointing at `probo-agent.exe`). The system tray helper
(`probo-agent tray`) shows enrollment status; enrollment itself happens
in the browser.

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

## Local install (macOS PKG)

Primary loop for testing browser enrollment (app + privileged helper):

```shell
export CODESIGN_IDENTITY="Developer ID Application: … (TEAMID)"
export INSTALLER_IDENTITY="Developer ID Installer: … (TEAMID)"   # optional
export APPLE_TEAM_ID="TEAMID"

make -C cmd/probo-agent install     # uninstall leftovers → build PKG → installer
make -C cmd/probo-agent uninstall   # full system teardown (idempotent)
make -C cmd/probo-agent clean       # uninstall + wipe dist/ caches / enroll-ui/.build
```

`install` always runs `uninstall` first so previous helpers, apps, and
Launch Services registrations cannot shadow the new build. Signing env
vars are required for `pkg` / `install` on Darwin.

### CLI-only install (no app / helper)

Binary-only path via `installer/install.sh` (any Unix). State under
`~/.local/share/probo-agent-dev`; staging under `~/.cache/probo-agent-dev`.

```shell
make -C cmd/probo-agent install-cli \
  PROBO_SERVER_URL=https://us.probo.com \
  PROBO_ENROLLMENT_TOKEN='…'
make -C cmd/probo-agent run
```
