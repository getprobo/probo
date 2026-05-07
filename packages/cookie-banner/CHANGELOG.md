# Changelog

All notable changes to the `@probo/cookie-banner` SDK will be documented in this file.

## Unreleased

## [0.3.0] - 2026-05-07

### Added

- Expose detected privacy regulation (GDPR, CCPA, etc.) on `BannerConfig`, via `CookieBannerClient` getter, and in the `probo-ready` event detail so themed-banner consumers can adapt their UI per regulation
- Adapt banner texts and button visibility per regulation (opt-out notice for CCPA, simple notice when no regulation applies); buttons whose text is empty are now hidden

### Fixed

- Defer banner button validation until config is loaded so required-button checks reflect the active consent mode

## [0.0.0] - 2026-04-27

### Added

- Initial scaffold of the cookie banner SDK with web components, headless and themed entrypoints, settings link element, Google Consent Mode v2 integration, PostHog consent plugin, Global Privacy Control (GPC) support, internationalization with default translations for English, French, German, and Spanish, and graceful config fetch failure handling.
