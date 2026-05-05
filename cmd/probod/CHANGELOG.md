# Changelog

All notable changes to `probod` (the server, including the bundled `@probo/console`, `@probo/trust`, and `@probo/ui` frontends) will be documented in this file.

## Unreleased

## [0.181.0] - 2026-05-05

### Added

- Add SCIM tools to MCP API
- Add SCIM commands to CLI
- Add cookie banner detection page for uncategorised patterns
- Add `last_detected_at` and `last_matched_at` tracking on cookie patterns
- Add `uncategorisedPatterns` GraphQL connection on `CookieBanner`

### Changed

- Accept CIDR ranges in proxy `trusted-proxies` configuration
- Rename `categories` to `consentCategories` on cookie banner API surfaces
- Move cookie management from separate Cookies tab into the Display page
- Filter uncategorised category from cookie banner config and version snapshots

## [0.180.0] - 2026-05-04

### Fixed

- Use natural sort for SOA document export rows

### Added

- Add risk publish to document system

## [0.179.1] - 2026-05-02

### Fixed

- Fix n8n cookieConsentRecord getAll operation

## [0.179.0] - 2026-05-02

### Added

- Add cookie banner operations to n8n node
- Add `excluded` flag to cookie patterns (GraphQL/MCP/CLI/n8n) with source badge in category table
- Validate cookie policy link in banner description

### Changed

- Skip draft cookie banner version for uncategorised-only merges
- Exclude uncategorised category from consent contract
- Run cookie detection regardless of banner state
- Stop bumping cookie banner version on no-op updates
- Exclude translations from cookie banner version snapshots
- Allow clearing optional fields in n8n cookie updates
- Bump `@probo/cookie-banner` to 0.2.0

### Fixed

- Clear pending cookie-consent queue before stopping on 404

## [0.178.0] - 2026-05-01

### Added

- Add MCP tools for cookie banner, category, pattern, version, and consent records
- Add CLI commands for cookie banner, category, pattern, and consent records

### Fixed

- Fix auditor access to processing activities
- Fix contract end date field cut off in Add Person dialog

## [0.177.1] - 2026-04-30

### Fixed

- Reveal cookie banner sidebar entry in IAM organizations
- Render cookie-consent placeholders when no prior consent exists
- Fix cookie-consent placeholder sizing for absolutely or sticky positioned elements
- Allow OIDC and magic-link sessions to assume password-only organizations

## [0.177.0] - 2026-04-30

### Added

- Add cookie patterns to group detected cookies by URL prefix, with auto-detection worker and console management
- Add `DurationInput` component to `@probo/ui`

### Changed

- Refactor cookie banner forms to react-hook-form
- Store cookie durations as `max_age_seconds`
- Update `@probo/cookie-banner` public exports and bump to 0.1.0

### Fixed

- Filter browser-extension cookies from detection

## [0.176.1] - 2026-04-29

### Fixed

- Fix empty text nodes in generated documents

## [0.176.0] - 2026-04-29

### Added

- Add vendor publish to document system, replacing snapshot mode

## [0.175.0] - 2026-04-29

### Added

- Add processing activity, DPIA and TIA publish to document system, replacing snapshot mode

### Changed

- Introspect OAuth2 refresh tokens per RFC 7662, honoring `token_type_hint`
- Invalidate other sessions on password change and all sessions on password reset
- Use forwarded headers for SCIM event client IP when running behind a load balancer
- Extract client IP from rightmost entry of `X-Forwarded-For` and `Forwarded` headers
- Update avatar initials colors

## [0.174.0] - 2026-04-28

### Added

- Add agent run supervisor with checkpoint persistence and resume across restarts
- Add finding and obligation publish to document system, replacing snapshot mode
- Add `--state` and `--contract-ended` filters to CLI/MCP/GraphQL user list
- Add Notion workspace name resolver for access review
- Add `X-SDK-Version` header to cookie banner SDK requests

### Changed

- Rename `excludeContractEnded` to `contractEnded` (two-way) across MCP, GraphQL, CLI, frontend
- Remove auditor's ability to publish SoA
- Request Google customer directory scope for access-review name sync

### Fixed

- Fix copy-paste in rich editor
- Fix long cookie name display and label colors in cookie banner
- Fix suspension checkpoint fallback in nested and parallel agent execution

## [0.173.0] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
