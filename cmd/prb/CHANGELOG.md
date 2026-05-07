# Changelog

All notable changes to the `prb` CLI will be documented in this file.

## Unreleased

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
