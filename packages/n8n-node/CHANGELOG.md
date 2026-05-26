# Changelog

All notable changes to the `@probo/n8n-nodes-probo` package will be documented in this file.

## Unreleased

## [0.188.0] - 2026-05-26

### Added

- Add `measure linkThirdParty`/`unlinkThirdParty` operations
- Add `thirdParty linkThirdParty`/`unlinkThirdParty`/`listChildThirdParties` operations for self-referential relations

### Changed

- Allow initial minor publishing of documents

## [0.187.1] - 2026-05-25

### Fixed

- Fix signature count mismatch in `getAllSignatures` — add a `state` filter to `DocumentVersionSignatureFilter` so results match the console's signatures tab

## [0.187.0] - 2026-05-22

### Added

- Add a `riskAssessment` resource exposing the full risk assessment hierarchy — assessments, scopes, nodes, processes, threats, and scenarios — with CRUD operations, scenario-to-risk and scenario-to-threat link/unlink, and scope Mermaid chart retrieval

## [0.186.0] - 2026-05-15

### Changed

- Rename the `vendor` resource and its operations to `thirdParty` across all node actions (breaking)

## [0.185.0] - 2026-05-13

### Changed

- Drop the `consentMode` field from cookie banner create/update operations and remove `consent_mode` from cookie banner outputs — consent mode is now derived from the visitor's geolocation at consent time (breaking)

## [0.184.0] - 2026-05-12

### Changed

- Replace `PREFIX` with `GLOB` in tracker pattern match type options (breaking)
- Drop `displayName` from tracker pattern update operations — it is now derived from pattern + match type (breaking)

## [0.183.0] - 2026-05-07

### Added

- Add `regulation` and `countryCode` fields on cookie consent record operations

## [0.182.0] - 2026-05-06

### Changed

- Replace `publishMinor`, `publishMajor`, and `requestApproval` document operations with a unified `publish` accepting a `minor` flag and required `changelog` (breaking)
- Rename `cookiePattern` operations to `trackerPattern` with new `trackerType` field (breaking)

### Removed

- Remove legacy `cookiePattern` operations

## [0.0.1] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
