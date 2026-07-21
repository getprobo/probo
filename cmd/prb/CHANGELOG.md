# Changelog

All notable changes to the `prb` CLI will be documented in this file.

## Unreleased

## [0.202.0] - 2026-07-21

### Added

- `prb compliance-portal update` gained `--title`, `--description`, `--website-url`, `--email`, and `--headquarter-address` flags to manage the compliance page profile
- `prb webhook` event choices now include `RIGHT_REQUEST_CREATED`, `RIGHT_REQUEST_UPDATED`, and `RIGHT_REQUEST_DELETED`

### Changed

- Renamed the `prb trust-center` command tree to `prb compliance-portal` to match the product and GraphQL rename

## [0.201.0] - 2026-07-20

### Added

- `prb trust-center commitment` and `prb trust-center commitment-group` commands to create, list, update, and delete compliance portal commitments and their groups

## [0.200.0] - 2026-07-20

### Changed

- Renamed the risk overview's "Inherent" label to "Initial" in command help text, for consistency with the console

## [0.199.0] - 2026-07-02

### Added

- `prb webhook` event choices now cover the full document lifecycle: `document.created`/`updated`/`archived`/`unarchived`/`deleted`, the `document.version.*` events, and their `signature.*` and `approval.*` sub-events

### Changed

- `prb auth login` defaults and region prompts now use `eu.probo.com` and `us.probo.com` instead of `*.console.getprobo.com`

### Removed

- `MEETING_*` webhook event choices removed from `prb webhook` create/update; they were never valid backend event types and only produced API rejections

## [0.198.0] - 2026-06-30

### Removed

- Access review campaigns no longer expose a framework-controls field

## [0.197.1] - 2026-06-25

### Fixed

- `prb login` now requests the correct OAuth2 scopes (removed redundant `:read` suffix variants)

## [0.197.0] - 2026-06-22

### Added

- `prb resource-alias` commands to set and remove aliases on trust center entries

## [0.196.0] - 2026-06-19

### Added

- `prb login` now requests all v1 API scopes so device tokens work under OAuth2 scope enforcement

## [0.195.0] - 2026-06-16

### Changed

- `prb access-review` campaign sources are now first-class: each campaign captures a source snapshot (name, connector) at start time so a review stays coherent when the underlying source is edited or deleted, and fetch attempts are tracked as an append-only log instead of a single denormalized status field

## [0.194.0] - 2026-06-11

### Added

- `active` status field on access entries exposed in `prb access-entry` commands

### Changed

- References updated to probo.com

## [0.193.0] - 2026-06-10

### Added

- Expose `regulation_source` (`detected`/`default`) on `prb consent-record list`/`view` to show whether the regulation was resolved from geolocation or fell back to GDPR

### Changed

- `prb third-party list --first-level` replaced by `--level <N>` (1 = direct, 2+ = indirect) to support arbitrary nesting depth

### Removed

- `prb third-party link`/`unlink` commands; sub-third-parties are now scoped by a `parent_third_party_id` on the third-party itself

## [0.192.0] - 2026-06-09

### Added

- Add `prb risk-assessment boundary` command group (`create`, `list`, `view`, `update`, `delete`) and `--boundary-id` flag on `node create`/`update` to group risk assessment nodes within a scope
- Add `prb cookie-banner regenerate-policy` command to re-trigger tracker policy generation for a banner that already has a published version
- Expose `common_tracker_pattern_id` on `prb tracker-pattern list`/`view` to show whether a pattern is linked to the common tracker catalog

## [0.191.0] - 2026-06-02

### Changed

- Replace `prb third-party assess` with `prb third-party vet` to enqueue async third-party vetting; the command now returns immediately after queuing the job instead of waiting for the report

## [0.190.0] - 2026-05-27

### Added

- Add `prb user archive` command to deactivate a user profile while keeping them in the organization

## [0.189.0] - 2026-05-26

### Added

- Add `prb third-party link`/`unlink` commands for self-referential third-party relations
- Add `prb measure link-third-party`/`unlink-third-party` commands

### Changed

- Allow initial minor publishing of documents

## [0.188.0] - 2026-05-22

### Added

- Add `prb risk-assessment` command group with nested `scope`, `node`, `process`, `threat`, and `scenario` subcommands for managing the hierarchical risk assessment system, including scenario-to-risk and scenario-to-threat link/unlink and Mermaid chart retrieval

## [0.187.0] - 2026-05-15

### Changed

- Rename `prb vendor*` command group to `prb third-party*` (breaking)

## [0.186.0] - 2026-05-13

### Changed

- Drop `--consent-mode` flag from `prb cookie-banner create`/`update` and remove the `consent_mode` column from `cookie-banner` outputs — consent mode is now derived from the visitor's geolocation at consent time (breaking)

## [0.185.0] - 2026-05-12

### Changed

- Update kit package

## [0.184.0] - 2026-05-12

### Added

- Add `prb tracker-resource` command group (`list`, `view`, `create`, `update`, `delete`, `move`) for managing detected scripts, iframes, and other tracker resources

### Changed

- Replace `PREFIX` match type with `GLOB` in `prb tracker-pattern` interactive prompts (breaking)
- Drop `--display-name` from `prb tracker-pattern update` — display names are now derived from pattern + match type (breaking)

## [0.183.1] - 2026-05-08

### Security

- Upgrade go to 1.26.3

## [0.183.0] - 2026-05-07

### Added

- Add `regulation` and `country code` fields on cookie consent records, plus the `STATEMENT_OF_APPLICABILITY` document type on `prb document update`

### Fixed

- Allow editing metadata (title, document type, classification) on generated document versions

## [0.182.0] - 2026-05-06

### Added

- Add `--minor` flag to generated-document publish commands

### Changed

- Replace `prb document publish-major` and `publish-minor` with `prb document publish [--minor]`
- Rename `prb cookie-pattern` command group to `prb tracker-pattern`

## [0.173.0] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
