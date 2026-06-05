# Changelog

All notable changes to the `@probo/cookie-banner` SDK will be documented in this file.

## Unreleased

## [0.9.0] - 2026-06-05

### Added

- Support Indonesian, Italian, Japanese, Korean, Polish, Portuguese, Turkish, Ukrainian, and Chinese

## [0.8.0] - 2026-06-03

### Added

- Expose tracker type on `CookieItem` and render it in the headless cookie list and themed banner

## [0.7.0] - 2026-05-27

### Added

- Re-export public domain types (`BannerConfig`, `Category`, `Regulation`, `ConsentAction`, `ConsentRecord`, `CookieItem`, `VisitorConsent`) from the package entry point

## [0.6.0] - 2026-05-26

### Added

- Eagerly bootstrap GCM (deny all consent types before config fetch) to close the gap where gtag could track during async config loading
- Track source (`script`/`pre-existing`) on detected storage trackers (localStorage, sessionStorage, indexedDB, cacheStorage)
- Mark page-world extension writes with a new `EXTENSION` source

### Changed

- Overhaul extension-activity detection: add synchronous DOM hooks (IDL setters, `setAttribute`, HTML-parsing entry points, fetch/XHR/sendBeacon) and drop ineffective `isExtensionCaller()` wraps

### Removed

- Remove the PostHog integration (use `data-cookie-consent` script blocking instead)

### Fixed

- Disconnect the previous `MutationObserver` in `load()`'s error path so failing loads no longer leak observers

## [0.5.0] - 2026-05-20

### Added

- Expose a programmatic consent API via a `ConsentManager` singleton (`@probo/cookie-banner/consent` ESM entrypoint and `window.Probo.consent` on the IIFE), so customers can read and react to consent state from their own bundled JavaScript — unblocking third-party SDKs that initialize programmatically and cannot be gated via `data-cookie-consent` attributes
- `ConsentManager` exposes a `subscribe()` method (unifying `onReady` + `onChange` with immediate replay) and `getAll()` now returns a cached, referentially stable snapshot, making it safe to use with React's `useSyncExternalStore`
- Share the `ConsentManager` singleton across bundles on `globalThis` so multiple copies of the SDK on the same page see the same state

## [0.4.1] - 2026-05-13

### Changed

- Skip the `/consents/:id` fetch for first-time visitors (no `visitorId` in localStorage); the visitor ID is now created lazily on the first consent action instead of triggering a guaranteed-404 request on every initial page load

## [0.4.0] - 2026-05-12

### Added

- `ResourceDetector` (renamed from `ThirdPartyDetector`) now picks up everything the browser loads via a single `PerformanceObserver`: tracking pixels, cross-origin stylesheets and web fonts, `fetch` / XHR / `sendBeacon` / `ping` calls, and `<video>` / `<audio>` / `<embed>` / `<object>` media — closing a real gap with headless cookie scanners that previously missed beacons fired after script teardown
- Detect registered service workers (reported as a `SERVICE_WORKER` tracker resource) and Cache Storage buckets (reported as a `CACHE_STORAGE` tracker)
- Detect HTTP-header cookies on Chromium browsers via the `CookieStore` change event (new `http` cookie source)
- Capture the script initiator URL on detected cookies and storage writes by walking the synchronous call stack, enabling per-vendor attribution

### Changed

- Collapse the three detectors (cookies, storage, resources) onto a single shared `ReportQueue`: one debounced POST instead of up to three, type-namespaced dedup keys (`c:` / `s:` / `r:`) that cannot collide across detectors, and a tab-close drain via `sendBeacon` (with keepalive-fetch fallback) so the last debounce window of detections is no longer lost on unload
- Report full origin+pathname for detected scripts and iframes (query string stripped) so resources served from the same domain but different paths can be distinguished
- Rename `ThirdPartyDetector` to `ResourceDetector` to match what it actually emits (breaking for SDK consumers importing it by name)

### Fixed

- Always attempt `sendBeacon` on flush so pending reports are no longer dropped during page unload when an async flush is in flight
- Keep pending entries queued until the transport confirms delivery, so transient network errors and bfcache restores no longer silently drop detection reports

## [0.3.1] - 2026-05-08

### Fixed

- Reopen the banner instead of the preference panel for OPT_OUT regulations (e.g. CCPA) when clicking the settings widget, since users only need Accept/Reject choices
- `ProboRejectButton` and `ProboCustomizeButton` now auto-hide when their corresponding text key is empty in the server config, so headless SDK consumers no longer need regulation-aware layout logic

## [0.3.0] - 2026-05-07

### Added

- Expose detected privacy regulation (GDPR, CCPA, etc.) on `BannerConfig`, via `CookieBannerClient` getter, and in the `probo-ready` event detail so themed-banner consumers can adapt their UI per regulation
- Adapt banner texts and button visibility per regulation (opt-out notice for CCPA, simple notice when no regulation applies); buttons whose text is empty are now hidden

### Fixed

- Defer banner button validation until config is loaded so required-button checks reflect the active consent mode

## [0.0.0] - 2026-04-27

### Added

- Initial scaffold of the cookie banner SDK with web components, headless and themed entrypoints, settings link element, Google Consent Mode v2 integration, PostHog consent plugin, Global Privacy Control (GPC) support, internationalization with default translations for English, French, German, and Spanish, and graceful config fetch failure handling.
