# Changelog

All notable changes to the `proboctl` CLI will be documented in this file.

## Unreleased

## [0.8.0] - 2026-06-24

### Changed

- Enriched common third-party catalog entries with DPA, ToS, SLA, status/security/trust pages, subprocessor lists, and certifications from production data

## [0.7.1] - 2026-06-23

### Fixed

- Bump `golang.org/x/image` to v0.43.0, remediating CVE-2026-33813 (denial of service via malformed WEBP parsing) and CVE-2026-46602 (missing tile-size limit in `x/image/tiff`)

## [0.7.0] - 2026-06-19

### Changed

- `common-tracker-pattern mark-first-party` now also blanks the stale description on the catalog row and its uncategorised org tracker patterns, so a terminal first-party row no longer keeps a description naming a cleared vendor
- Unlinking a vendor now has full parity with first-party cleanup: it clears the stale description on both the catalog row and its uncategorised org tracker patterns and remaps those patterns to drop the stale vendor and re-resolve (re-arming catalog enrichment)

## [0.6.0] - 2026-06-18

### Added

- `proboctl common-tracker-pattern upsert` — create or update a catalog pattern by its natural key (tracker type, pattern, max age), applying each field only when its flag is passed and honoring the FIRST_PARTY and description-preservation invariants; `--enrich` re-arms the row for the async enrichment worker
- `proboctl common-tracker-pattern mark-first-party` and an `--attribution` filter on `common-tracker-pattern list` to audit and remediate wrongly attributed rows

### Changed

- `common-tracker-pattern` upsert now normalizes the attribution verdict when a vendor is linked or unlinked without `--attribution`: UNDETERMINED is promoted to THIRD_PARTY on link and THIRD_PARTY is downgraded to UNDETERMINED on unlink

## [0.5.0] - 2026-06-18

### Added

- `--enrich` flag on `common-third-party upsert` triggers enrichment immediately after upsert
- Enrichment provenance (state, attempts, last-run status, error, per-field confidence) now visible in `common-third-party` and `common-tracker-pattern` list and show views; enrichment tracking unified with outcome-based status

## [0.4.0] - 2026-06-12

### Added

- `proboctl common-third-party reenrich` — re-arm the async enrichment worker for selected catalog rows (verbatim via `--id`/`--slug`, or across the catalog via `--category`/`--keyword`/`--state`/`--status`), gated by `--dry-run` and `--yes`
- `proboctl common-third-party stats` — summarize the catalog by enrichment state and last run status
- `--state`/`--status` filters and `STATE`/`STATUS` columns on `common-third-party list`; enrichment state, attempts, last run status, error, per-field provenance, and discovered domains on `common-third-party show`

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
