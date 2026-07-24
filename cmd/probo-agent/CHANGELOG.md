# Changelog

All notable changes to the `probo-agent` device posture agent will be
documented in this file.

## [0.3.0] - 2026-07-24

### Added

- Branded Finder icon for `Probo Agent.app` (generated from a master PNG
  via `sips`/`iconutil` at PKG build time).

### Fixed

- macOS auto-update posture check now reads all five Software Update
  preferences backing the System Settings toggles, resolving managed
  (MDM) values before system ones, so disabled downloads/installs are
  correctly reported as failing instead of PASS.

## [0.2.0] - 2026-07-24

### Added

- macOS privileged helper (`com.probo.agent.helper`) embedded in
  `Probo Agent.app` and installed by PKG postinstall for XPC-driven
  browser enrollment (no SMJobBless / admin prompt on enroll).
- Hidden `probo-agent enroll-url --preflight` JSON output for the URL handler.
- `make -C cmd/probo-agent install|uninstall|clean` for local macOS PKG
  test loops (install tears down leftovers first).

### Changed

- Browser enrollment via `Probo Agent.app` uses HelperClient + XPC only
  (osascript elevation and enroll-time SMJobBless removed).
- macOS PKG / app builds require `CODESIGN_IDENTITY` and `APPLE_TEAM_ID`.
- CLI `enroll-url` on macOS refuses elevation; use the signed app deeplink
  or `sudo probo-agent install`.
- macOS `probo-agent uninstall` requires root (`sudo`).
- PKG preinstall removes stale privileged helper files on upgrade;
  postinstall reinstalls the helper as root.
- macOS release now ships a single universal (arm64 + x86_64) fat
  `probo-agent_<version>_darwin.pkg` instead of separate per-arch packages.

### Fixed

- Windows elevated install/uninstall no longer reports success when the
  user cancels the UAC prompt.

## [0.1.1] - 2026-06-11

### Changed

- Support email updated to hello@probo.com

## [0.1.0] - 2026-05-26

### Added

- Initial release of the Probo device posture agent.
- `probo-agent install`, `uninstall`, `run`, `status`, `collect` CLI
  commands.
- Managed OS service installation for macOS (`launchd`), Linux
  (`systemd`), FreeBSD (`rc.d`), and Windows (Service Control Manager).
- v1 posture check set per OS: disk encryption, screen lock, firewall,
  time sync, OS version, auto update, password policy, remote login.
- Enrollment / heartbeat / posture reporting against the new
  `/api/agent/v1` Probo REST API.
- Auto-update: the agent periodically checks GitHub Releases for a
  newer `probo-agent/v*` tag and self-installs it. The running binary
  is swapped atomically and the OS service supervisor is asked to
  restart via a dedicated exit code (`75`).
- Cosign signature verification of every release before installation:
  `checksums.txt.bundle` is verified with `sigstore-go` against the
  Sigstore public-good trust root, pinned to the GitHub Actions OIDC
  identity for `release-probo-agent.yaml` on a tagged commit. Releases
  without a Sigstore bundle, with an invalid bundle, or signed by a
  different workflow are rejected without touching the running
  binary.
- `probo-agent update [--check]` command for manual one-shot upgrade.
- `probo-agent install --no-auto-update` flag to opt out of automatic
  upgrades; the flag is persisted in `config.json` as
  `updates_disabled`.
- `probo-agent status` now reports the configured update interval and
  whether auto-update is enabled.
- Screen lock detection for additional Linux desktop environments: KDE
  Plasma, i3, Sway, Hyprland, Xfce, MATE, Cinnamon, UKUI, and LightDM.

### Fixed

- macOS launchd service label corrected to `com.probo.agent`.
- FreeBSD check command failures are now handled before reading service
  status.
- macOS postinstall script no longer uses `eval` to parse configuration.
- Windows agent key file replacement is now performed atomically.
- Windows service uninstall is now idempotent.
- FreeBSD `rc.d` install validates executable and state directory paths
  to prevent shell injection.