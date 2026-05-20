# Changelog

All notable changes to the `proboctl` CLI will be documented in this file.

## Unreleased

## [0.1.0] - 2026-05-20

### Added

- Initial release of `proboctl`, a Cobra-based CLI for Probo instance management that connects directly to PostgreSQL
- `proboctl seed common-third-parties` — import the bundled third-party catalog (formerly the standalone `common-third-parties-import` command); `data.json` is embedded in the binary
- `proboctl seed common-tracker-patterns` — import bundled tracker patterns (formerly the standalone `common-tracker-patterns-import` command)
