# Changelog

All notable changes to the `prb` CLI will be documented in this file.

## Unreleased

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
