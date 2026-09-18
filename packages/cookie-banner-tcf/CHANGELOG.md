# Changelog

All notable changes to the `@probo/cookie-banner-tcf` SDK will be documented
in this file.

## Unreleased

### Changed

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
  non-cookie storage, and http(s) privacy / legitimate-interest links
