# Changelog

All notable changes to `probod-bootstrap` will be documented in this file.

## Unreleased

### Added

- AWS Systems Manager Parameter Store resolution: env values prefixed with `awsps://<parameter-name>` are fetched at startup using the standard AWS SDK credential chain and cached per run
- Secrets Manager prefix `awssm://<secret-id>` as an explicit alias alongside the existing `aws://<secret-id>` prefix

## [0.2.0] - 2026-06-24

### Breaking Changes

- **All bootstrap env vars are now prefixed with `PROBOD_`** (e.g. `AUTH_COOKIE_SECRET` → `PROBOD_AUTH_COOKIE_SECRET`). Deployments must rename every bootstrap env var before upgrading.

### Added

- AWS Secrets Manager resolution: env values prefixed with `aws://<secret-id>` are fetched at startup using the standard AWS SDK credential chain and cached per run

## [0.1.2] - 2026-06-11

### Changed

- Support email updated to hello@probo.com

## [0.1.1] - 2026-05-08

### Fixed

- Fix builder tracing address default port

### Security

- Upgrade go to 1.26.3

## [0.1.0] - 2026-04-27

### Added

- First independent release. Generates a `probod` configuration file from environment variables. Previously shipped without a version string and bundled with the monorepo `vX.Y.Z` tag.
