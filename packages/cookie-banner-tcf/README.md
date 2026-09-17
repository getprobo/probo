# @probo/cookie-banner-tcf

IAB Transparency and Consent Framework (TCF 2.3) addon for
[`@probo/cookie-banner`](../cookie-banner). It owns the `__tcfapi` stub,
`CmpApi`, TC-string encoder, and the TCF first and second layers.

`@probo/cookie-banner` stays IAB-free. Install both packages, or load this
package's IIFE, which bundles the host banner plus the IAB libraries.

This capability is hidden (SQL-flip only). When the flag is on, the console
snippet includes an inline `__tcfapi` stub plus this package's IIFE. When
`config.tcf` is present under GDPR, the addon replaces the category banner
with TCF disclosures and a purpose/vendor panel.

## Script tag (IIFE)

Place the stub as high in the page as possible so vendors can queue before
the SDK loads. Keep `cmpId` in sync with `TCF_CMP_ID`.

The stub must **run** before ad/vendor tags. Inline is the usual IAB shape
because the parser executes it immediately (no download). It does not have
to be inline: a tiny first **blocking** (no `async` / `defer`) script from
a CSP-allowed origin is fine. `async` / `defer` is not — that is the
timing hole.

If CSP blocks inline scripts, either:

- keep this stub and allow it with a `nonce` or sha256 hash, or
- host the same stub as a blocking first `<script src>` (the full TCF IIFE
  also calls `installTCFStub()` at the top, but only after that file
  downloads, so it cannot replace an early stub if vendors are already
  on the page).

```html
<script>
(function () {
  var w = window;
  if (typeof w.__tcfapi === "function") {
    return;
  }
  var q = [];
  function stub() {
    var command = arguments[0];
    var callback = arguments[2];
    if (command === "ping" && typeof callback === "function") {
      callback({
        gdprApplies: true,
        cmpLoaded: false,
        cmpStatus: "stub",
        displayStatus: "hidden",
        apiVersion: "2.2",
        cmpId: 4095,
        cmpVersion: 1
      }, true);
      return;
    }
    q.push(arguments);
  }
  stub.q = q;
  w.__tcfapi = stub;
})();
</script>
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
