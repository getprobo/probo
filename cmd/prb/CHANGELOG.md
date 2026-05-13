# Changelog

All notable changes to the `prb` CLI will be documented in this file.

## Unreleased

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
