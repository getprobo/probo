# Changelog

All notable changes to the `proboctl` CLI will be documented in this file.

## Unreleased

## [0.2.0] - 2026-06-09

### Added

- `proboctl common-tracker-pattern` commands (`list`, `show`, `stats`, `reenrich`) for inspecting and re-running enrichment on the global common-tracker-pattern catalog; selection anchors via `--id`, `--linked-banner`, `--linked-org`, or `--common-third-party`, narrowed by `--tracker-type`/`--keyword`/`--state`/`--without-description`
- `proboctl common-third-party` commands (`list`, `show`) for inspecting the global common-third-party catalog
- `proboctl cookie-banner reset-trackers <banner-gid>` to rebuild a banner's uncategorised, non-excluded tracker patterns from `detected_trackers` and re-arm the analysis and mapping workers (`--mapping-only` skips the rebuild)
- Cursor-pagination flags on list commands (`--first`/`--after`, `--last`/`--before`), mirroring the GraphQL connection arguments, with cursors emitted in the output

### Changed

- `--first`/`--last` default to 50 when omitted; reject `--first` combined with `--before` (previously silently flipped to backward pagination)

## [0.1.0] - 2026-05-20

### Added

- Initial release of `proboctl`, a Cobra-based CLI for Probo instance management that connects directly to PostgreSQL
- `proboctl seed common-third-parties` — import the bundled third-party catalog (formerly the standalone `common-third-parties-import` command); `data.json` is embedded in the binary
- `proboctl seed common-tracker-patterns` — import bundled tracker patterns (formerly the standalone `common-tracker-patterns-import` command)
