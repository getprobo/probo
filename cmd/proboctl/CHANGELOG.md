# Changelog

All notable changes to the `proboctl` CLI will be documented in this file.

## Unreleased

## [0.3.1] - 2026-06-11

### Changed

- References updated to probo.com

## [0.3.0] - 2026-06-10

### Added

- `proboctl common-third-party upsert` — create or update a vendor in the global catalog (partial-merge, slug-keyed)
- `proboctl common-tracker-pattern link`/`unlink` — repoint or detach a catalog pattern's third party (link re-arms enrichment and remaps uncategorised org trackers)
- `proboctl common-tracker-pattern set-description` — write a description, mark the row enriched, and backfill linked org patterns
- `proboctl cookie-banner reset-trackers --keyword <substring>` — scope both glob decomposition and mapping reset to patterns whose pattern or display name contains the substring

### Changed

- `proboctl common-tracker-pattern reenrich` no longer requires a selection anchor; filter flags (e.g. `--without-description`) now select across the whole catalog when no anchor is supplied

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
