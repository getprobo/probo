# @probo/cookie-banner-tcf

IAB Transparency and Consent Framework (TCF 2.3) addon for
[`@probo/cookie-banner`](../cookie-banner). It owns the `__tcfapi` stub,
`CmpApi`, TC-string encoder, and the TCF first and second layers.

`@probo/cookie-banner` stays IAB-free. Install both packages, or load this
package's IIFE, which bundles the host banner plus the IAB libraries.

This capability is hidden (SQL-flip only). Code snippets in the console still
point at the non-TCF IIFE. When `config.tcf` is present under GDPR, the addon
replaces the category banner with TCF disclosures and a purpose/vendor panel.

## Script tag (IIFE)

```html
<script
  src="https://cdn.jsdelivr.net/npm/@probo/cookie-banner-tcf/dist/cookie-banner-tcf.iife.js"
  data-banner-id="YOUR_BANNER_ID"
  data-base-url="https://your-probo-instance.com/api/cookie-banner/v1/"
  data-position="bottom-left"
></script>

<probo-settings-link></probo-settings-link>
```

## ES module

```bash
npm install @probo/cookie-banner @probo/cookie-banner-tcf
```

```js
import { bootThemedBanner } from "@probo/cookie-banner";
import { installTCFStub, startTCF } from "@probo/cookie-banner-tcf";

installTCFStub();
startTCF();
bootThemedBanner();
```
