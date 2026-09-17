// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Keep in sync with packages/cookie-banner-tcf/src/constants.ts
const TCF_CMP_ID = 4095;
const TCF_CMP_VERSION = 1;

const DEFAULT_IIFE =
  "https://cdn.jsdelivr.net/npm/@probo/cookie-banner/dist/cookie-banner.iife.js";
const TCF_IIFE =
  "https://cdn.jsdelivr.net/npm/@probo/cookie-banner-tcf/dist/cookie-banner-tcf.iife.js";

export function cookieBannerEmbedSnippet(input: {
  bannerId: string;
  baseUrl: string;
  tcf: boolean;
}): string {
  const tag = `<script
  src="${input.tcf ? TCF_IIFE : DEFAULT_IIFE}"
  data-banner-id="${input.bannerId}"
  data-base-url="${input.baseUrl}"
  data-position="bottom-left"
></script>`;

  if (!input.tcf) {
    return tag;
  }

  return `${tcfStubSnippet()}\n${tag}`;
}

function tcfStubSnippet(): string {
  return `<script>
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
        cmpId: ${TCF_CMP_ID},
        cmpVersion: ${TCF_CMP_VERSION}
      }, true);
      return;
    }
    q.push(arguments);
  }
  stub.q = q;
  w.__tcfapi = stub;
})();
</script>`;
}
