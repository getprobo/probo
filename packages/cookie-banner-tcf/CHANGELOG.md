# Changelog

All notable changes to the `@probo/cookie-banner-tcf` SDK will be documented
in this file.

## Unreleased

## [0.2.0] - 2026-11-17

### Added

- The IIFE bundles `@probo/cookie-banner` 0.22.0 (0.1.0 bundled 0.19.0)
- While a sample is open, each page reports a page view, and cookies
  already listed on the banner are still reported
- A discovery hit includes `category_granted`, and the batch includes
  `consent_action` when the visitor has one. The grant follows current
  consent for the matching tracker. A `*` in the pattern is a glob: an
  exact name wins, and otherwise the match with the longest fixed text
  wins. Necessary categories are always granted, and an unmatched name
  uses the banner's default for non-necessary trackers
- Cookie detections report the cookie `Domain`, and a richer observation
  replaces an earlier one without demoting its source

### Changed

- `document.cookie` and named storage writes (for example
  `localStorage[key] = value`) are classified as `SCRIPT` instead of
  `HTTP` or `PRE_EXISTING`. Hooks start before the config fetch, and the
  leftover cookie scan waits for the config so inventoried cookies are
  not re-queued on every page load

### Fixed

- Storage writes report the size of the value actually stored, instead of
  failing on `null` or reporting `NaN` for numbers, and convert the key
  only once
- Extension cookies are no longer reported as site trackers
- Tracker reports drop a cookie `Max-Age` that does not fit a 32-bit
  integer instead of failing the whole batch
- Android Edge is counted as Edge, not Chrome, and a 404 on page-view
  recording no longer causes endless retries

## [0.1.0] - 2026-09-30

### Added

- IAB TCF 2.3 addon for the Probo cookie banner. Load
  `cookie-banner-tcf.iife.js` in place of the default banner IIFE when the
  hidden TCF capability is on. It encodes a TC string with
  `@iabtechlabtcf/core`, installs a `__tcfapi` stub, and starts
  `@iabtechlabtcf/cmpapi`. Accept and reject store an optional `tc` field
  on `probo_consent`. TCF markup applies to GDPR opt-in only; notice and
  opt-out keep their presentation
- GDPR first layer lists store/access, purposes, special features, and
  partner count, and names the nature of the personal data processed
  (unique identifiers and browsing data), that choices are
  service-specific, that consent can be changed at any time via Cookie
  settings, and the right to object where a vendor relies on legitimate
  interest, per TCF Policy C(b)(II), C(b)(VII) and C(b)(VIII)
- Second layer lets visitors toggle purposes, purpose legitimate
  interest, special features, and per-vendor consent and legitimate
  interest. Stacks group purposes when the GVL includes them; leftovers
  sit under Other purposes. It discloses purpose illustrations, vendor
  cookie and non-cookie storage, and privacy / legitimate-interest links.
  Each partner row lists that vendor's purposes by legal basis, special
  purposes (Always on), features, special features, and data categories.
  Cookie lines carry max duration and whether they may be refreshed;
  retention and a device-storage disclosure link appear when the GVL
  provides them. Each purpose row shows how many partners seek consent or
  rely on legitimate interest for it. Vendor consent toggles appear only
  when the vendor declares consent or flexible purposes
- Preference panel is 720px under TCF, headed by purposes and partners
  rather than cookie categories, with labeled Consent / Legitimate
  interest / Opt-in toggles and no visible IAB ids. Order is purposes,
  special features, partners, then storage
- `__tcfapi` `displayStatus` follows the banner UI: `visible` while the
  first layer, preferences, or privacy choices are open, `hidden` when
  they close, `disabled` when TCF is inactive. Ping reports the GVL
  version from GET config before consent, and stored choices are
  re-encoded onto the current GVL; a TCF policy-version change discards
  the cookie TC and re-prompts. Preference toggles push a draft TC string
  to `CmpApi` while the panel is open, which Save then persists
- CMP ID and version come from `config.tcf.cmp_id` and
  `config.tcf.cmp_version` on GET config rather than being hardcoded; the
  stub omits those fields until the CMP loads
