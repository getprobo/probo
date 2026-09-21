# Changelog

All notable changes to the `@probo/cookie-banner-tcf` SDK will be documented
in this file.

## Unreleased

### Fixed

- Locator `__tcfapi` messages resolve the live page function so the IAB
  CMP validator sees `cmpStatus: "loaded"` after `CmpApi` starts. The
  stub also returns its command queue when called with no arguments,
  which is how `CmpApi` drains calls queued before load
- `CmpApi` ping reports the GVL version from GET config before consent
  instead of an empty TC model (`vendorListVersion` 0). Stored choices
  are re-encoded onto the current GVL; a TCF policy-version change
  still discards the cookie TC and re-prompts
- Preference toggles push a draft TC string to `CmpApi` while the panel
  is open, so withdrawing a vendor after accept-all updates `getTCData`
  before Save. Save still persists that draft
- First-layer store/access copy names the nature of personal data
  processed (unique identifiers and browsing data) per TCF Policy
  C(b)(II). Scope copy states that choices are service-specific
  per C(b)(VII). Withdrawal copy says consent can be changed at
  any time via Cookie settings per C(b)(VIII). When a vendor relies
  on legitimate interest, first-layer copy also states the right to
  object, and the second layer repeats the nature of the data and
  that choices are service-specific. Scope copy names choices, not
  only consent. Related sentences share paragraphs so translations
  stay on separate keys without a block per sentence
- Panel Accept all / Reject all persist the TC string on `probo_consent`
  even if `CmpApi.update` throws; those footer clicks are handled on
  the host so they cannot miss the headless buttons

### Changed

- `CmpApi` and TC encode require `config.tcf.cmp_id` and
  `config.tcf.cmp_version` from GET config. The `__tcfapi` stub omits
  those fields until the CMP loads. The SDK no longer hardcodes CMP ID
  or version
- `__tcfapi` `displayStatus` follows the banner UI: `visible` while the
  first layer, preferences, or privacy choices are open, `hidden` when
  they close, and `disabled` when TCF is inactive
- Encodes and reports IAB CMP ID 4095 instead of the placeholder 2

### Added

- Encodes a TCF 2.3 string with `@iabtechlabtcf/core`, installs a `__tcfapi`
  stub, and starts `@iabtechlabtcf/cmpapi`. Load `cookie-banner-tcf.iife.js`
  in place of the default banner IIFE when the hidden TCF capability is on.
  Accept/reject stores an optional `tc` field on `probo_consent`
- GDPR TCF first layer lists store/access, purposes, special features, and
  partner count; the second layer lets visitors toggle purposes, purpose
  legitimate interest, special features, and per-vendor consent/LI. Stacks
  group purposes when the GVL includes them. Disclosure copy uses
  `data-text` keys. Save encodes those bits; accept still grants all
  disclosed vendors and reject grants none
- Second layer discloses special purposes, features, and data categories
  used by partners (no toggles), purpose illustrations, vendor cookie /
  non-cookie storage, and http(s) privacy / legitimate-interest links.
  Each partner row lists that vendor's purposes by legal basis, special
  purposes, features, and special features. Cookie lines include max
  duration and whether it may be refreshed; purpose-specific retention
  and an http(s) device-storage disclosure link appear when the GVL
  provides them. Each purpose row shows how many partners seek consent
  or rely on legitimate interest for that purpose
- Preference panel is 720px for TCF. The header describes purposes and
  partners instead of cookie categories. Choice rows keep a leading spacer
  (no visible IAB ids) and labeled Consent / Legitimate interest /
  Opt-in toggles. Order is purposes
  (IAB stacks nested under that heading, leftovers under Other
  purposes), special features, then partners. Special purposes, features,
  and data categories sit in a collapsed More information block; special
  purposes keep an Always on chip
