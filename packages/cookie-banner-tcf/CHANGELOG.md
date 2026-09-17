# Changelog

All notable changes to the `@probo/cookie-banner-tcf` SDK will be documented
in this file.

## Unreleased

### Added

- Encodes a TCF 2.2 string with `@iabtechlabtcf/core`, installs a `__tcfapi`
  stub, and starts `@iabtechlabtcf/cmpapi`. Load `cookie-banner-tcf.iife.js`
  in place of the default banner IIFE when the hidden TCF capability is on.
  Accept/reject stores an optional `tc` field on `probo_consent`
