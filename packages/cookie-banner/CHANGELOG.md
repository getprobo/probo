# Changelog

All notable changes to the `@probo/cookie-banner` SDK will be documented in this file.

## Unreleased

### Changed

- API calls to the retired hosted console hosts (`eu.console.getprobo.com`, `us.console.getprobo.com`) are rewritten to `eu.probo.com` and `us.probo.com` so existing snippets keep working after the domain migration. Self-hosted and already-migrated base URLs are unchanged.

## [0.16.0] - 2026-08-19

### Added

- `google-consent-mode="off"` (and `data-google-consent-mode="off"` on the script tag) disables the Google Consent Mode integration entirely, for sites that already manage Consent Mode through their own CMP or GTM setup

### Fixed

- The Beacon-based unload report used a content type that is not CORS-safelisted, so browsers could silently drop it instead of sending it

## [0.15.0] - 2026-08-13

### Added

- Operators can disable resource detection per banner: the SDK skips `ResourceDetector` and the server ignores resources on the shared report endpoint, while cookies and storage keep reporting normally

## [0.14.0] - 2026-08-11

### Changed

- Privacy Choices wording keys are regulation-scoped: `privacy_choices_*` → `privacy_choices_*_ccpa`, falling back to the unsuffixed key for configs that have not migrated. The compact opt-out CTA stays on `button_opt_out` (generic); California receives the statutory label from the server via `button_opt_out_ccpa`

## [0.13.0] - 2026-08-05

### Changed

- `Regulation` is a `string` alias rather than an exhaustive union, so the SDK no longer mirrors every backend regulation identifier. Banner behavior is driven by the server-provided `layout`, and new jurisdictions (US state laws, Canadian provincial PIPA) no longer require an SDK release

## [0.12.0] - 2026-08-05

### Added

- CCPA Privacy Choices panel (`privacy_choices` state + `<probo-privacy-choices>`): under CCPA, the settings link opens a dedicated surface that describes the right to opt out of sale/sharing and the right to limit sensitive personal information, with a Do Not Sell or Share action (11 CCR § 7015). SPI is explained as a statement in v1 (no toggle)
- `<probo-settings-link>` always shows the statutory “Your Privacy Choices” label plus the official CPPA opt-out icon under CCPA (overriding integrator children); otherwise keeps children or falls back to `aria_cookie_settings`; optional GPC badge when the opt-out preference signal was applied
- Soft-validate that `<probo-settings-link>` is present (warn + `probo-validation`), since it is the sole reopen control
- Structured `layout` object in the banner config (`presentation`, `initial_state`, `reopen_state`, `default_non_necessary_granted`, `buttons`, `settings_link`). The banner now renders from these explicit fields instead of inferring behavior from hardcoded regulation checks and empty text keys. A config without `layout` means a self-hosted Probo backend older than probod v0.246.0 (the first release that sends it): the client logs an error asking to update probod and falls back to a strict opt-in layout
- First-class notice-only presentation (single acknowledge button) for jurisdictions with no cookie-consent law and implied-consent regimes (Japan APPI, Mexico LFPDPPP). Dismissing it records a distinct `ACKNOWLEDGE` consent action (via `<probo-acknowledge-button>` / `client.acknowledge()`) rather than `ACCEPT_ALL`, so the audit trail reflects an acknowledgement
- The themed `<probo-cookie-banner>` is now a thin dispatcher that mounts a focused renderer per presentation (opt-in / opt-out / notice); each renderer owns only its own markup, wording keys, and quirks, replacing the single component's runtime button-toggling
- Export `resolveLayout` and `resolveBannerText` (plus the `BannerLayout` / `BannerText` / `Presentation` types) so headless integrators can resolve the active presentation and its wording

### Changed

- Under CCPA, `layout.reopen_state` is `privacy_choices` (not the compact opt-out banner). Other OPT_OUT regimes (PIPEDA, Canadian PIPA, LGPD) still reopen the compact banner
- Under CCPA, the banner starts closed by default (cookies already follow the opt-out model until the visitor opts out)
- Auto-filled settings-link text inherits font and color from the host; the CCPA icon scales with `1em`
- Initial state, reopen target, default toggle state, button visibility, and the statutory settings-link style are now driven by the server-provided `layout` rather than `regulation === "CCPA"` / `consent_mode` branching
- Layout-aware clients (0.12+) receive the raw text keys and select their own wording from `layout.presentation`; the server-side key remap is now a shim for pre-0.12 clients only. Headless integrators on 0.12+ should use `resolveBannerText` / `layout` to render the correct copy

### Removed

- Remove the floating `<probo-settings-button>` widget and the `reopen-widget` attribute — place `<probo-settings-link>` in the header or footer instead (**breaking**)

## [0.11.0] - 2026-08-04

### Added

- Dutch (nl-NL) translations for cookie banner texts and labels

## [0.10.1] - 2026-07-28

### Fixed

- Grant all Google Consent Mode types when the banner config is unavailable (discovery mode), so GTM-managed tags can fire and be detected instead of staying blocked by the eager deny-all bootstrap

## [0.10.0] - 2026-06-19

### Changed

- Local storage, IndexedDB, and cache-storage trackers without an expiry now display a localized "persistent" duration rather than "session"

## [0.9.3] - 2026-06-12

### Fixed

- Hold the page scroll lock on sites using JS-driven smooth-scroll libraries (e.g. Lenis, as used by Framer) that bypass `overflow: hidden`. The body is now pinned with `position: fixed` (scroll position saved and restored) to remove the viewport's scroll distance entirely, with capture-phase wheel/touch cancellation as defense-in-depth for inner scroll containers; the preference panel's own list can still scroll

## [0.9.2] - 2026-06-11

### Changed

- References updated to probo.com

## [0.9.1] - 2026-06-10

### Fixed

- Lock page scroll while the preference panel is open and add `overscroll-behavior: contain` to the category list, so wheel events at panel edges no longer chain to the host page
- Exclude the SDK's own served bundle URL from tracker initiator attribution and from the resource detector, so third-party/extension writes are no longer misattributed to `cookie-banner.iife.js` when the SDK is served from a CDN distinct from the API host

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
