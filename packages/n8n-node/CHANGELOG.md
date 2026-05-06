# Changelog

All notable changes to the `@probo/n8n-nodes-probo` package will be documented in this file.

## Unreleased

## [0.182.0] - 2026-05-06

### Changed

- Replace `publishMinor`, `publishMajor`, and `requestApproval` document operations with a unified `publish` accepting a `minor` flag and required `changelog` (breaking)
- Rename `cookiePattern` operations to `trackerPattern` with new `trackerType` field (breaking)

### Removed

- Remove legacy `cookiePattern` operations

## [0.0.1] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
